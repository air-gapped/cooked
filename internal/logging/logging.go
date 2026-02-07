package logging

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Setup initializes the default slog logger with JSON output to stdout.
func Setup() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}

type contextKey string

const loggerKey contextKey = "logger"

// WithLogger returns a context with the given logger attached.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context, falling back to slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// RequestFields holds all fields logged per request, matching SPEC format.
type RequestFields struct {
	Method      string
	Path        string
	Upstream    string
	Status      int
	Cache       string
	UpstreamMs  int64
	RenderMs    int64
	TotalMs     int64
	ContentType string
	Bytes       int64
}

// LogRequest logs a completed request with structured fields.
func LogRequest(logger *slog.Logger, f RequestFields) {
	level := slog.LevelInfo
	if f.Status >= 500 {
		level = slog.LevelError
	} else if f.Status >= 400 {
		level = slog.LevelWarn
	}

	logger.Log(context.Background(), level, "request",
		"method", f.Method,
		"path", f.Path,
		"upstream", f.Upstream,
		"status", f.Status,
		"cache", f.Cache,
		"upstream_ms", f.UpstreamMs,
		"render_ms", f.RenderMs,
		"total_ms", f.TotalMs,
		"content_type", f.ContentType,
		"bytes", f.Bytes,
	)
}

// ByteCountingWriter wraps http.ResponseWriter to capture status code and bytes written.
type ByteCountingWriter struct {
	http.ResponseWriter
	StatusCode int
	Bytes      int64
}

// WriteHeader captures the status code.
func (w *ByteCountingWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures bytes written.
func (w *ByteCountingWriter) Write(b []byte) (int, error) {
	if w.StatusCode == 0 {
		w.StatusCode = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.Bytes += int64(n)
	return n, err
}

// Middleware returns an HTTP middleware that logs requests with timing.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &ByteCountingWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)

		if wrapped.StatusCode == 0 {
			wrapped.StatusCode = 200
		}

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.StatusCode,
			"bytes", wrapped.Bytes,
			"total_ms", time.Since(start).Milliseconds(),
		)
	})
}
