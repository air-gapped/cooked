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
