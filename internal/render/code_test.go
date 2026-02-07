package render

import (
	"strings"
	"testing"
)

func TestCodeRenderer_Python(t *testing.T) {
	r := NewCodeRenderer()
	source := []byte("def hello():\n    print('world')\n")
	html, err := r.Render(source, "python")
	if err != nil {
		t.Fatal(err)
	}

	s := string(html)
	if !strings.Contains(s, `class="cooked-code-block"`) {
		t.Error("missing cooked-code-block class")
	}
	if !strings.Contains(s, `data-language="python"`) {
		t.Error("missing data-language attribute")
	}
	if !strings.Contains(s, `data-line-count="2"`) {
		t.Error("missing or wrong data-line-count")
	}
	if !strings.Contains(s, `class="cooked-code-language"`) {
		t.Error("missing language label")
	}
	if !strings.Contains(s, `class="cooked-copy-btn"`) {
		t.Error("missing copy button")
	}
}

func TestCodeRenderer_Go(t *testing.T) {
	r := NewCodeRenderer()
	source := []byte("package main\n\nfunc main() {\n}\n")
	html, err := r.Render(source, "go")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), `data-language="go"`) {
		t.Error("missing data-language=\"go\"")
	}
}

func TestCodeRenderer_UnknownLanguage(t *testing.T) {
	r := NewCodeRenderer()
	source := []byte("some unknown content\n")
	html, err := r.Render(source, "")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "cooked-code-block") {
		t.Error("should still wrap in code block")
	}
}

func TestCodeRenderer_NoTrailingNewline(t *testing.T) {
	r := NewCodeRenderer()
	source := []byte("line1\nline2")
	html, err := r.Render(source, "text")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), `data-line-count="2"`) {
		t.Error("expected 2 lines for input without trailing newline")
	}
}

func TestCodeRenderer_LineCount(t *testing.T) {
	r := NewCodeRenderer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single line no newline", "hello", `data-line-count="1"`},
		{"single line with newline", "hello\n", `data-line-count="1"`},
		{"two lines with newline", "a\nb\n", `data-line-count="2"`},
		{"two lines no trailing newline", "a\nb", `data-line-count="2"`},
		{"empty", "", `data-line-count="0"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html, err := r.Render([]byte(tc.input), "text")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(html), tc.want) {
				t.Errorf("got %s, want %s in output", string(html), tc.want)
			}
		})
	}
}

func TestRenderPlaintext(t *testing.T) {
	source := []byte("Hello <world> & \"stuff\"")
	html := RenderPlaintext(source)

	s := string(html)
	if !strings.Contains(s, "<pre><code>") {
		t.Error("missing pre/code wrapper")
	}
	// HTML entities should be escaped
	if !strings.Contains(s, "&lt;world&gt;") {
		t.Error("expected HTML escaping of angle brackets")
	}
	if !strings.Contains(s, "&amp;") {
		t.Error("expected HTML escaping of ampersand")
	}
}
