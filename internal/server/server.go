package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/air-gapped/cooked/internal/cache"
	"github.com/air-gapped/cooked/internal/config"
	"github.com/air-gapped/cooked/internal/fetch"
	"github.com/air-gapped/cooked/internal/logging"
	"github.com/air-gapped/cooked/internal/render"
	"github.com/air-gapped/cooked/internal/rewrite"
	"github.com/air-gapped/cooked/internal/sanitize"
	cookedtemplate "github.com/air-gapped/cooked/internal/template"
)

// Server is the main cooked HTTP server.
type Server struct {
	cfg        *config.Config
	version    string
	fetcher    *fetch.CachedClient
	mdRender   *render.MarkdownRenderer
	codeRender *render.CodeRenderer
	tmpl       *cookedtemplate.Renderer
	assets     fs.FS
	allowlist  *Allowlist
	mux        *http.ServeMux
}

// New creates a new cooked server with all dependencies.
// Extra fetch options can be passed for testing (e.g. disabling SSRF protection).
func New(cfg *config.Config, version string, assets fs.FS, extraFetchOpts ...fetch.Option) *Server {
	allowlist := ParseAllowlist(cfg.AllowedUpstreams)

	// Build fetch client options.
	var fetchOpts []fetch.Option
	fetchOpts = append(fetchOpts, extraFetchOpts...)

	// When an allowlist is configured, the operator has defined trusted
	// upstreams — that IS the security boundary. Disable blanket SSRF
	// dial-time blocking so private IPs (10.x, 172.16.x, etc.) work.
	if allowlist != nil {
		fetchOpts = append(fetchOpts, fetch.WithSSRFProtection(false))
	}

	// Add redirect validator that enforces the allowlist on redirect targets.
	if allowlist != nil {
		fetchOpts = append(fetchOpts, fetch.WithRedirectValidator(func(target *url.URL) error {
			if !allowlist.Allows(target.Host) {
				return fmt.Errorf("redirect target %q not in allowed upstreams", target.Host)
			}
			return nil
		}))
	}

	client := fetch.NewClient(cfg.FetchTimeout, cfg.MaxFileSize, cfg.TLSSkipVerify, fetchOpts...)
	memCache := cache.New(cfg.CacheTTL, cfg.CacheMaxSize)
	cachedClient := fetch.NewCachedClient(client, memCache)

	s := &Server{
		cfg:        cfg,
		version:    version,
		fetcher:    cachedClient,
		mdRender:   render.NewMarkdownRenderer(),
		codeRender: render.NewCodeRenderer(),
		tmpl:       cookedtemplate.NewRenderer(),
		assets:     assets,
		allowlist:  allowlist,
		mux:        http.NewServeMux(),
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /_cooked/{path...}", s.handleAsset)
	s.mux.HandleFunc("GET /{$}", s.handleLanding)
	s.mux.HandleFunc("GET /{upstream...}", s.handleRender)
}

// Handler returns the server's HTTP handler with middleware applied.
func (s *Server) Handler() http.Handler {
	var h http.Handler = s.mux
	h = s.loggingMiddleware(h)
	return h
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.tmpl.RenderLanding(s.version, s.cfg.DefaultTheme))
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := r.PathValue("path")
	data, err := fs.ReadFile(s.assets, assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set content type based on extension
	switch {
	case len(assetPath) > 3 && assetPath[len(assetPath)-3:] == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case len(assetPath) > 4 && assetPath[len(assetPath)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css")
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Extract upstream URL from path
	rawUpstream := ExtractUpstreamFromPath(r.URL.Path, r.URL.RawQuery)

	// Parse and validate
	upstream, err := ParseUpstreamURL(rawUpstream)
	if err != nil {
		s.renderError(w, rawUpstream, 400, "bad-request", fmt.Sprintf("Invalid URL: %v", err))
		return
	}

	// Check allowed upstreams (nil allowlist = allow all, falls through to SSRF)
	if !s.allowlist.Allows(upstream.Host) {
		s.renderError(w, rawUpstream, 403, "blocked", "This upstream is not in the allowed list")
		return
	}

	// SSRF protection: when no allowlist is set, block private/loopback IPs.
	// When an allowlist IS set, the operator has defined the trust boundary
	// and SSRF dial-time blocking is disabled in the fetch client.
	if s.allowlist == nil {
		private, err := IsPrivateAddress(upstream.Host)
		if err != nil {
			s.renderError(w, rawUpstream, 502, "unreachable", fmt.Sprintf("Could not resolve host: %v", err))
			return
		}
		if private {
			s.renderError(w, rawUpstream, 403, "blocked", "Fetching from private/loopback addresses is not allowed")
			return
		}
	}

	// Fetch from upstream (with caching)
	result, cachedEntry, err := s.fetcher.Fetch(rawUpstream)
	if err != nil {
		if isTimeout(err) {
			s.renderError(w, rawUpstream, 504, "timeout",
				fmt.Sprintf("Upstream request timed out after %s", s.cfg.FetchTimeout))
			return
		}
		if isTooLarge(err) {
			s.renderError(w, rawUpstream, 413, "too-large",
				fmt.Sprintf("File too large (limit is %d bytes)", s.cfg.MaxFileSize))
			return
		}
		s.renderError(w, rawUpstream, 502, "unreachable",
			fmt.Sprintf("Could not reach upstream server: %v", err))
		return
	}

	// If cache hit, serve from cache
	if cachedEntry != nil && (result.CacheStatus == cache.StatusHit || result.CacheStatus == cache.StatusRevalidated || result.CacheStatus == cache.StatusStale) {
		s.serveFromCache(w, rawUpstream, cachedEntry, result, start)
		return
	}

	// Non-200 upstream
	if result.StatusCode != 200 {
		s.renderError(w, rawUpstream, result.StatusCode, "upstream-error",
			fmt.Sprintf("Upstream returned %d", result.StatusCode))
		return
	}

	// Detect file type
	fileInfo := render.DetectFile(upstream.Path)

	// Render based on content type
	var htmlContent []byte
	var meta *render.MarkdownMeta
	renderStart := time.Now()

	switch fileInfo.ContentType {
	case render.TypeMarkdown:
		htmlContent, meta, err = s.mdRender.Render(result.Body)
		if err != nil {
			slog.Error("render markdown failed", "error", err, "upstream", rawUpstream)
			s.renderError(w, rawUpstream, 500, "render-error", "Failed to render markdown")
			return
		}

	case render.TypeMDX:
		preprocessed := render.PreprocessMDX(result.Body)
		htmlContent, meta, err = s.mdRender.Render(preprocessed)
		if err != nil {
			slog.Error("render mdx failed", "error", err, "upstream", rawUpstream)
			s.renderError(w, rawUpstream, 500, "render-error", "Failed to render MDX")
			return
		}

	case render.TypeCode:
		htmlContent, err = s.codeRender.Render(result.Body, fileInfo.Language)
		if err != nil {
			slog.Error("render code failed", "error", err, "upstream", rawUpstream)
			s.renderError(w, rawUpstream, 500, "render-error", "Failed to render code")
			return
		}

	case render.TypePlaintext:
		htmlContent = render.RenderPlaintext(result.Body)

	default:
		s.renderError(w, rawUpstream, 415, "unsupported",
			"This file type is not supported for rendering")
		return
	}

	renderMs := time.Since(renderStart).Milliseconds()

	// Sanitize HTML (only for markdown/mdx which may contain upstream HTML)
	if fileInfo.ContentType == render.TypeMarkdown || fileInfo.ContentType == render.TypeMDX {
		htmlContent = sanitize.HTML(htmlContent)
	}

	// Rewrite relative URLs
	if fileInfo.ContentType == render.TypeMarkdown || fileInfo.ContentType == render.TypeMDX {
		htmlContent = rewrite.RelativeURLs(htmlContent, rawUpstream, s.cfg.BaseURL)
	}

	// Load embedded CSS
	lightCSS := readAssetString(s.assets, "github-markdown-light.css")
	darkCSS := readAssetString(s.assets, "github-markdown-dark.css")

	// Build page data
	pageData := cookedtemplate.PageData{
		Version:        s.version,
		UpstreamURL:    rawUpstream,
		ContentType:    fileInfo.ContentType,
		CacheStatus:    string(result.CacheStatus),
		UpstreamStatus: result.StatusCode,
		FileSize:       result.ContentLen,
		LastModified:   result.LastModified,
		DefaultTheme:   s.cfg.DefaultTheme,
		Content:        template.HTML(htmlContent),
		MermaidPath:    "/_cooked/mermaid.min.js",
	}

	if meta != nil {
		pageData.Title = meta.Title
		pageData.HasMermaid = meta.HasMermaid
		pageData.HeadingCount = meta.HeadingCount
		pageData.CodeBlockCount = meta.CodeBlockCount
		pageData.Headings = meta.Headings
	}

	// Render full page
	page := s.tmpl.RenderPage(pageData, lightCSS, darkCSS)

	// Store in cache
	s.fetcher.Store(rawUpstream, cache.Entry{
		HTML:         page,
		ETag:         result.ETag,
		LastModified: result.LastModified,
		Size:         int64(len(page)),
		ContentType:  string(fileInfo.ContentType),
	})

	// Set response headers
	s.setResponseHeaders(w, rawUpstream, result.StatusCode, string(result.CacheStatus),
		string(fileInfo.ContentType), renderMs, result.FetchMs, s.version)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(200)
	w.Write(page)
}

func (s *Server) serveFromCache(w http.ResponseWriter, rawUpstream string, entry *cache.Entry, result *fetch.CachedResult, start time.Time) {
	s.setResponseHeaders(w, rawUpstream, 200, string(result.CacheStatus),
		entry.ContentType, 0, result.FetchMs, s.version)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(200)
	w.Write(entry.HTML)
}

func (s *Server) renderError(w http.ResponseWriter, upstreamURL string, statusCode int, errType, message string) {
	page := s.tmpl.RenderError(cookedtemplate.ErrorData{
		Version:      s.version,
		UpstreamURL:  upstreamURL,
		StatusCode:   statusCode,
		ErrorType:    errType,
		Message:      message,
		DefaultTheme: s.cfg.DefaultTheme,
	})

	s.setResponseHeaders(w, upstreamURL, statusCode, "", "error", 0, 0, s.version)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(page)
}

func (s *Server) setResponseHeaders(w http.ResponseWriter, upstream string, upstreamStatus int,
	cacheStatus, contentType string, renderMs, upstreamMs int64, version string) {

	// F-10: Security response headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")

	w.Header().Set("X-Cooked-Version", version)
	w.Header().Set("X-Cooked-Upstream", redactUpstream(upstream))
	w.Header().Set("X-Cooked-Upstream-Status", fmt.Sprintf("%d", upstreamStatus))
	w.Header().Set("X-Cooked-Cache", cacheStatus)
	w.Header().Set("X-Cooked-Content-Type", contentType)
	w.Header().Set("X-Cooked-Render-Ms", fmt.Sprintf("%d", renderMs))
	w.Header().Set("X-Cooked-Upstream-Ms", fmt.Sprintf("%d", upstreamMs))
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &logging.ByteCountingWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)

		if wrapped.StatusCode == 0 {
			wrapped.StatusCode = 200
		}

		logging.LogRequest(slog.Default(), logging.RequestFields{
			Method:      r.Method,
			Path:        r.URL.Path,
			Upstream:    wrapped.Header().Get("X-Cooked-Upstream"),
			Status:      wrapped.StatusCode,
			Cache:       wrapped.Header().Get("X-Cooked-Cache"),
			UpstreamMs:  parseHeaderInt64(wrapped.Header().Get("X-Cooked-Upstream-Ms")),
			RenderMs:    parseHeaderInt64(wrapped.Header().Get("X-Cooked-Render-Ms")),
			TotalMs:     time.Since(start).Milliseconds(),
			ContentType: wrapped.Header().Get("X-Cooked-Content-Type"),
			Bytes:       wrapped.Bytes,
		})
	})
}

// redactUpstream strips query, fragment, and userinfo from an upstream URL
// to avoid leaking tokens or credentials in headers and logs.
func redactUpstream(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func parseHeaderInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func readAssetString(fsys fs.FS, name string) string {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return ""
	}
	return string(data)
}

func isTimeout(err error) bool {
	return err != nil && (contains(err.Error(), "timeout") || contains(err.Error(), "deadline"))
}

func isTooLarge(err error) bool {
	return err != nil && contains(err.Error(), "too large")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
