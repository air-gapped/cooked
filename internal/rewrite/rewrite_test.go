package rewrite

import (
	"strings"
	"testing"
)

func TestRelativeURLs_MarkdownLinks(t *testing.T) {
	html := []byte(`<a href="CONTRIBUTING.md">Contributing</a>`)
	got := string(RelativeURLs(html, "https://cgit.internal/repo/plain/README.md", ""))

	want := `/https://cgit.internal/repo/plain/CONTRIBUTING.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant href containing %s", got, want)
	}
}

func TestRelativeURLs_MarkdownSubdir(t *testing.T) {
	html := []byte(`<a href="docs/guide.md">Guide</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", ""))

	want := `/https://example.com/repo/docs/guide.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestRelativeURLs_Images(t *testing.T) {
	html := []byte(`<img src="./docs/arch.png" alt="architecture">`)
	got := string(RelativeURLs(html, "https://cgit.internal/repo/plain/README.md", ""))

	// Images should point directly to upstream, not through cooked
	want := `https://cgit.internal/repo/plain/docs/arch.png`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant src containing %s", got, want)
	}
}

func TestRelativeURLs_AbsoluteURLsUntouched(t *testing.T) {
	html := []byte(`<a href="https://example.com/other">Other</a>`)
	got := string(RelativeURLs(html, "https://upstream.com/file.md", ""))

	if !strings.Contains(got, `href="https://example.com/other"`) {
		t.Errorf("absolute URL was modified: %s", got)
	}
}

func TestRelativeURLs_FragmentOnly(t *testing.T) {
	html := []byte(`<a href="#section">Section</a>`)
	got := string(RelativeURLs(html, "https://example.com/file.md", ""))

	if !strings.Contains(got, `href="#section"`) {
		t.Errorf("fragment-only link was modified: %s", got)
	}
}

func TestRelativeURLs_MarkdownWithFragment(t *testing.T) {
	html := []byte(`<a href="other.md#section">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", ""))

	if !strings.Contains(got, "#section") {
		t.Errorf("fragment was lost: %s", got)
	}
	if !strings.Contains(got, "other.md") {
		t.Errorf("markdown link not rewritten: %s", got)
	}
}

func TestRelativeURLs_WithBaseURL(t *testing.T) {
	html := []byte(`<a href="other.md">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", "https://cooked.example.com"))

	want := `https://cooked.example.com/https://example.com/repo/other.md`
	if !strings.Contains(got, want) {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestRelativeURLs_DataURIUntouched(t *testing.T) {
	html := []byte(`<img src="data:image/png;base64,abc">`)
	got := string(RelativeURLs(html, "https://example.com/file.md", ""))

	if !strings.Contains(got, "data:image/png") {
		t.Errorf("data URI was modified: %s", got)
	}
}

func TestRelativeURLs_ProtocolRelativeUntouched(t *testing.T) {
	html := []byte(`<a href="//cdn.example.com/lib.js">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/file.md", ""))

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
		got := RelativeURLs([]byte(html), upstreamURL, baseURL)

		// Determinism
		got2 := RelativeURLs([]byte(html), upstreamURL, baseURL)
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
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", ""))

	if !strings.Contains(got, "?ref=main") {
		t.Errorf("query string was lost: %s", got)
	}
}
