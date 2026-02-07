package render

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeGolden(t *testing.T) {
	r := NewCodeRenderer()

	fixtures, err := filepath.Glob(filepath.Join(fixturesDir, "code", "*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no code fixtures found")
	}

	goldenSubdir := filepath.Join(goldenDir, "code")

	for _, fixturePath := range fixtures {
		filename := filepath.Base(fixturePath)
		info := DetectFile(fixturePath)
		if info.ContentType != TypeCode {
			continue
		}

		// Use full filename as test name to disambiguate sample.go vs sample.py
		t.Run(filename, func(t *testing.T) {
			input, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			got, err := r.Render(input, info.Language)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			goldenPath := filepath.Join(goldenSubdir, filename+".html")

			if *update {
				if err := os.MkdirAll(goldenSubdir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file not found (run with -update to create): %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s (run with -update to regenerate)\n"+
					"got %d bytes, want %d bytes\n"+
					"first diff at byte %d",
					filename, len(got), len(want), firstDiff(got, want))
			}

			// Verify SPEC code block structure
			s := string(got)
			if !strings.Contains(s, `class="cooked-code-block"`) {
				t.Error("missing cooked-code-block class")
			}
			if !strings.Contains(s, `data-language="`+info.Language+`"`) {
				t.Errorf("missing data-language=%q", info.Language)
			}
			if !strings.Contains(s, `class="cooked-copy-btn"`) {
				t.Error("missing copy button")
			}
			if !strings.Contains(s, `class="cooked-code-language"`) {
				t.Error("missing language label")
			}
		})
	}
}

func TestPlaintextGolden(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join(fixturesDir, "plaintext", "*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no plaintext fixtures found")
	}

	goldenSubdir := filepath.Join(goldenDir, "plaintext")

	for _, fixturePath := range fixtures {
		filename := filepath.Base(fixturePath)

		t.Run(filename, func(t *testing.T) {
			input, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			got := RenderPlaintext(input)

			goldenPath := filepath.Join(goldenSubdir, filename+".html")

			if *update {
				if err := os.MkdirAll(goldenSubdir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file not found (run with -update to create): %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s (run with -update to regenerate)\n"+
					"got %d bytes, want %d bytes\n"+
					"first diff at byte %d",
					filename, len(got), len(want), firstDiff(got, want))
			}

			// Verify plaintext structure
			s := string(got)
			if !strings.Contains(s, "<pre><code>") {
				t.Error("missing pre/code wrapper")
			}
		})
	}
}
