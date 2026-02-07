package sanitize

import (
	"strings"
	"testing"
)

func TestHTML_StripScript(t *testing.T) {
	input := `<p>Hello</p><script>alert('xss')</script><p>World</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<script") || strings.Contains(got, "alert") {
		t.Errorf("script not stripped: %s", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Error("safe content was removed")
	}
}

func TestHTML_StripIframe(t *testing.T) {
	input := `<p>Before</p><iframe src="evil.com"></iframe><p>After</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<iframe") {
		t.Errorf("iframe not stripped: %s", got)
	}
}

func TestHTML_StripObject(t *testing.T) {
	input := `<object data="evil.swf"></object>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<object") {
		t.Errorf("object not stripped: %s", got)
	}
}

func TestHTML_StripEmbed(t *testing.T) {
	input := `<embed src="evil.swf">`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<embed") {
		t.Errorf("embed not stripped: %s", got)
	}
}

func TestHTML_StripForm(t *testing.T) {
	input := `<form action="evil"><input type="text" value="x"></form>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<form") {
		t.Errorf("form not stripped: %s", got)
	}
}

// F-1: style, meta, base, link must be stripped
func TestHTML_StripStyleTag(t *testing.T) {
	input := `<style>body{background:red}</style><p>Content</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<style") {
		t.Errorf("style tag not stripped: %s", got)
	}
	if !strings.Contains(got, "Content") {
		t.Error("safe content was removed")
	}
}

func TestHTML_StripMetaTag(t *testing.T) {
	input := `<meta http-equiv="refresh" content="0;url=evil.com"><p>Content</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<meta") {
		t.Errorf("meta tag not stripped: %s", got)
	}
}

func TestHTML_StripBaseTag(t *testing.T) {
	input := `<base href="https://evil.com/"><p>Content</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<base") {
		t.Errorf("base tag not stripped: %s", got)
	}
}

func TestHTML_StripLinkTag(t *testing.T) {
	input := `<link rel="stylesheet" href="https://evil.com/steal.css"><p>Content</p>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<link") {
		t.Errorf("link tag not stripped: %s", got)
	}
}

func TestHTML_StripEventHandlers(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"onclick", `<div onclick="alert('xss')">Click</div>`},
		{"onerror", `<img onerror="alert('xss')" src="x">`},
		{"onload", `<body onload="alert('xss')">`},
		{"onmouseover", `<a onmouseover="evil()">link</a>`},
		{"mixed case", `<div ONCLICK="evil()">test</div>`},
		{"single quotes", `<div onclick='evil()'>test</div>`},
		{"no quotes", `<div onclick=evil()>test</div>`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(HTML([]byte(tc.input)))
			lower := strings.ToLower(got)
			if strings.Contains(lower, "onclick") ||
				strings.Contains(lower, "onerror") ||
				strings.Contains(lower, "onload") ||
				strings.Contains(lower, "onmouseover") {
				t.Errorf("event handler not stripped: %s", got)
			}
		})
	}
}

func TestHTML_PreservesSafeContent(t *testing.T) {
	input := `<h1>Title</h1><p>Hello <strong>world</strong></p><a href="https://example.com">link</a>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "Title") {
		t.Error("h1 content was stripped")
	}
	if !strings.Contains(got, "<strong>world</strong>") {
		t.Error("strong was stripped")
	}
	if !strings.Contains(got, `href="https://example.com"`) {
		t.Error("anchor href was stripped")
	}
}

func TestHTML_PreservesDetails(t *testing.T) {
	input := `<details><summary>Click</summary>Content</details>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<details>") {
		t.Error("details element was stripped")
	}
}

func TestHTML_PreservesImages(t *testing.T) {
	input := `<img src="image.png" alt="photo">`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<img") {
		t.Error("img was stripped")
	}
}

func TestHTML_PreservesHeadingIDs(t *testing.T) {
	input := `<h1 id="hello-world">Hello World</h1>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, `id="hello-world"`) {
		t.Errorf("heading id was stripped: %s", got)
	}
}

func TestHTML_PreservesCodeBlockClasses(t *testing.T) {
	input := `<pre tabindex="0" class="chroma"><code><span class="kn">package</span></code></pre>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, `class="chroma"`) {
		t.Errorf("pre class was stripped: %s", got)
	}
	if !strings.Contains(got, `class="kn"`) {
		t.Errorf("span class was stripped: %s", got)
	}
	if !strings.Contains(got, `tabindex="0"`) {
		t.Errorf("tabindex was stripped: %s", got)
	}
}

func TestHTML_PreservesFootnoteRoles(t *testing.T) {
	input := `<a href="#fn:1" class="footnote-ref" role="doc-noteref">1</a>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, `role="doc-noteref"`) {
		t.Errorf("footnote role was stripped: %s", got)
	}
}

func TestHTML_PreservesTaskListCheckboxes(t *testing.T) {
	input := `<li><input checked="" disabled="" type="checkbox"> Done</li>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<input") {
		t.Errorf("checkbox input was stripped: %s", got)
	}
	if !strings.Contains(got, "disabled") {
		t.Errorf("disabled attribute was stripped: %s", got)
	}
}

func TestHTML_PreservesDataAttributes(t *testing.T) {
	input := `<div class="cooked-code-block" data-language="go">code</div>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, `data-language="go"`) {
		t.Errorf("data-language was stripped: %s", got)
	}
}

func TestHTML_PreservesButtons(t *testing.T) {
	input := `<button class="cooked-copy-btn" data-state="idle">Copy</button>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<button") {
		t.Errorf("button was stripped: %s", got)
	}
}

