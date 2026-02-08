package render

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAsciiDocGolden(t *testing.T) {
	r := NewAsciiDocRenderer()

	fixtures, err := filepath.Glob(filepath.Join(fixturesDir, "asciidoc", "*.adoc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no asciidoc fixtures found")
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

			goldenPath := filepath.Join(goldenDir, "asciidoc", name+".html")

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
			if meta.Title == "" {
				t.Error("expected a title in metadata")
			}
		})
	}
}
