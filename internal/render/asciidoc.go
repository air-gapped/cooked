package render

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/bytesparadise/libasciidoc"
	"github.com/bytesparadise/libasciidoc/pkg/configuration"
	logrus "github.com/sirupsen/logrus"
)

// includeRe matches AsciiDoc include:: directives that would try to read local files.
var includeRe = regexp.MustCompile(`(?m)^(include::)(.+\[.*\])\s*$`)

// AsciiDocRenderer renders AsciiDoc content to HTML.
type AsciiDocRenderer struct{}

// NewAsciiDocRenderer creates a new AsciiDoc renderer.
// Silences logrus output from libasciidoc to keep our JSON logging clean.
func NewAsciiDocRenderer() *AsciiDocRenderer {
	logrus.SetLevel(logrus.FatalLevel)
	return &AsciiDocRenderer{}
}

// Render converts AsciiDoc source to HTML and extracts metadata.
func (r *AsciiDocRenderer) Render(source []byte) ([]byte, *MarkdownMeta, error) {
	// Neutralize include:: directives — remote documents can't resolve local includes.
	safe := includeRe.ReplaceAll(source, []byte("// include (not available): $2"))

	cfg := configuration.NewConfiguration()

	var buf bytes.Buffer
	metadata, err := libasciidoc.Convert(bytes.NewReader(safe), &buf, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("render asciidoc: %w", err)
	}

	meta := &MarkdownMeta{
		Title: metadata.Title,
	}

	return buf.Bytes(), meta, nil
}
