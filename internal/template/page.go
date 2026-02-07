package template

import (
	"bytes"
	"fmt"
	"html"
	htmltemplate "html/template"
	"strings"
	"time"

	"github.com/air-gapped/cooked/internal/render"
)

// PageData holds all the data needed to render a full HTML page.
type PageData struct {
	Version        string
	UpstreamURL    string
	ContentType    render.ContentType
	CacheStatus    string
	UpstreamStatus int
	FileSize       int64
	LastModified   string // ISO timestamp
	DefaultTheme   string
	Title          string
	Content        htmltemplate.HTML
	HasMermaid     bool
	HasTOC         bool
	HeadingCount   int
	CodeBlockCount int
	Headings       []render.Heading
	MermaidPath    string // path to embedded mermaid.js
}

// ErrorData holds data for error pages.
type ErrorData struct {
	Version      string
	UpstreamURL  string
	StatusCode   int
	ErrorType    string
	Message      string
	DefaultTheme string
}

// Renderer renders full HTML pages.
type Renderer struct {
	chromaLightCSS string
	chromaDarkCSS  string
}

// NewRenderer creates a template renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		chromaLightCSS: chromaLight,
		chromaDarkCSS:  chromaDark,
	}
}

// RenderPage produces a complete HTML page.
func (r *Renderer) RenderPage(data PageData, lightCSS, darkCSS string) []byte {
	var buf bytes.Buffer

	title := data.Title
	if title == "" {
		// Extract filename from URL
		parts := strings.Split(data.UpstreamURL, "/")
		if len(parts) > 0 {
			title = parts[len(parts)-1]
		}
	}
	if title == "" {
		title = "cooked"
	}

	escapedURL := html.EscapeString(data.UpstreamURL)
	truncatedURL := truncateURL(data.UpstreamURL, 80)

	// HTML head
	fmt.Fprintf(&buf, `<!DOCTYPE html>
<html lang="en"
      data-theme="%s"
      data-cooked-version="%s"
      data-upstream-url="%s"
      data-content-type="%s"
      data-cache-status="%s">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s — cooked</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%s">
  <style>
`,
		html.EscapeString(data.DefaultTheme),
		html.EscapeString(data.Version),
		escapedURL,
		html.EscapeString(string(data.ContentType)),
		html.EscapeString(data.CacheStatus),
		html.EscapeString(title),
		faviconSVG,
	)

	// Embedded CSS
	writeThemeCSS(&buf, lightCSS, darkCSS, r.chromaLightCSS, r.chromaDarkCSS)
	writeLayoutCSS(&buf)

	fmt.Fprintf(&buf, `
  </style>
</head>
<body>
`)

	// Header
	fmt.Fprintf(&buf, "  <!-- cooked: header -->\n")
	writeHeader(&buf, data, escapedURL, truncatedURL)

	// TOC
	hasTOC := len(data.Headings) >= 3
	fmt.Fprintf(&buf, "  <!-- cooked: table of contents -->\n")
	if hasTOC {
		writeTOC(&buf, data.Headings)
	}

	// Content
	fmt.Fprintf(&buf, "  <!-- cooked: content -->\n")
	fmt.Fprintf(&buf, `  <main>
    <article id="cooked-content"
             class="markdown-body"
             data-has-mermaid="%v"
             data-has-toc="%v"
             data-heading-count="%d"
             data-code-block-count="%d">
      %s
    </article>
  </main>
`,
		data.HasMermaid,
		hasTOC,
		data.HeadingCount,
		data.CodeBlockCount,
		data.Content,
	)

	// Scripts
	fmt.Fprintf(&buf, "  <!-- cooked: scripts -->\n")
	writeScripts(&buf)

	// Mermaid
	if data.HasMermaid && data.MermaidPath != "" {
		fmt.Fprintf(&buf, "  <script src=\"%s\"></script>\n", html.EscapeString(data.MermaidPath))
		fmt.Fprintf(&buf, "  <script>mermaid.initialize({startOnLoad: true, theme: 'default'});</script>\n")
	}

	fmt.Fprintf(&buf, "</body>\n</html>\n")

	return buf.Bytes()
}

