package render

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"

	"github.com/bytesparadise/libasciidoc"
	"github.com/bytesparadise/libasciidoc/pkg/configuration"
	"github.com/sirupsen/logrus"
)

// includeRe matches AsciiDoc include:: directives that would try to read local files.
var includeRe = regexp.MustCompile(`(?m)^(include::)(.+\[.*\])\s*$`)

// slogHook is a logrus hook that forwards log entries to slog.
type slogHook struct{}

func (h *slogHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (h *slogHook) Fire(entry *logrus.Entry) error {
	attrs := make([]slog.Attr, 0, len(entry.Data)+1)
	attrs = append(attrs, slog.String("source", "libasciidoc"))
	for k, v := range entry.Data {
		attrs = append(attrs, slog.Any(k, v))
	}

	var level slog.Level
	switch entry.Level {
	case logrus.WarnLevel:
		level = slog.LevelWarn
	default:
		level = slog.LevelError
	}

	slog.LogAttrs(context.Background(), level, entry.Message, attrs...)
	return nil
}

// AsciiDocRenderer renders AsciiDoc content to HTML.
type AsciiDocRenderer struct{}

// NewAsciiDocRenderer creates a new AsciiDoc renderer.
func NewAsciiDocRenderer() *AsciiDocRenderer {
	// Redirect logrus (used by libasciidoc) into our structured slog output.
	logrus.SetOutput(io.Discard) // suppress default text output
	logrus.AddHook(&slogHook{})  // forward to slog instead
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
