package testdata_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/air-gapped/cooked/internal/render"
)

// TestFixtures_Render verifies that all fixture files can be processed
// by the render pipeline without errors. This catches malformed fixtures
// before they're used by golden file or integration tests.
func TestFixtures_Render(t *testing.T) {
	mdRenderer := render.NewMarkdownRenderer()
	codeRenderer := render.NewCodeRenderer()

	fixtures, err := filepath.Glob("fixtures/*/*")
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixture files found")
	}

	for _, path := range fixtures {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) == 0 {
				t.Fatal("fixture file is empty")
			}

			info := render.DetectFile(path)

			switch info.ContentType {
			case render.TypeMarkdown:
				html, meta, err := mdRenderer.Render(data)
				if err != nil {
					t.Fatalf("markdown render failed: %v", err)
				}
				if len(html) == 0 {
					t.Error("markdown rendered to empty output")
				}
				if meta.HeadingCount == 0 {
					t.Error("expected at least one heading")
				}

			case render.TypeMDX:
				preprocessed := render.PreprocessMDX(data)
				html, _, err := mdRenderer.Render(preprocessed)
				if err != nil {
					t.Fatalf("MDX render failed: %v", err)
				}
				if len(html) == 0 {
					t.Error("MDX rendered to empty output")
				}

			case render.TypeCode:
				html, err := codeRenderer.Render(data, info.Language)
				if err != nil {
					t.Fatalf("code render failed: %v", err)
				}
				if len(html) == 0 {
					t.Error("code rendered to empty output")
				}

			case render.TypePlaintext:
				html := render.RenderPlaintext(data)
				if len(html) == 0 {
					t.Error("plaintext rendered to empty output")
				}

			default:
				t.Errorf("unexpected content type %q for fixture %s", info.ContentType, path)
			}
		})
	}
}
