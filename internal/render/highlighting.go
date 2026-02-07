package render

import (
	"bytes"
	"fmt"
	gohtml "html"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ChromaHighlighting is a goldmark extension that syntax-highlights fenced code
// blocks using chroma with CSS classes. It replaces goldmark-highlighting/v2 to
// avoid its stale chroma v2.2.0 pin and removes the need for post-processing
// regex to wrap code blocks in the cooked structure.
type ChromaHighlighting struct{}

func (e *ChromaHighlighting) Extend(md goldmark.Markdown) {
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&chromaRenderer{
				formatter: chromahtml.New(chromahtml.WithClasses(true)),
			}, 500),
		),
	)
}

type chromaRenderer struct {
	formatter *chromahtml.Formatter
}

func (r *chromaRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *chromaRenderer) renderFencedCodeBlock(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.FencedCodeBlock)

	lang := ""
	if n.Info != nil {
		lang = string(n.Language(source))
	}

	// Collect code from line segments.
	var code bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		code.Write(line.Value(source))
	}

	// Run chroma: lexer → tokenise → format.
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code.String())
	if err != nil {
		return ast.WalkStop, fmt.Errorf("chroma tokenise: %w", err)
	}

	var highlighted bytes.Buffer
	if err := r.formatter.Format(&highlighted, styles.Fallback, iterator); err != nil {
		return ast.WalkStop, fmt.Errorf("chroma format: %w", err)
	}

	// Write the cooked wrapper directly — no post-processing needed.
	fmt.Fprintf(w, `<div class="cooked-code-block" data-language="%s">`, gohtml.EscapeString(lang))
	_, _ = w.WriteString("\n<div class=\"cooked-code-header\">\n")
	if lang != "" {
		fmt.Fprintf(w, `<span class="cooked-code-language">%s</span>`, gohtml.EscapeString(lang))
		_ = w.WriteByte('\n')
	}
	_, _ = w.WriteString("<button class=\"cooked-copy-btn\" data-state=\"idle\">Copy</button>\n")
	_, _ = w.WriteString("</div>\n")
	_, _ = w.Write(highlighted.Bytes())
	_, _ = w.WriteString("\n</div>")

	return ast.WalkContinue, nil
}
