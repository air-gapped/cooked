package rewrite

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/air-gapped/cooked/internal/render"
)

// hrefRe matches href="..." and src="..." attributes in HTML.
var hrefRe = regexp.MustCompile(`(?i)((?:href|src)\s*=\s*")([^"]+)(")`)

// RelativeURLs rewrites relative URLs in HTML content.
// Markdown file links are rewritten through cooked (e.g. /https://upstream/path/CONTRIBUTING.md).
// Non-markdown file links (images, etc.) are rewritten to point directly at upstream.
// Absolute URLs are left untouched.
func RelativeURLs(html []byte, upstreamURL, baseURL string) []byte {
	// Parse upstream URL to determine base path
	u, err := url.Parse(upstreamURL)
	if err != nil {
		return html
	}

	// Base is the upstream URL with the filename stripped
	basePath := path.Dir(u.Path)
	if basePath == "." {
		basePath = ""
	}
	base := u.Scheme + "://" + u.Host + basePath

	return hrefRe.ReplaceAllFunc(html, func(match []byte) []byte {
		parts := hrefRe.FindSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		prefix := string(parts[1]) // href="
		href := string(parts[2])   // the URL
		suffix := string(parts[3]) // "

		// Skip absolute URLs
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "//") {
			return match
		}

		// Skip fragments and data URIs
		if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "data:") || strings.HasPrefix(href, "mailto:") {
			return match
		}

		// Separate fragment from path
		fragment := ""
		if idx := strings.Index(href, "#"); idx >= 0 {
			fragment = href[idx:]
			href = href[:idx]
		}

		// Separate query from path
		query := ""
		if idx := strings.Index(href, "?"); idx >= 0 {
			query = href[idx:]
			href = href[:idx]
		}

		if href == "" {
			// Fragment-only or query-only link
			return match
		}

		// Resolve relative path against base
		resolved := resolveRelative(base, href)

		if render.IsMarkdownLink(href) {
			// Markdown links go through cooked
			cookedPrefix := "/"
			if baseURL != "" {
				cookedPrefix = strings.TrimRight(baseURL, "/") + "/"
			}
			return []byte(prefix + cookedPrefix + resolved + query + fragment + suffix)
		}

		// Non-markdown links point directly to upstream
		return []byte(prefix + resolved + query + fragment + suffix)
	})
}

// resolveRelative resolves a relative href against a base URL.
func resolveRelative(base, href string) string {
	// Handle ./ prefix
	href = strings.TrimPrefix(href, "./")

	// If relative, join with base
	if !strings.Contains(href, "://") {
		return strings.TrimRight(base, "/") + "/" + href
	}
	return href
}
