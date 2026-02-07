package fetch

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
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

// Client fetches content from upstream URLs.
type Client struct {
	httpClient  *http.Client
	maxFileSize int64
}

// NewClient creates a fetch client with the given configuration.
func NewClient(timeout time.Duration, maxFileSize int64, tlsSkipVerify bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
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
