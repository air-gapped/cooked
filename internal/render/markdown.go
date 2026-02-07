package render

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	gmermaid "go.abhg.dev/goldmark/mermaid"
)

// MarkdownMeta holds metadata extracted during rendering.
type MarkdownMeta struct {
	HeadingCount   int
	HasMermaid     bool
	CodeBlockCount int
	Headings       []Heading
	Title          string // from first H1 or frontmatter
}

// Heading represents a heading in the document for TOC generation.
type Heading struct {
	Level int
	Text  string
	ID    string
}

// MarkdownRenderer renders markdown content to HTML.
type MarkdownRenderer struct {
	md goldmark.Markdown
}

// NewMarkdownRenderer creates a new markdown renderer with all SPEC extensions.
func NewMarkdownRenderer() *MarkdownRenderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.DefinitionList,
			extension.Typographer,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
			&gmermaid.Extender{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	return &MarkdownRenderer{md: md}
}

// Render converts markdown source to HTML and extracts metadata.
func (r *MarkdownRenderer) Render(source []byte) ([]byte, *MarkdownMeta, error) {
	// Strip YAML frontmatter before rendering
	content, title := stripFrontmatter(source)

	var buf bytes.Buffer
	reader := text.NewReader(content)
	doc := r.md.Parser().Parse(reader)

	meta := &MarkdownMeta{Title: title}
	extractMeta(doc, content, meta)

	if err := r.md.Renderer().Render(&buf, content, doc); err != nil {
		return nil, nil, fmt.Errorf("render markdown: %w", err)
	}

	return buf.Bytes(), meta, nil
}

// extractMeta walks the AST to count headings, code blocks, and detect mermaid.
func extractMeta(doc ast.Node, source []byte, meta *MarkdownMeta) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			meta.HeadingCount++
			// Extract heading text
			var text strings.Builder
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if t, ok := c.(*ast.Text); ok {
					text.Write(t.Segment.Value(source))
				}
			}
			id := ""
			if idAttr, ok := node.AttributeString("id"); ok {
				if idBytes, ok := idAttr.([]byte); ok {
					id = string(idBytes)
				}
			}
			meta.Headings = append(meta.Headings, Heading{
				Level: node.Level,
				Text:  text.String(),
				ID:    id,
			})
			if meta.Title == "" && node.Level == 1 {
				meta.Title = text.String()
			}

		case *ast.FencedCodeBlock:
			meta.CodeBlockCount++

		default:
			// mermaid extension transforms fenced code blocks into its own node type
			if n.Kind() == gmermaid.Kind {
				meta.HasMermaid = true
			}
		}

		return ast.WalkContinue, nil
	})
}

var frontmatterRe = regexp.MustCompile(`(?s)\A---\n(.+?)\n---\n`)

// stripFrontmatter removes YAML frontmatter and extracts the title field.
func stripFrontmatter(source []byte) ([]byte, string) {
	match := frontmatterRe.FindSubmatch(source)
	if match == nil {
		return source, ""
	}

	// Extract title from frontmatter
	title := ""
	for _, line := range strings.Split(string(match[1]), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			title = strings.Trim(title, "\"'")
			break
		}
	}

	// Return content after frontmatter
	return source[len(match[0]):], title
}
