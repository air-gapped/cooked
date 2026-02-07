package render

import (
	"strings"
	"testing"
)

func TestMarkdownRenderer_BasicMarkdown(t *testing.T) {
	r := NewMarkdownRenderer()
	html, meta, err := r.Render([]byte("# Hello\n\nWorld\n"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "<h1") {
		t.Error("expected <h1> in output")
	}
	if !strings.Contains(string(html), "Hello") {
		t.Error("expected Hello in output")
	}
	if !strings.Contains(string(html), "<p>World</p>") {
		t.Error("expected <p>World</p> in output")
	}

	if meta.HeadingCount != 1 {
		t.Errorf("HeadingCount = %d, want 1", meta.HeadingCount)
	}
	if meta.Title != "Hello" {
		t.Errorf("Title = %q, want Hello", meta.Title)
	}
}

func TestMarkdownRenderer_GFMFeatures(t *testing.T) {
	r := NewMarkdownRenderer()

	// Tables
	input := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	html, _, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "<table>") {
		t.Error("expected <table> for GFM table")
	}

	// Strikethrough
	input = "~~deleted~~\n"
	html, _, err = r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "<del>") {
		t.Error("expected <del> for strikethrough")
	}

	// Task lists
	input = "- [x] done\n- [ ] todo\n"
	html, _, err = r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "checkbox") || !strings.Contains(string(html), "checked") {
		t.Error("expected checkbox elements for task list")
	}

	// Autolinks
	input = "Visit https://example.com for details\n"
	html, _, err = r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `href="https://example.com"`) {
		t.Error("expected autolinked URL")
	}
}

func TestMarkdownRenderer_SyntaxHighlighting(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "```go\nfunc main() {}\n```\n"
	html, meta, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if meta.CodeBlockCount != 1 {
		t.Errorf("CodeBlockCount = %d, want 1", meta.CodeBlockCount)
	}

	// Should use CSS classes for highlighting
	if !strings.Contains(string(html), "class=") {
		t.Error("expected CSS class-based highlighting")
	}
}

func TestMarkdownRenderer_MermaidDetection(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "```mermaid\ngraph TD\n  A-->B\n```\n"
	_, meta, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if !meta.HasMermaid {
		t.Error("HasMermaid = false, want true")
	}
}

func TestMarkdownRenderer_NoMermaid(t *testing.T) {
	r := NewMarkdownRenderer()
	_, meta, err := r.Render([]byte("# No mermaid here\n"))
	if err != nil {
		t.Fatal(err)
	}

	if meta.HasMermaid {
		t.Error("HasMermaid = true, want false")
	}
}

func TestMarkdownRenderer_HeadingIDs(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "## Foo Bar\n\n## Hello World\n"
	html, meta, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if meta.HeadingCount != 2 {
		t.Errorf("HeadingCount = %d, want 2", meta.HeadingCount)
	}

	// Auto heading IDs should be generated
	s := string(html)
	if !strings.Contains(s, `id="foo-bar"`) {
		t.Error("expected id=\"foo-bar\" for heading")
	}
	if !strings.Contains(s, `id="hello-world"`) {
		t.Error("expected id=\"hello-world\" for heading")
	}
}

func TestMarkdownRenderer_TOCHeadings(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "# Title\n## Section 1\n### Sub 1.1\n## Section 2\n"
	_, meta, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(meta.Headings) != 4 {
		t.Fatalf("got %d headings, want 4", len(meta.Headings))
	}

	checks := []struct {
		level int
		text  string
	}{
		{1, "Title"},
		{2, "Section 1"},
		{3, "Sub 1.1"},
		{2, "Section 2"},
	}

	for i, c := range checks {
		if meta.Headings[i].Level != c.level {
			t.Errorf("heading[%d].Level = %d, want %d", i, meta.Headings[i].Level, c.level)
		}
		if meta.Headings[i].Text != c.text {
			t.Errorf("heading[%d].Text = %q, want %q", i, meta.Headings[i].Text, c.text)
		}
	}
}

func TestMarkdownRenderer_Frontmatter(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "---\ntitle: My Document\nauthor: Test\n---\n\n# Heading\n\nContent\n"
	html, meta, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if meta.Title != "My Document" {
		t.Errorf("Title = %q, want My Document", meta.Title)
	}

	// Frontmatter should not appear in output
	if strings.Contains(string(html), "---") {
		t.Error("frontmatter should be stripped from output")
	}
}

func TestMarkdownRenderer_UnsafeHTML(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "<details>\n<summary>Click me</summary>\nHidden content\n</details>\n"
	html, _, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "<details>") {
		t.Error("unsafe HTML should be preserved in rendering")
	}
}

func TestMarkdownRenderer_Footnotes(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "Text with a footnote[^1].\n\n[^1]: This is the footnote.\n"
	html, _, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "footnote") {
		t.Error("expected footnote content in output")
	}
}

func TestMarkdownRenderer_DefinitionList(t *testing.T) {
	r := NewMarkdownRenderer()
	input := "Term\n:   Definition\n"
	html, _, err := r.Render([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "<dl>") || !strings.Contains(string(html), "<dt>") {
		t.Error("expected definition list elements")
	}
}

func TestStripFrontmatter_NoFrontmatter(t *testing.T) {
	input := []byte("# Hello\nContent\n")
	content, title := stripFrontmatter(input)
	if string(content) != string(input) {
		t.Error("content should be unchanged without frontmatter")
	}
	if title != "" {
		t.Errorf("title should be empty, got %q", title)
	}
}

func TestStripFrontmatter_QuotedTitle(t *testing.T) {
	input := []byte("---\ntitle: \"My Doc\"\n---\nContent\n")
	_, title := stripFrontmatter(input)
	if title != "My Doc" {
		t.Errorf("title = %q, want My Doc", title)
	}
}
