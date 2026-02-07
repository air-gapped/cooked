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
