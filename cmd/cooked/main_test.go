package main_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// buildBinary compiles the cooked binary into t.TempDir and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "cooked")
	cmd := exec.Command("go", "build",
		"-ldflags", `-X main.version=v0.0.0-test -X main.commit=abc1234 -X main.date=2026-01-01T00:00:00Z`,
		"-o", bin,
		"./",
	)
	cmd.Dir = filepath.Join(".") // run from cmd/cooked
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestVersion(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\n%s", err, out)
	}

	got := strings.TrimSpace(string(out))
	want := "cooked v0.0.0-test (abc1234) built 2026-01-01T00:00:00Z"
	if got != want {
		t.Errorf("--version output = %q, want %q", got, want)
	}
}

func TestGracefulShutdown(t *testing.T) {
	bin := buildBinary(t)

	// Find a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--listen", addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	// Wait for the server to be ready by polling /healthz.
	healthURL := fmt.Sprintf("http://%s/healthz", addr)
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		cmd.Process.Kill()
		t.Fatal("server did not become ready within 5s")
	}

	// Send SIGTERM for graceful shutdown.
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		cmd.Process.Kill()
		t.Fatalf("send signal: %v", err)
	}

	// Wait for process to exit.
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	select {
	case err := <-waitDone:
		if err != nil {
			t.Fatalf("process exited with error: %v", err)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatal("server did not shut down within 10s after SIGTERM")
	}

	// Verify the server is no longer listening.
	_, err = http.Get(healthURL)
	if err == nil {
		t.Error("server still responding after shutdown")
	}
}
