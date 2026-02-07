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

func TestRelativeURLs_QueryString(t *testing.T) {
	html := []byte(`<a href="other.md?ref=main">Link</a>`)
	got := string(RelativeURLs(html, "https://example.com/repo/README.md", ""))

	if !strings.Contains(got, "?ref=main") {
		t.Errorf("query string was lost: %s", got)
	}
}
