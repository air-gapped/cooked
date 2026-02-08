package render

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/niklasfasching/go-org/org"
)

// OrgRenderer renders Org-mode content to HTML.
type OrgRenderer struct{}

// NewOrgRenderer creates a new Org-mode renderer.
func NewOrgRenderer() *OrgRenderer {
	return &OrgRenderer{}
}

// Render converts Org-mode source to HTML and extracts metadata.
func (r *OrgRenderer) Render(source []byte) ([]byte, *MarkdownMeta, error) {
	conf := org.New()
	conf.Log = log.New(io.Discard, "", 0) // suppress warnings

	writer := org.NewHTMLWriter()
	writer.TopLevelHLevel = 1 // map * headings to <h1>

	doc := conf.Parse(bytes.NewReader(source), "")
	htmlStr, err := doc.Write(writer)
	if err != nil {
		return nil, nil, fmt.Errorf("render org: %w", err)
	}

	meta := &MarkdownMeta{}

	// Extract title from #+TITLE keyword or first headline
	if title, ok := doc.BufferSettings["TITLE"]; ok && title != "" {
		meta.Title = title
	} else {
		meta.Title = firstOrgHeadlineTitle(doc)
	}

	return []byte(htmlStr), meta, nil
}

// firstOrgHeadlineTitle returns the text of the first headline in the document.
func firstOrgHeadlineTitle(doc *org.Document) string {
	for _, node := range doc.Nodes {
		if h, ok := node.(org.Headline); ok {
			return strings.TrimSpace(org.String(h.Title...))
		}
	}
	return ""
}
