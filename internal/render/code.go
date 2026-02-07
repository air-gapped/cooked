package render

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// CodeRenderer renders code files with syntax highlighting.
type CodeRenderer struct {
	formatter *chromahtml.Formatter
}

// NewCodeRenderer creates a code renderer using chroma CSS classes.
func NewCodeRenderer() *CodeRenderer {
	return &CodeRenderer{
		formatter: chromahtml.New(
			chromahtml.WithClasses(true),
			chromahtml.WithLineNumbers(true),
		),
	}
}

// Render highlights source code and wraps it in the SPEC code block structure.
func (r *CodeRenderer) Render(source []byte, language string) ([]byte, error) {
	code := string(source)
	lineCount := strings.Count(code, "\n")
	if len(code) > 0 && code[len(code)-1] != '\n' {
		lineCount++
	}

	// Find lexer
	var lexer chroma.Lexer
	if language != "" {
		lexer = lexers.Get(language)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, fmt.Errorf("tokenize code: %w", err)
	}

	// Format with chroma
	var highlighted bytes.Buffer
	if err := r.formatter.Format(&highlighted, styles.Fallback, iterator); err != nil {
		return nil, fmt.Errorf("format code: %w", err)
	}

	// Wrap in SPEC code block structure
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<div class="cooked-code-block" data-language="%s" data-line-count="%d">`, html.EscapeString(language), lineCount)
	fmt.Fprintf(&buf, "\n  <div class=\"cooked-code-header\">\n")
	fmt.Fprintf(&buf, `    <span class="cooked-code-language">%s</span>`, html.EscapeString(language))
	fmt.Fprintf(&buf, "\n    <button class=\"cooked-copy-btn\" data-state=\"idle\">Copy</button>\n")
	fmt.Fprintf(&buf, "  </div>\n")
	buf.Write(highlighted.Bytes())
	fmt.Fprintf(&buf, "\n</div>")

	return buf.Bytes(), nil
}

// RenderPlaintext renders plain text content as monospace pre-formatted text.
func RenderPlaintext(source []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("<pre><code>")
	buf.WriteString(html.EscapeString(string(source)))
	buf.WriteString("</code></pre>")
	return buf.Bytes()
}
