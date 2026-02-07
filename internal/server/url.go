package server

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/air-gapped/cooked/internal/ssrf"
)

// ParseUpstreamURL extracts and validates the upstream URL from the request path.
// The path should be the full URL (scheme included) after stripping the leading "/".
func ParseUpstreamURL(rawPath string) (*url.URL, error) {
	if rawPath == "" {
		return nil, fmt.Errorf("empty upstream URL")
	}

	u, err := url.Parse(rawPath)
	if err != nil {
		return nil, fmt.Errorf("parse upstream url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q: only http and https are allowed", u.Scheme)
	}

	if u.Host == "" {
		return nil, fmt.Errorf("missing host in upstream URL")
	}

	return u, nil
}

// CheckAllowedUpstream verifies that the upstream host matches one of the
// allowed upstreams using exact or subdomain matching.
// If allowedUpstreams is empty, all hosts are allowed.
func CheckAllowedUpstream(host string, allowedUpstreams string) bool {
	if allowedUpstreams == "" {
		return true
	}

	// Strip port from host for matching
	hostname := strings.ToLower(host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = strings.ToLower(h)
	}

	entries := strings.Split(allowedUpstreams, ",")
	for _, entry := range entries {
		allowed := strings.ToLower(strings.TrimSpace(entry))
		if allowed == "" {
			continue
		}
		// Exact match or subdomain match (hostname ends with "."+allowed)
		if hostname == allowed || strings.HasSuffix(hostname, "."+allowed) {
			return true
		}
	}
	return false
}

// IsPrivateAddress returns true if the given hostname resolves to a private/loopback address.
// Used as a fast-fail SSRF check; real enforcement is at dial time via the custom DialContext.
func IsPrivateAddress(hostname string) (bool, error) {
	// Strip port if present
	host := hostname
	if h, _, err := net.SplitHostPort(hostname); err == nil {
		host = h
	}

	// Try parsing as IP directly
	if ip := net.ParseIP(host); ip != nil {
		return ssrf.IsBlockedIP(ip), nil
	}

	// Resolve hostname
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, fmt.Errorf("resolve host %q: %w", host, err)
	}

	for _, ip := range ips {
		if ssrf.IsBlockedIP(ip) {
			return true, nil
		}
	}
	return false, nil
}

// ExtractUpstreamFromPath takes the request path (with leading /) and query string,
// returns the full upstream URL string.
// Go's ServeMux normalizes // to / via 301 redirect, so we also handle
// paths like /http:/host/path and /https:/host/path by restoring the double slash.
func ExtractUpstreamFromPath(path, rawQuery string) string {
	// Strip leading /
	upstream := strings.TrimPrefix(path, "/")

	// Restore double-slash after scheme if ServeMux cleaned it
	if strings.HasPrefix(upstream, "http:/") && !strings.HasPrefix(upstream, "http://") {
		upstream = "http://" + upstream[len("http:/"):]
	}
	if strings.HasPrefix(upstream, "https:/") && !strings.HasPrefix(upstream, "https://") {
		upstream = "https://" + upstream[len("https:/"):]
	}

	if rawQuery != "" {
		upstream += "?" + rawQuery
	}
	return upstream
}
