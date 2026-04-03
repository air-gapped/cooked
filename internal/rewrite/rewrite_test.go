package rewrite

import (
	"strings"
	"testing"
)

func TestRelativeURLs_MarkdownLinks(t *testing.T) {
	html := []byte(`<a href="CONTRIBUTING.md">Contributing</a>`)
	got := string(RelativeURLs(html, "https://cgit.internal/repo/plain/README.md", "", ""))

	want := `/https://cgit.internal/repo/plain/CONTRIBUTING.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant href containing %s", got, want)
	}
}

func TestRelativeURLs_MarkdownSubdir(t *testing.T) {
	html := []byte(`<a href="docs/guide.md">Guide</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "", ""))

	want := `/https://example.com/repo/docs/guide.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestRelativeURLs_ImagesSrcProxied(t *testing.T) {
	html := []byte(`<img src="./docs/arch.png" alt="architecture">`)
	got := string(RelativeURLs(html, "https://cgit.internal/repo/plain/README.md", "", "/_cooked/raw/"))

	want := `/_cooked/raw/https://cgit.internal/repo/plain/docs/arch.png`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant src containing %s", got, want)
	}
}

func TestRelativeURLs_ImagesSrcDirectWhenNoPrefix(t *testing.T) {
	html := []byte(`<img src="./docs/arch.png" alt="architecture">`)
	got := string(RelativeURLs(html, "https://cgit.internal/repo/plain/README.md", "", ""))

	// Without rawProxyPrefix, images point directly to upstream
	want := `https://cgit.internal/repo/plain/docs/arch.png`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant src containing %s", got, want)
	}
}

func TestRelativeURLs_HrefNonRenderableNotProxied(t *testing.T) {
	html := []byte(`<a href="archive.zip">Download</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "", "/_cooked/raw/"))

	// href on non-renderable files should NOT go through raw proxy
	if strings.Contains(got, "/_cooked/raw/") {
		t.Errorf("href was proxied through raw, should point at upstream: %s", got)
	}
	want := `https://example.com/repo/archive.zip`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant href containing %s", got, want)
	}
}

func TestRelativeURLs_AbsoluteURLsUntouched(t *testing.T) {
	html := []byte(`<a href="https://example.com/other">Other</a>`)
	got := string(RelativeURLs(html, "https://upstream.com/file.md", "", "/_cooked/raw/"))

	if !strings.Contains(got, `href="https://example.com/other"`) {
		t.Errorf("absolute URL was modified: %s", got)
	}
}

func TestRelativeURLs_AbsoluteSrcUntouched(t *testing.T) {
	html := []byte(`<img src="https://img.shields.io/badge/build-passing.svg">`)
	got := string(RelativeURLs(html, "https://upstream.com/file.md", "", "/_cooked/raw/"))

	if !strings.Contains(got, `src="https://img.shields.io/badge/build-passing.svg"`) {
		t.Errorf("absolute src URL was modified: %s", got)
	}
}

func TestRelativeURLs_FragmentOnly(t *testing.T) {
	html := []byte(`<a href="#section">Section</a>`)
	got := string(RelativeURLs(html, "https://example.com/file.md", "", ""))

	if !strings.Contains(got, `href="#section"`) {
		t.Errorf("fragment-only link was modified: %s", got)
	}
}

func TestRelativeURLs_MarkdownWithFragment(t *testing.T) {
	html := []byte(`<a href="other.md#section">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "", ""))

	if !strings.Contains(got, "#section") {
		t.Errorf("fragment was lost: %s", got)
	}
	if !strings.Contains(got, "other.md") {
		t.Errorf("markdown link not rewritten: %s", got)
	}
}

func TestRelativeURLs_WithBaseURL(t *testing.T) {
	html := []byte(`<a href="other.md">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "https://cooked.example.com", ""))

	want := `https://cooked.example.com/https://example.com/repo/other.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestRelativeURLs_WithBaseURLImageProxy(t *testing.T) {
	html := []byte(`<img src="logo.png">`)
	got := string(RelativeURLs(html, "https://upstream.com/repo/README.md", "https://cooked.example.com", "https://cooked.example.com/_cooked/raw/"))

	want := `https://cooked.example.com/_cooked/raw/https://upstream.com/repo/logo.png`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestRelativeURLs_DataURIUntouched(t *testing.T) {
	html := []byte(`<img src="data:image/png;base64,abc">`)
	got := string(RelativeURLs(html, "https://example.com/file.md", "", "/_cooked/raw/"))

	if !strings.Contains(got, "data:image/png") {
		t.Errorf("data URI was modified: %s", got)
	}
}

