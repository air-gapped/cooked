package template

import (
	"html/template"
	"strings"
	"testing"

	"github.com/air-gapped/cooked/internal/render"
)

func TestRenderPage_RequiredIDs(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderPage(PageData{
		Version:      "v0.1.0",
		UpstreamURL:  "https://example.com/README.md",
		ContentType:  render.TypeMarkdown,
		CacheStatus:  "miss",
		DefaultTheme: "auto",
		Title:        "README",
		Content:      template.HTML("<h1>Hello</h1>"),
		HeadingCount: 5,
		Headings: []render.Heading{
			{Level: 1, Text: "A", ID: "a"},
			{Level: 2, Text: "B", ID: "b"},
			{Level: 2, Text: "C", ID: "c"},
			{Level: 3, Text: "D", ID: "d"},
			{Level: 2, Text: "E", ID: "e"},
		},
	}, "", ""))

	requiredIDs := []string{
		`id="cooked-header"`,
		`id="cooked-source-link"`,
		`id="cooked-size"`,
		`id="cooked-type"`,
		`id="cooked-theme-toggle"`,
		`id="cooked-content"`,
		`id="cooked-toc-toggle"`,
		`id="cooked-toc"`,
	}

	for _, id := range requiredIDs {
		if !strings.Contains(html, id) {
			t.Errorf("missing required ID: %s", id)
		}
	}
}

func TestRenderPage_DataAttributes(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderPage(PageData{
		Version:        "v0.1.0",
		UpstreamURL:    "https://example.com/README.md",
		ContentType:    render.TypeMarkdown,
		CacheStatus:    "miss",
		UpstreamStatus: 200,
		FileSize:       14832,
		DefaultTheme:   "auto",
		Content:        template.HTML("<p>Hello</p>"),
		HasMermaid:     true,
		HeadingCount:   2,
		CodeBlockCount: 5,
	}, "", ""))

	checks := []string{
		`data-theme="auto"`,
		`data-cooked-version="v0.1.0"`,
		`data-upstream-url="https://example.com/README.md"`,
		`data-content-type="markdown"`,
		`data-cache-status="miss"`,
		`data-upstream-status="200"`,
		`data-file-size="14832"`,
		`data-has-mermaid="true"`,
		`data-heading-count="2"`,
		`data-code-block-count="5"`,
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("missing data attribute: %s", check)
		}
	}
}

func TestRenderPage_NoTOCWithFewHeadings(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		Content:      template.HTML("<p>Hello</p>"),
		HeadingCount: 2,
		Headings:     []render.Heading{{Level: 1, Text: "A"}, {Level: 2, Text: "B"}},
	}, "", ""))

	if strings.Contains(html, `id="cooked-toc-toggle"`) {
		t.Error("TOC toggle should not be present with < 3 headings")
	}
	if strings.Contains(html, `id="cooked-toc"`) {
		t.Error("TOC nav should not be present with < 3 headings")
	}
}

func TestRenderPage_HTMLComments(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	comments := []string{
		"<!-- cooked: header -->",
		"<!-- cooked: table of contents -->",
		"<!-- cooked: content -->",
		"<!-- cooked: scripts -->",
	}

	for _, comment := range comments {
		if !strings.Contains(html, comment) {
			t.Errorf("missing section comment: %s", comment)
		}
	}
}

func TestRenderPage_MermaidScript(t *testing.T) {
	r := NewRenderer()

	// With mermaid
	html := string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		HasMermaid:   true,
		MermaidPath:  "/_cooked/mermaid.min.js",
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	if !strings.Contains(html, `src="/_cooked/mermaid.min.js"`) {
		t.Error("mermaid script tag should be present when HasMermaid=true")
	}

	// Without mermaid
	html = string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		HasMermaid:   false,
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	if strings.Contains(html, "mermaid.min.js") {
		t.Error("mermaid script should not be present when HasMermaid=false")
	}
}

func TestRenderPage_PrintCSS(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	if !strings.Contains(html, "@media print") {
		t.Error("missing print media query")
	}
}

func TestRenderError_RequiredElements(t *testing.T) {
	r := NewRenderer()
	html := string(r.RenderError(ErrorData{
		Version:      "v0.1.0",
		UpstreamURL:  "https://example.com/missing.md",
		StatusCode:   404,
		ErrorType:    "upstream-error",
		Message:      "Not Found",
		DefaultTheme: "auto",
	}))

	checks := []string{
		`id="cooked-error"`,
		`data-status-code="404"`,
		`data-error-type="upstream-error"`,
		`data-content-type="error"`,
		`id="cooked-source-link"`,
		`id="cooked-theme-toggle"`,
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("error page missing: %s", check)
		}
	}

	// Should include link to upstream
	if !strings.Contains(html, "View original file") {
		t.Error("error page missing link to upstream")
	}
}

func TestRenderPage_Title(t *testing.T) {
	r := NewRenderer()

	// Title from data
	html := string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		Title:        "My Document",
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	if !strings.Contains(html, "<title>My Document — cooked</title>") {
		t.Error("title not set from data.Title")
	}

	// Title from URL when no title provided
	html = string(r.RenderPage(PageData{
		DefaultTheme: "auto",
		UpstreamURL:  "https://example.com/README.md",
		Content:      template.HTML("<p>Hello</p>"),
	}, "", ""))

	if !strings.Contains(html, "<title>README.md — cooked</title>") {
		t.Error("title not extracted from URL")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{14832, "14.5 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
	}

	for _, tc := range tests {
		got := formatFileSize(tc.bytes)
		if got != tc.want {
			t.Errorf("formatFileSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestTruncateURL(t *testing.T) {
	short := "https://example.com/file.md"
	if truncateURL(short, 80) != short {
		t.Error("short URL should not be truncated")
	}

	long := strings.Repeat("a", 100)
	got := truncateURL(long, 80)
	if len(got) != 80 {
		t.Errorf("truncated length = %d, want 80", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("truncated URL should end with ...")
	}
}