// RenderError produces an error page.
func (r *Renderer) RenderError(data ErrorData) []byte {
	var buf bytes.Buffer

	escapedURL := html.EscapeString(data.UpstreamURL)

	fmt.Fprintf(&buf, `<!DOCTYPE html>
<html lang="en"
      data-theme="%s"
      data-cooked-version="%s"
      data-upstream-url="%s"
      data-content-type="error"
      data-error-type="%s">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Error — cooked</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%s">
  <style>
`,
		html.EscapeString(data.DefaultTheme),
		html.EscapeString(data.Version),
		escapedURL,
		html.EscapeString(data.ErrorType),
		faviconSVG,
	)

	writeLayoutCSS(&buf)

	fmt.Fprintf(&buf, `
  </style>
</head>
<body>
  <!-- cooked: header -->
  <header id="cooked-header">
    <div class="cooked-meta">
      <a id="cooked-source-link" href="%s" title="%s">%s</a>
    </div>
    <div class="cooked-controls">
      <button id="cooked-theme-toggle" title="Toggle theme">&#x25D1;</button>
    </div>
  </header>
  <!-- cooked: content -->
  <main>
    <div id="cooked-error"
         data-status-code="%d"
         data-error-message="%s">
      <h1>%d %s</h1>
      <p>%s</p>
      <p><a href="%s">View original file</a></p>
    </div>
  </main>
  <!-- cooked: scripts -->
`,
		escapedURL, escapedURL, html.EscapeString(truncateURL(data.UpstreamURL, 80)),
		data.StatusCode, html.EscapeString(data.Message),
		data.StatusCode, html.EscapeString(http_status_text(data.StatusCode)),
		html.EscapeString(data.Message),
		escapedURL,
	)

	writeScripts(&buf)

	fmt.Fprintf(&buf, "</body>\n</html>\n")

	return buf.Bytes()
}

func http_status_text(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 408:
		return "Request Timeout"
	case 413:
		return "Payload Too Large"
	case 502:
		return "Bad Gateway"
	case 504:
		return "Gateway Timeout"
	default:
		return "Error"
	}
}

func writeHeader(buf *bytes.Buffer, data PageData, escapedURL, truncatedURL string) {
	lastModifiedAttr := ""
	if data.LastModified != "" {
		lastModifiedAttr = fmt.Sprintf(`data-last-modified="%s"`, html.EscapeString(data.LastModified))
	}

	fmt.Fprintf(buf, `  <header id="cooked-header"
          data-upstream-status="%d"
          data-file-size="%d"
          %s>
    <div class="cooked-meta">
      <a id="cooked-source-link" href="%s" title="%s">%s</a>
`,
		data.UpstreamStatus,
		data.FileSize,
		lastModifiedAttr,
		escapedURL, escapedURL,
		html.EscapeString(truncatedURL),
	)

	if data.LastModified != "" {
		relative := formatRelativeTime(data.LastModified)
		fmt.Fprintf(buf, `      <time id="cooked-modified" datetime="%s" title="%s">Modified %s</time>
`,
			html.EscapeString(data.LastModified),
			html.EscapeString(data.LastModified),
			html.EscapeString(relative),
		)
	}

	fmt.Fprintf(buf, `      <span id="cooked-size">%s</span>
      <span id="cooked-type">%s</span>
    </div>
    <div class="cooked-controls">
`,
		html.EscapeString(formatFileSize(data.FileSize)),
		html.EscapeString(contentTypeLabel(data.ContentType)),
	)

	if len(data.Headings) >= 3 {
		fmt.Fprintf(buf, "      <button id=\"cooked-toc-toggle\" title=\"Table of contents\">&#9776;</button>\n")
	}

	fmt.Fprintf(buf, "      <button id=\"cooked-theme-toggle\" title=\"Toggle theme\">&#x25D1;</button>\n")
	fmt.Fprintf(buf, "    </div>\n  </header>\n")
}

func writeTOC(buf *bytes.Buffer, headings []render.Heading) {
	fmt.Fprintf(buf, "  <nav id=\"cooked-toc\" hidden>\n    <ul>\n")
	for _, h := range headings {
		fmt.Fprintf(buf, `      <li data-level="%d"><a href="#%s">%s</a></li>
`,
			h.Level,
			html.EscapeString(h.ID),
			html.EscapeString(h.Text),
		)
	}
	fmt.Fprintf(buf, "    </ul>\n  </nav>\n")
}

func truncateURL(u string, maxLen int) string {
	if len(u) <= maxLen {
		return u
	}
	return u[:maxLen-3] + "..."
}

func formatFileSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatRelativeTime(isoTime string) string {
	t, err := time.Parse(time.RFC1123, isoTime)
	if err != nil {
		t, err = time.Parse(time.RFC3339, isoTime)
	}
	if err != nil {
		return isoTime
	}

	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}
}

func contentTypeLabel(ct render.ContentType) string {
	switch ct {
	case render.TypeMarkdown:
		return "Markdown"
	case render.TypeMDX:
		return "MDX"
	case render.TypeCode:
		return "Code"
	case render.TypePlaintext:
		return "Plain Text"
	default:
		return "Unknown"
	}
}