func TestHTML_PreservesDefinitionLists(t *testing.T) {
	input := `<dl><dt>Term</dt><dd>Definition</dd></dl>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<dl>") || !strings.Contains(got, "<dt>") || !strings.Contains(got, "<dd>") {
		t.Errorf("definition list elements stripped: %s", got)
	}
}

func TestHTML_PreservesStrikethrough(t *testing.T) {
	input := `<del>deleted</del>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<del>") {
		t.Errorf("del was stripped: %s", got)
	}
}

// F-07: XSS hardening — javascript:, vbscript:, data:text/html URIs
func TestHTML_StripDangerousURIs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mustNot string
	}{
		{"javascript double-quoted", `<a href="javascript:alert(1)">click</a>`, "javascript:"},
		{"javascript single-quoted", `<a href='javascript:alert(1)'>click</a>`, "javascript:"},
		{"vbscript double-quoted", `<a href="vbscript:MsgBox(1)">click</a>`, "vbscript:"},
		{"data:text/html", `<a href="data:text/html,<script>alert(1)</script>">click</a>`, "data:text/html"},
		{"mixed case", `<a HREF="JavaScript:alert(1)">click</a>`, "javascript:"},
		{"src attribute", `<img src="javascript:alert(1)">`, "javascript:"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(HTML([]byte(tc.input)))
			lower := strings.ToLower(got)
			if strings.Contains(lower, strings.ToLower(tc.mustNot)) {
				t.Errorf("dangerous URI not stripped: %s", got)
			}
		})
	}
}

// F-6: Entity-encoded javascript: URIs must be caught
func TestHTML_StripEntityEncodedJavascriptURI(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"hex encoded", `<a href="&#x6A;avascript:alert(1)">click</a>`},
		{"decimal encoded", `<a href="&#106;avascript:alert(1)">click</a>`},
		{"mixed encoding", `<a href="&#x6A;&#x61;vascript:alert(1)">click</a>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(HTML([]byte(tc.input)))
			lower := strings.ToLower(got)
			if strings.Contains(lower, "javascript") {
				t.Errorf("entity-encoded javascript URI not stripped: %s", got)
			}
		})
	}
}

func TestHTML_PreservesSafeHrefs(t *testing.T) {
	input := `<a href="https://example.com">link</a>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "https://example.com") {
		t.Error("safe https href was stripped")
	}
}

func FuzzSanitizeHTML(f *testing.F) {
	seeds := []string{
		// XSS vectors
		`<script>alert('xss')</script>`,
		`<img onerror="alert(1)" src=x>`,
		`<div onclick="evil()">click</div>`,
		`<body onload="alert('xss')">`,
		`<iframe src="javascript:alert(1)"></iframe>`,
		`<object data="evil.swf"></object>`,
		`<embed src="evil.swf">`,
		`<form action="evil"><input type="text"></form>`,
		// F-1: newly blocked tags
		`<style>body{background:red}</style>`,
		`<meta http-equiv="refresh" content="0;url=evil">`,
		`<base href="https://evil.com/">`,
		`<link rel="stylesheet" href="evil.css">`,
		// Encoded payloads
		`<SCRIPT>alert(1)</SCRIPT>`,
		`<Script>alert(1)</Script>`,
		// Malformed tags
		`<script`,
		`<script >`,
		`<script/src="evil.js">`,
		`</script>`,
		`<iframe src="x"`,
		// Nested dangerous elements
		`<div><script><script>double</script></script></div>`,
		`<script><iframe>nested</iframe></script>`,
		`<form><iframe></iframe></form>`,
		// Event handlers in various forms
		`<a ONMOUSEOVER="evil()">link</a>`,
		`<div OnClick='evil()'>test</div>`,
		`<img onerror=evil() src=x>`,
		`<div onmouseenter="evil()" onmouseleave="evil()">both</div>`,
		// Safe content that must survive
		`<h1>Title</h1><p>Hello <strong>world</strong></p>`,
		`<a href="https://example.com">link</a>`,
		`<img src="photo.jpg" alt="image">`,
		`<details><summary>Click</summary>Content</details>`,
		`<pre><code>console.log('hi')</code></pre>`,
		// Unicode
		`<script>alert('日本語')</script>`,
		`<p>Ünïcödé content</p>`,
		// Edge cases
		"",
		"<>",
		"plain text with no tags",
		`<div style="color:red">styled</div>`,
		// Entity-encoded URI attacks (F-6)
		`<a href="&#x6A;avascript:alert(1)">click</a>`,
		`<a href="&#106;avascript:alert(1)">click</a>`,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got := HTML([]byte(input))

		// Determinism
		got2 := HTML([]byte(input))
		if string(got) != string(got2) {
			t.Errorf("non-deterministic output for input %q", input)
		}

		output := strings.ToLower(string(got))

		// Output must not contain dangerous tags
		dangerousTags := []string{"<script", "<iframe", "<object", "<embed",
			"<form", "<style", "<meta", "<base", "<link rel="}
		for _, tag := range dangerousTags {
			if strings.Contains(output, tag) {
				t.Errorf("output contains dangerous tag %q: %s", tag, output)
			}
		}

		// Output must not contain event handlers
		if strings.Contains(output, " on") {
			// More precise check: on followed by known handler names
			handlers := []string{"onclick", "onerror", "onload", "onmouseover",
				"onmouseenter", "onmouseleave", "onfocus", "onblur"}
			for _, h := range handlers {
				if strings.Contains(output, h) {
					t.Errorf("output contains event handler %q: %s", h, output)
				}
			}
		}

		// Idempotency: sanitizing twice should equal sanitizing once
		got3 := HTML(got)
		if string(got3) != string(got) {
			t.Errorf("not idempotent: single=%q double=%q", string(got), string(got3))
		}
	})
}
