package fetch

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/air-gapped/cooked/internal/ssrf"
)

// Result holds the outcome of an upstream fetch.
type Result struct {
	Body         []byte
	StatusCode   int
	ContentType  string
	ETag         string
	LastModified string
	ContentLen   int64
	FetchMs      int64
}

// RedirectValidator is called on each redirect hop. It receives the target URL
// and should return a non-nil error to block the redirect.
type RedirectValidator func(target *url.URL) error

// Option configures the fetch Client.
type Option func(*options)

type options struct {
	redirectValidator RedirectValidator
	ssrfProtection    bool
}

// WithRedirectValidator sets a callback to validate each redirect hop.
func WithRedirectValidator(v RedirectValidator) Option {
	return func(o *options) { o.redirectValidator = v }
}

// WithSSRFProtection enables or disables the SSRF dial-time IP check.
// Enabled by default.
func WithSSRFProtection(enabled bool) Option {
	return func(o *options) { o.ssrfProtection = enabled }
}

// Client fetches content from upstream URLs.
type Client struct {
	httpClient  *http.Client
	maxFileSize int64
}

// maxRedirects is the maximum number of HTTP redirects to follow.
const maxRedirects = 5

// NewClient creates a fetch client with the given configuration.
func NewClient(timeout time.Duration, maxFileSize int64, tlsSkipVerify bool, opts ...Option) *Client {
	cfg := options{ssrfProtection: true}
	for _, o := range opts {
		o(&cfg)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if cfg.ssrfProtection {
		// F-03: Custom DialContext for DNS TOCTOU protection.
		// Resolves DNS, checks all IPs against ssrf.IsBlockedIP, then dials only if safe.
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("ssrf dial: split host port %q: %w", addr, err)
			}

			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("ssrf dial: resolve %q: %w", host, err)
			}

			for _, ipAddr := range ips {
				if ssrf.IsBlockedIP(ipAddr.IP) {
					return nil, fmt.Errorf("ssrf dial: blocked IP %s for host %q", ipAddr.IP, host)
				}
			}

			dialer := &net.Dialer{}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		}
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// F-01: CheckRedirect to validate each redirect hop.
	validator := cfg.redirectValidator
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("too many redirects (max %d)", maxRedirects)
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("redirect to unsupported scheme %q", req.URL.Scheme)
		}
		if validator != nil {
			if err := validator(req.URL); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
		}
		return nil
	}

	return &Client{
		httpClient:  client,
		maxFileSize: maxFileSize,
	}
}

// Fetch retrieves content from the given URL. It does not forward any credentials
// from the original browser request. If ifNoneMatch or ifModifiedSince are set,
// a conditional GET is performed.
func (c *Client) Fetch(rawURL, ifNoneMatch, ifModifiedSince string) (*Result, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Conditional GET headers
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream fetch: %w", err)
	}
	defer resp.Body.Close()

	fetchMs := time.Since(start).Milliseconds()

	// 304 Not Modified — no body to read
	if resp.StatusCode == http.StatusNotModified {
		return &Result{
			StatusCode:   304,
			ETag:         resp.Header.Get("ETag"),
			LastModified: resp.Header.Get("Last-Modified"),
			FetchMs:      fetchMs,
		}, nil
	}

	// Check Content-Length before reading body
	if resp.ContentLength > 0 && resp.ContentLength > c.maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (limit %d)", resp.ContentLength, c.maxFileSize)
	}

	// Read body with size limit
	limited := io.LimitReader(resp.Body, c.maxFileSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if int64(len(body)) > c.maxFileSize {
		return nil, fmt.Errorf("file too large: exceeds %d bytes limit", c.maxFileSize)
	}

	return &Result{
		Body:         body,
		StatusCode:   resp.StatusCode,
		ContentType:  resp.Header.Get("Content-Type"),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		ContentLen:   int64(len(body)),
		FetchMs:      fetchMs,
	}, nil
}
