package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cookedembed "github.com/air-gapped/cooked/embed"
	"github.com/air-gapped/cooked/internal/config"
	"github.com/air-gapped/cooked/internal/server"
)

// Set by linker via -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Check for --version before full flag parsing
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Printf("cooked %s (%s) built %s\n", version, commit, date)
			os.Exit(0)
		}
	}

	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cooked: %v\n", err)
		os.Exit(1)
	}

	// Set up structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if cfg.TLSSkipVerify {
		slog.Warn("TLS certificate verification disabled for upstream fetches")
	}

	slog.Info("config loaded",
		"listen", cfg.Listen,
		"cache_ttl", cfg.CacheTTL.String(),
		"cache_max_size", cfg.CacheMaxSize,
		"fetch_timeout", cfg.FetchTimeout.String(),
		"max_file_size", cfg.MaxFileSize,
		"default_theme", cfg.DefaultTheme,
		"tls_skip_verify", cfg.TLSSkipVerify,
	)

	// Create server with all dependencies
	srv := server.New(cfg, version, cookedembed.Assets)

	httpServer := &http.Server{
		Addr:    cfg.Listen,
		Handler: srv.Handler(),
	}

	// Start server in background
	go func() {
		slog.Info("server started", "listen", cfg.Listen, "version", version)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down")

	// Graceful shutdown with 30s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("shutdown complete")
}
