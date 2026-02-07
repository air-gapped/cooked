package sanitize

import (
	"github.com/microcosm-cc/bluemonday"
)

// policy is the shared, immutable sanitization policy.
// Built once at init time; bluemonday policies are safe for concurrent use.
var policy *bluemonday.Policy

func init() {
	// Start from UGCPolicy which allows common safe HTML elements
	// (headings, paragraphs, lists, tables, links, images, code blocks, etc.)
	// and strips everything dangerous (script, iframe, object, embed, form,
	// style, meta, base, link, event handlers, javascript:/vbscript: URIs).
	p := bluemonday.UGCPolicy()

	// Allow id on all elements — goldmark generates heading anchors (id="foo")
	// and footnote refs (id="fnref:1", id="fn:1").
	p.AllowAttrs("id").Globally()

	// Allow class on all elements — chroma syntax highlighting uses class
	// on spans/pre, cooked wrappers use class on div/button/span.
	p.AllowAttrs("class").Globally()

	// Allow role on all elements — goldmark footnotes use role="doc-noteref",
	// role="doc-backlink", role="doc-endnotes".
	p.AllowAttrs("role").Globally()

	// Allow data-* attributes — cooked code wrappers use data-language,
	// data-state on buttons.
	p.AllowDataAttributes()

	// Allow tabindex on pre — chroma sets tabindex="0" for keyboard nav.
	p.AllowAttrs("tabindex").OnElements("pre")

	// Allow disabled checkbox inputs for GFM task lists.
	// Only disabled+checked+type are needed; the input is inert.
	p.AllowAttrs("type", "checked", "disabled").OnElements("input")

	// Allow button elements for cooked copy-to-clipboard buttons.
	p.AllowElements("button")

	// Allow del/ins for GFM strikethrough.
	p.AllowElements("del", "ins")

	// Allow definition list elements.
	p.AllowElements("dl", "dt", "dd")

	// Allow details/summary.
	p.AllowElements("details", "summary")

	policy = p
}

// HTML strips dangerous elements and attributes from HTML content.
// This processes the HTML after goldmark rendering but BEFORE cooked's own
// scripts are injected (so cooked's scripts are never stripped).
func HTML(input []byte) []byte {
	return policy.SanitizeBytes(input)
}
