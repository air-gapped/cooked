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
	if !strings.Contains(got, "<p>Hello</p>") {
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
	input := `<form action="evil"><input type="text"></form>`
	got := string(HTML([]byte(input)))
	if strings.Contains(got, "<form") || strings.Contains(got, "<input") {
		t.Errorf("form/input not stripped: %s", got)
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
			if strings.Contains(strings.ToLower(got), "onclick") ||
				strings.Contains(strings.ToLower(got), "onerror") ||
				strings.Contains(strings.ToLower(got), "onload") ||
				strings.Contains(strings.ToLower(got), "onmouseover") {
				t.Errorf("event handler not stripped: %s", got)
			}
		})
	}
}

func TestHTML_PreservesSafeContent(t *testing.T) {
	input := `<h1>Title</h1><p>Hello <strong>world</strong></p><a href="https://example.com">link</a>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "<h1>Title</h1>") {
		t.Error("h1 was stripped")
	}
	if !strings.Contains(got, "<strong>world</strong>") {
		t.Error("strong was stripped")
	}
	if !strings.Contains(got, `<a href="https://example.com">link</a>`) {
		t.Error("anchor was stripped")
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
		// Encoded payloads
		`<scr` + `ipt>alert(1)</sc` + `ript>`,
		`<SCRIPT>alert(1)</SCRIPT>`,
		`<Script>alert(1)</Script>`,
		// Malformed tags
		`<script`,
		`<script >`,
		`<script/src="evil.js">`,
		`</script>`,
		`<iframe src="x"`,
		`<scr ipt>alert(1)</scr ipt>`,
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
		string(make([]byte, 4096)),
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

		output := string(got)

		// Output must not match the sanitizer's own dangerous patterns.
		// Block patterns: <tag ...>...</tag>
		for _, re := range blockPatterns {
			if re.MatchString(output) {
				t.Errorf("output matches block pattern: %s", output)
			}
		}
		// Self-closing patterns: <tag ...> or <tag .../>
		for _, re := range selfClosingPatterns {
			if re.MatchString(output) {
				t.Errorf("output matches self-closing pattern: %s", output)
			}
		}

		// Output must not contain event handlers
		if eventHandlerRe.MatchString(output) {
			t.Errorf("output contains event handler: %s", output)
		}

		// Idempotency: sanitizing twice should equal sanitizing once
		got3 := HTML(got)
		if string(got3) != output {
			t.Errorf("not idempotent: single=%q double=%q", output, string(got3))
		}
	})
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

func TestHTML_PreservesSafeHrefs(t *testing.T) {
	input := `<a href="https://example.com">link</a><a href="data:image/png;base64,abc">img</a>`
	got := string(HTML([]byte(input)))
	if !strings.Contains(got, "https://example.com") {
		t.Error("safe https href was stripped")
	}
	if !strings.Contains(got, "data:image/png") {
		t.Error("safe data:image URI was stripped")
	}
}

func TestContainsDangerousContent(t *testing.T) {
	if !ContainsDangerousContent(`<script>alert('hi')</script>`) {
		t.Error("should detect script")
	}
	if !ContainsDangerousContent(`<iframe src="x"></iframe>`) {
		t.Error("should detect iframe")
	}
	if !ContainsDangerousContent(`<div onclick="evil()">`) {
		t.Error("should detect event handler")
	}
	if ContainsDangerousContent(`<p>safe content</p>`) {
		t.Error("should not flag safe content")
	}
}
