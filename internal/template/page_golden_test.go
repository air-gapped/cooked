package template

import (
	"bytes"
	"flag"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"github.com/air-gapped/cooked/internal/render"
)

var update = flag.Bool("update", false, "update golden files")

const goldenDir = "../../testdata/golden/template"

// templateTestCase defines a full-page rendering scenario.
type templateTestCase struct {
	name string
	data PageData
}

func templateTestCases() []templateTestCase {
	return []templateTestCase{
		{
			name: "markdown_with_toc",
			data: PageData{
				Version:        "v0.1.0-test",
				UpstreamURL:    "https://example.com/README.md",
				ContentType:    render.TypeMarkdown,
				CacheStatus:    "miss",
				UpstreamStatus: 200,
				FileSize:       4096,
				DefaultTheme:   "auto",
				Title:          "Project README",
				Content:        template.HTML("<h1 id=\"title\">Title</h1>\n<h2 id=\"install\">Install</h2>\n<h2 id=\"usage\">Usage</h2>\n<h3 id=\"basic\">Basic</h3>\n<p>Hello world</p>"),
				HasMermaid:     false,
				HasTOC:         true,
				HeadingCount:   4,
				CodeBlockCount: 0,
				Headings: []render.Heading{
					{Level: 1, Text: "Title", ID: "title"},
					{Level: 2, Text: "Install", ID: "install"},
					{Level: 2, Text: "Usage", ID: "usage"},
					{Level: 3, Text: "Basic", ID: "basic"},
				},
			},
		},
		{
			name: "markdown_with_mermaid",
			data: PageData{
				Version:        "v0.1.0-test",
				UpstreamURL:    "https://example.com/architecture.md",
				ContentType:    render.TypeMarkdown,
				CacheStatus:    "hit",
				UpstreamStatus: 200,
				FileSize:       2048,
				DefaultTheme:   "auto",
				Title:          "Architecture",
				Content:        template.HTML("<h1 id=\"arch\">Architecture</h1>\n<pre class=\"mermaid\">graph TD\n    A-->B</pre>"),
				HasMermaid:     true,
				HeadingCount:   1,
				CodeBlockCount: 0,
				MermaidPath:    "/_cooked/mermaid.min.js",
			},
		},
		{
			name: "code_file",
			data: PageData{
				Version:        "v0.1.0-test",
				UpstreamURL:    "https://example.com/main.py",
				ContentType:    render.TypeCode,
				CacheStatus:    "miss",
				UpstreamStatus: 200,
				FileSize:       512,
				DefaultTheme:   "auto",
				Title:          "main.py",
				Content:        template.HTML("<div class=\"cooked-code-block\" data-language=\"python\">\n<pre><code>print('hello')</code></pre>\n</div>"),
				HeadingCount:   0,
				CodeBlockCount: 1,
			},
		},
		{
			name: "no_toc_few_headings",
			data: PageData{
				Version:        "v0.1.0-test",
				UpstreamURL:    "https://example.com/short.md",
				ContentType:    render.TypeMarkdown,
				CacheStatus:    "miss",
				UpstreamStatus: 200,
				FileSize:       256,
				DefaultTheme:   "auto",
				Title:          "Short Doc",
				Content:        template.HTML("<h1 id=\"hi\">Hi</h1>\n<p>Short.</p>"),
				HeadingCount:   1,
				Headings:       []render.Heading{{Level: 1, Text: "Hi", ID: "hi"}},
			},
		},
	}
}

func TestPageGolden(t *testing.T) {
	r := NewRenderer()

	for _, tc := range templateTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			// Use empty CSS to keep golden files focused on structure.
			got := r.RenderPage(tc.data, "", "")

			goldenPath := filepath.Join(goldenDir, tc.name+".html")

			if *update {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file not found (run with -update to create): %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s (run with -update to regenerate)\n"+
					"got %d bytes, want %d bytes\n"+
					"first diff at byte %d",
					tc.name, len(got), len(want), firstDiff(got, want))
			}
		})
	}
}

func TestErrorPageGolden(t *testing.T) {
	r := NewRenderer()

	cases := []struct {
		name string
		data ErrorData
	}{
		{
			name: "error_404",
			data: ErrorData{
				Version:      "v0.1.0-test",
				UpstreamURL:  "https://example.com/missing.md",
				StatusCode:   404,
				ErrorType:    "upstream-error",
				Message:      "Not Found",
				DefaultTheme: "auto",
			},
		},
		{
			name: "error_502",
			data: ErrorData{
				Version:      "v0.1.0-test",
				UpstreamURL:  "https://down.example.com/doc.md",
				StatusCode:   502,
				ErrorType:    "upstream-error",
				Message:      "Bad Gateway: connection refused",
				DefaultTheme: "auto",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := r.RenderError(tc.data)

			goldenPath := filepath.Join(goldenDir, tc.name+".html")

			if *update {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file not found (run with -update to create): %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s (run with -update to regenerate)\n"+
					"got %d bytes, want %d bytes\n"+
					"first diff at byte %d",
					tc.name, len(got), len(want), firstDiff(got, want))
			}
		})
	}
}

func firstDiff(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
