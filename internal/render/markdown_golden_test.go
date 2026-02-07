package render

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

const (
	fixturesDir = "../../testdata/fixtures"
	goldenDir   = "../../testdata/golden"
)

func TestMarkdownGolden(t *testing.T) {
	r := NewMarkdownRenderer()

	fixtures, err := filepath.Glob(filepath.Join(fixturesDir, "markdown", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no markdown fixtures found")
	}

	for _, fixturePath := range fixtures {
		name := filepath.Base(fixturePath)
		name = name[:len(name)-len(filepath.Ext(name))]

		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			got, meta, err := r.Render(input)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			goldenPath := filepath.Join(goldenDir, "markdown", name+".html")

			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
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
					name, len(got), len(want), firstDiff(got, want))
			}

			// Sanity check metadata
			if meta.HeadingCount == 0 {
				t.Error("expected at least one heading in metadata")
			}
		})
	}
}

func TestMDXGolden(t *testing.T) {
	r := NewMarkdownRenderer()

	fixtures, err := filepath.Glob(filepath.Join(fixturesDir, "mdx", "*.mdx"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no MDX fixtures found")
	}

	for _, fixturePath := range fixtures {
		name := filepath.Base(fixturePath)
		name = name[:len(name)-len(filepath.Ext(name))]

		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			preprocessed := PreprocessMDX(input)
			got, _, err := r.Render(preprocessed)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			goldenPath := filepath.Join(goldenDir, "markdown", "mdx_"+name+".html")

			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
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
				t.Errorf("output mismatch for mdx_%s (run with -update to regenerate)\n"+
					"got %d bytes, want %d bytes\n"+
					"first diff at byte %d",
					name, len(got), len(want), firstDiff(got, want))
			}
		})
	}
}

// firstDiff returns the byte offset of the first difference between a and b.
func firstDiff(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
