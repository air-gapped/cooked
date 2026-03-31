package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestSetup_SetsDefaultLogger(t *testing.T) {
	before := slog.Default()
	logger := Setup()
	after := slog.Default()

	if logger == nil {
		t.Fatal("Setup() returned nil")
	}
	if after == before {
		t.Error("Setup() did not change slog.Default()")
	}

	// Restore to avoid affecting other tests
	slog.SetDefault(before)
}

func TestLogRequest_Fields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	LogRequest(logger, RequestFields{
		Method:      "GET",
		Path:        "/https://example.com/README.md",
		Upstream:    "https://example.com/README.md",
		Status:      200,
		Cache:       "miss",
		UpstreamMs:  45,
		RenderMs:    12,
		TotalMs:     57,
		ContentType: "markdown",
		Bytes:       14832,
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	checks := map[string]any{
		"msg":          "request",
		"method":       "GET",
		"path":         "/https://example.com/README.md",
		"upstream":     "https://example.com/README.md",
		"cache":        "miss",
		"content_type": "markdown",
	}

	for k, want := range checks {
		if entry[k] != want {
			t.Errorf("%s = %v, want %v", k, entry[k], want)
		}
	}

	// Numeric fields come as float64 from JSON
	numChecks := map[string]float64{
		"status":      200,
		"upstream_ms": 45,
		"render_ms":   12,
		"total_ms":    57,
		"bytes":       14832,
	}
	for k, want := range numChecks {
		got, ok := entry[k].(float64)
		if !ok {
			t.Errorf("%s is not a number: %v", k, entry[k])
			continue
		}
		if got != want {
			t.Errorf("%s = %v, want %v", k, got, want)
		}
	}
}

func TestLogRequest_Levels(t *testing.T) {
	tests := []struct {
		status    int
		wantLevel string
	}{
		{200, "INFO"},
		{304, "INFO"},
		{404, "WARN"},
		{403, "WARN"},
		{500, "ERROR"},
		{502, "ERROR"},
	}

	for _, tc := range tests {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		LogRequest(logger, RequestFields{Status: tc.status})

		var entry map[string]any
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("status %d: not valid JSON: %v", tc.status, err)
		}
		if entry["level"] != tc.wantLevel {
			t.Errorf("status %d: level = %v, want %v", tc.status, entry["level"], tc.wantLevel)
		}
	}
}

func TestContextLogger(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	ctx := WithLogger(context.Background(), logger)
	got := FromContext(ctx)
	if got != logger {
		t.Error("FromContext did not return the logger set with WithLogger")
	}
}

func TestFromContext_Default(t *testing.T) {
	got := FromContext(context.Background())
	if got == nil {
		t.Error("FromContext returned nil for empty context")
	}
}

func TestByteCountingWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &ByteCountingWriter{ResponseWriter: rec}

	w.WriteHeader(201)
	if w.StatusCode != 201 {
		t.Errorf("StatusCode = %d, want 201", w.StatusCode)
	}

	n, err := w.Write([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Errorf("Write returned %d, want 11", n)
	}
	if w.Bytes != 11 {
		t.Errorf("Bytes = %d, want 11", w.Bytes)
	}
}

func TestByteCountingWriter_DefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &ByteCountingWriter{ResponseWriter: rec}

	w.Write([]byte("hi"))
	if w.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200 (default)", w.StatusCode)
	}
}