func TestRelativeURLs_ProtocolRelativeUntouched(t *testing.T) {
	html := []byte(`<a href="//cdn.example.com/lib.js">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/file.md", "", ""))

	if !strings.Contains(got, `//cdn.example.com/lib.js`) {
		t.Errorf("protocol-relative URL was modified: %s", got)
	}
}

func FuzzRewriteRelativeURLs(f *testing.F) {
	type seed struct {
		html        string
		upstreamURL string
		baseURL     string
	}
	seeds := []seed{
		// Standard href/src patterns
		{`<a href="CONTRIBUTING.md">link</a>`, "https://example.com/repo/README.md", ""},
		{`<img src="./docs/arch.png" alt="img">`, "https://example.com/repo/README.md", ""},
		{`<a href="docs/guide.md">Guide</a>`, "https://example.com/repo/README.md", ""},
		// Fragment
		{`<a href="#section">Sec</a>`, "https://example.com/file.md", ""},
		{`<a href="other.md#heading">Link</a>`, "https://example.com/repo/README.md", ""},
		// Query strings
		{`<a href="other.md?ref=main">Link</a>`, "https://example.com/repo/README.md", ""},
		{`<a href="file.md?a=1&b=2#frag">Link</a>`, "https://example.com/repo/README.md", ""},
		// Protocol-relative URLs
		{`<a href="//cdn.example.com/lib.js">Link</a>`, "https://example.com/file.md", ""},
		// Data URIs
		{`<img src="data:image/png;base64,abc">`, "https://example.com/file.md", ""},
		// Mailto
		{`<a href="mailto:user@example.com">Email</a>`, "https://example.com/file.md", ""},
		// Absolute URLs (should be untouched)
		{`<a href="https://other.com/page">Link</a>`, "https://example.com/file.md", ""},
		{`<a href="http://other.com/page">Link</a>`, "https://example.com/file.md", ""},
		// With base URL
		{`<a href="other.md">Link</a>`, "https://example.com/repo/README.md", "https://cooked.example.com"},
		// Non-markdown files
		{`<img src="image.png">`, "https://example.com/repo/README.md", ""},
		{`<a href="archive.zip">Download</a>`, "https://example.com/repo/README.md", ""},
		// Edge cases
		{`no html tags`, "https://example.com/file.md", ""},
		{"", "https://example.com/file.md", ""},
		{`<a href="">empty</a>`, "https://example.com/file.md", ""},
		// Unicode
		{`<a href="日本語.md">Link</a>`, "https://example.com/repo/README.md", ""},
		{`<img src="images/图片.png">`, "https://例え.jp/repo/README.md", ""},
		// Multiple links
		{`<a href="a.md">A</a><a href="b.md">B</a>`, "https://example.com/repo/README.md", ""},
		// Malformed upstream URL
		{`<a href="other.md">Link</a>`, "://bad-url", ""},
		{`<a href="other.md">Link</a>`, "", ""},
	}
	for _, s := range seeds {
		f.Add(s.html, s.upstreamURL, s.baseURL)
	}

	f.Fuzz(func(t *testing.T, html, upstreamURL, baseURL string) {
		got := RelativeURLs([]byte(html), upstreamURL, baseURL, "/_cooked/raw/")

		// Determinism
		got2 := RelativeURLs([]byte(html), upstreamURL, baseURL, "/_cooked/raw/")
		if string(got) != string(got2) {
			t.Errorf("non-deterministic output")
		}

		output := string(got)

		// Absolute URLs in the input must remain untouched
		if strings.Contains(html, `href="https://`) {
			if !strings.Contains(output, "https://") {
				t.Error("absolute https URL was removed")
			}
		}

		// Fragment-only links must remain untouched
		if strings.Contains(html, `href="#`) && !strings.Contains(output, "#") {
			t.Error("fragment-only link was removed")
		}

		// Data URIs must remain untouched
		if strings.Contains(html, `"data:`) && !strings.Contains(output, "data:") {
			t.Error("data URI was removed")
		}

		// Mailto links must remain untouched
		if strings.Contains(html, `"mailto:`) && !strings.Contains(output, "mailto:") {
			t.Error("mailto link was removed")
		}
	})
}

func TestRelativeURLs_QueryString(t *testing.T) {
	html := []byte(`<a href="other.md?ref=main">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "", ""))

	if !strings.Contains(got, "?ref=main") {
		t.Errorf("query string was lost: %s", got)
	}
}
