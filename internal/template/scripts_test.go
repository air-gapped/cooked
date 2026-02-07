package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/air-gapped/cooked/internal/render"
)

func TestWriteScripts_ThemeCycling(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	// The theme cycles auto → light → dark → auto
	if !strings.Contains(script, `current === 'auto' ? 'light'`) {
		t.Error("missing auto → light transition")
	}
	if !strings.Contains(script, `current === 'light' ? 'dark'`) {
		t.Error("missing light → dark transition")
	}
	// dark falls through to 'auto' via the ternary
	if !strings.Contains(script, `: 'auto'`) {
		t.Error("missing dark → auto transition")
	}
}

func TestWriteScripts_CookieSetting(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.Contains(script, `_cooked_theme=`) {
		t.Error("missing theme cookie name")
	}
	if !strings.Contains(script, `path=/`) {
		t.Error("missing cookie path")
	}
	if !strings.Contains(script, `max-age=31536000`) {
		t.Error("missing cookie max-age (1 year)")
	}
	if !strings.Contains(script, `SameSite=Lax`) {
		t.Error("missing SameSite attribute")
	}
}

func TestWriteScripts_URLParamOverride(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.Contains(script, `URLSearchParams`) {
		t.Error("missing URLSearchParams usage")
	}
	if !strings.Contains(script, `params.get('_cooked_theme')`) {
		t.Error("missing _cooked_theme param extraction")
	}
	// Should validate against known values
	if !strings.Contains(script, `['auto','light','dark']`) {
		t.Error("missing theme value validation list")
	}
}

func TestWriteScripts_ThemeToggleButton(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.Contains(script, `getElementById('cooked-theme-toggle')`) {
		t.Error("missing theme toggle button lookup")
	}
	// Should have icon mappings
	if !strings.Contains(script, `auto:`) && !strings.Contains(script, `light:`) {
		t.Error("missing theme icon mappings")
	}
	// Should update button on click
	if !strings.Contains(script, `addEventListener('click'`) {
		t.Error("missing click event listener")
	}
}

func TestWriteScripts_TOCToggle(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.Contains(script, `getElementById('cooked-toc-toggle')`) {
		t.Error("missing TOC toggle button lookup")
	}
	if !strings.Contains(script, `getElementById('cooked-toc')`) {
		t.Error("missing TOC element lookup")
	}
	if !strings.Contains(script, `toc.hidden = !toc.hidden`) {
		t.Error("missing TOC hidden toggle")
	}
}

func TestWriteScripts_TOCMobileClose(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	// On mobile (<=768px), clicking a TOC link should close the TOC
	if !strings.Contains(script, `window.innerWidth <= 768`) {
		t.Error("missing mobile width check for TOC close")
	}
	if !strings.Contains(script, `e.target.tagName === 'A'`) {
		t.Error("missing anchor tag check for TOC link click")
	}
}

func TestWriteScripts_CopyButton(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.Contains(script, `.cooked-copy-btn`) {
		t.Error("missing copy button selector")
	}
	if !strings.Contains(script, `navigator.clipboard.writeText`) {
		t.Error("missing clipboard API call")
	}
	if !strings.Contains(script, `'Copied!'`) {
		t.Error("missing 'Copied!' feedback text")
	}
	if !strings.Contains(script, `data-state`) {
		t.Error("missing data-state attribute update")
	}
	// Should reset back to "Copy" after timeout
	if !strings.Contains(script, `setTimeout`) {
		t.Error("missing setTimeout for copy button reset")
	}
	if !strings.Contains(script, `'Copy'`) {
		t.Error("missing 'Copy' reset text")
	}
}

func TestWriteScripts_WrappedInScriptTag(t *testing.T) {
	var buf bytes.Buffer
	writeScripts(&buf)
	script := buf.String()

	if !strings.HasPrefix(strings.TrimSpace(script), "<script>") {
		t.Error("script output should start with <script> tag")
	}
	if !strings.HasSuffix(strings.TrimSpace(script), "</script>") {
		t.Error("script output should end with </script> tag")
	}
}

func TestWriteTOC_Structure(t *testing.T) {
	headings := []render.Heading{
		{Level: 1, Text: "Introduction", ID: "introduction"},
		{Level: 2, Text: "Getting Started", ID: "getting-started"},
		{Level: 3, Text: "Prerequisites", ID: "prerequisites"},
	}

	var buf bytes.Buffer
	writeTOC(&buf, headings)
	html := buf.String()

	if !strings.Contains(html, `id="cooked-toc"`) {
		t.Error("missing TOC nav id")
	}
	if !strings.Contains(html, `hidden`) {
		t.Error("TOC should start hidden")
	}
	if !strings.Contains(html, "<nav") {
		t.Error("TOC should be a nav element")
	}
	if !strings.Contains(html, "<ul>") {
		t.Error("TOC should contain a list")
	}
}

func TestWriteTOC_HeadingLevels(t *testing.T) {
	headings := []render.Heading{
		{Level: 1, Text: "Title", ID: "title"},
		{Level: 2, Text: "Section", ID: "section"},
		{Level: 3, Text: "Subsection", ID: "subsection"},
	}

	var buf bytes.Buffer
	writeTOC(&buf, headings)
	html := buf.String()

	if !strings.Contains(html, `data-level="1"`) {
		t.Error("missing data-level for h1")
	}
	if !strings.Contains(html, `data-level="2"`) {
		t.Error("missing data-level for h2")
	}
	if !strings.Contains(html, `data-level="3"`) {
		t.Error("missing data-level for h3")
	}
}

func TestWriteTOC_AnchorLinks(t *testing.T) {
	headings := []render.Heading{
		{Level: 1, Text: "Hello World", ID: "hello-world"},
		{Level: 2, Text: "Usage & Notes", ID: "usage--notes"},
		{Level: 2, Text: "API <Reference>", ID: "api-reference"},
	}

	var buf bytes.Buffer
	writeTOC(&buf, headings)
	html := buf.String()

	if !strings.Contains(html, `href="#hello-world"`) {
		t.Error("missing anchor link for hello-world")
	}
	if !strings.Contains(html, `href="#usage--notes"`) {
		t.Error("missing anchor link for usage--notes")
	}
	// Text should be HTML-escaped
	if !strings.Contains(html, `API &lt;Reference&gt;`) {
		t.Error("TOC text should be HTML-escaped")
	}
	if !strings.Contains(html, `Usage &amp; Notes`) {
		t.Error("TOC text with ampersand should be HTML-escaped")
	}
}

func TestWriteTOC_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeTOC(&buf, nil)
	html := buf.String()

	// Should still produce the nav structure, just empty
	if !strings.Contains(html, `id="cooked-toc"`) {
		t.Error("empty TOC should still have nav element")
	}
	if !strings.Contains(html, "<ul>") {
		t.Error("empty TOC should still have list")
	}
	if strings.Contains(html, "<li") {
		t.Error("empty TOC should have no list items")
	}
}
