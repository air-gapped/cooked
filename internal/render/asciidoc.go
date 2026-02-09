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
	case logrus.ErrorLevel:
		level = slog.LevelError
	case logrus.FatalLevel:
		// Use a custom level higher than Error to reflect fatal severity.
		level = slog.Level(slog.LevelError + 1)
	case logrus.PanicLevel:
		// Use a custom level higher than Fatal to reflect panic severity.
		level = slog.Level(slog.LevelError + 2)
	default:
		// Fallback for any other levels (e.g., info/debug) if they ever appear.
		level = slog.LevelInfo
	}

	slog.LogAttrs(context.Background(), level, entry.Message, attrs...)
	return nil
}

// AsciiDocRenderer renders AsciiDoc content to HTML.
type AsciiDocRenderer struct{}

// NewAsciiDocRenderer creates a new AsciiDoc renderer.
func NewAsciiDocRenderer() *AsciiDocRenderer {
	// Logging for libasciidoc is configured per-render in Render to avoid
	// mutating the global logrus logger used elsewhere in the application.
	return &AsciiDocRenderer{}
}

// Render converts AsciiDoc source to HTML and extracts metadata.
func (r *AsciiDocRenderer) Render(source []byte) ([]byte, *MarkdownMeta, error) {
	// Neutralize include:: directives — remote documents can't resolve local includes.
	safe := includeRe.ReplaceAll(source, []byte("// include (not available): $2"))

	cfg := configuration.NewConfiguration()

	// Configure a logger specifically for libasciidoc so that we don't
	// affect the global logrus configuration used by other components.
	logger := logrus.New()
	logger.SetOutput(io.Discard) // suppress default text output
	logger.AddHook(&slogHook{})  // forward to slog instead

	// If the configuration type supports injecting a logger, assign it here
	// so that libasciidoc uses our locally configured logger instance.
	cfg.Logger = logger

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
