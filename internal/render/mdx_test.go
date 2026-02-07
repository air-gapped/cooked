package render

import (
	"strings"
	"testing"
)

func TestPreprocessMDX_StripImports(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"default import", `import Tabs from '@theme/Tabs';`},
		{"named import", `import { TabItem } from '@docusaurus/theme';`},
		{"star import", `import * as utils from './utils';`},
		{"side effect import", `import './styles.css';`},
		{"double quotes", `import Foo from "./Foo"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := string(PreprocessMDX([]byte(tc.input + "\n\nContent\n")))
			if strings.Contains(got, "import") {
				t.Errorf("import not stripped: %s", got)
			}
			if !strings.Contains(got, "Content") {
				t.Error("content was lost")
			}
		})
	}
}

func TestPreprocessMDX_StripExports(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`export default function Layout() {}`},
		{`export const metadata = { title: 'Hello' };`},
		{`export let foo = 'bar';`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := string(PreprocessMDX([]byte(tc.input + "\n\nContent\n")))
			if strings.Contains(got, "export") {
				t.Errorf("export not stripped: %s", got)
			}
		})
	}
}

func TestPreprocessMDX_PreserveFrontmatter(t *testing.T) {
	input := "---\ntitle: My Doc\nauthor: Test\n---\n\n# Hello\n"
	got := string(PreprocessMDX([]byte(input)))

	if !strings.Contains(got, "---") {
		t.Error("frontmatter was stripped")
	}
	if !strings.Contains(got, "title: My Doc") {
		t.Error("frontmatter content was lost")
	}
	if !strings.Contains(got, "# Hello") {
		t.Error("markdown content was lost")
	}
}

func TestPreprocessMDX_SelfClosingJSX(t *testing.T) {
	input := "Before\n\n<CustomComponent />\n\nAfter\n"
	got := string(PreprocessMDX([]byte(input)))

	if strings.Contains(got, "CustomComponent") {
		t.Errorf("self-closing JSX not removed: %s", got)
	}
	if !strings.Contains(got, "Before") || !strings.Contains(got, "After") {
		t.Error("surrounding content was lost")
	}
}

func TestPreprocessMDX_ContainerJSX_WithLabel(t *testing.T) {
	input := `<TabItem label="npm">

` + "```bash\nnpm install foo\n```" + `

</TabItem>`

	got := string(PreprocessMDX([]byte(input)))

	// Label should become bold heading
	if !strings.Contains(got, "**npm**") {
		t.Errorf("label not rendered as bold heading: %s", got)
	}

	// Content should be preserved
	if !strings.Contains(got, "npm install foo") {
		t.Error("code block content was lost")
	}

	// Tags should be gone
	if strings.Contains(got, "<TabItem") || strings.Contains(got, "</TabItem>") {
		t.Error("JSX tags not stripped")
	}
}

func TestPreprocessMDX_ContainerJSX_NoLabel(t *testing.T) {
	input := "<Tabs>\nContent here\n</Tabs>\n"
	got := string(PreprocessMDX([]byte(input)))

	if strings.Contains(got, "<Tabs") || strings.Contains(got, "</Tabs>") {
		t.Error("JSX tags not stripped")
	}
	if !strings.Contains(got, "Content here") {
		t.Error("inner content was lost")
	}
}

func TestPreprocessMDX_JSXExpressions(t *testing.T) {
	input := "Hello {props.name}, the count is {count}.\n"
	got := string(PreprocessMDX([]byte(input)))

	if strings.Contains(got, "{props.name}") || strings.Contains(got, "{count}") {
		t.Errorf("JSX expressions not stripped: %s", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Error("text content was lost")
	}
}

func FuzzPreprocessMDX(f *testing.F) {
	seeds := []string{
		// Real MDX content
		"---\ntitle: Hello\n---\n\nimport Tabs from '@theme/Tabs';\n\n# Hello\n",
		"<TabItem label=\"npm\">\n\n```bash\nnpm install\n```\n\n</TabItem>",
		"<Tabs>\n<TabItem label=\"npm\">content</TabItem>\n</Tabs>",
		"export default function Layout() {}",
		"import { useState } from 'react';",
		// Self-closing JSX
		"<CustomComponent />",
		"<Widget prop=\"value\" />",
		// JSX expressions
		"Hello {props.name}, count is {count}.",
		// Nested braces
		"Text {outer {inner}} end",
		"{{{deeply nested}}}",
		// Malformed JSX
		"<Unclosed",
		"<Partial label=\"no-close>content",
		"</Orphan>",
		"<Not A Tag>",
		"< Component />",
		// Attribute injection
		`<Tab label="x" onclick="evil()">content</Tab>`,
		`<Tab title='"><script>alert(1)</script>'>hi</Tab>`,
		// Unicode
		"<Comp label=\"日本語\" />\n",
		"import 变量 from '模块';\n",
		"---\ntitle: Ünïcödé\n---\n\n# Héllo\n",
		// Empty and minimal
		"",
		"\n",
		"---\n---\n",
		"plain markdown with no JSX at all",
		// Large nested structure
		"<Outer><Middle><Inner>deep</Inner></Middle></Outer>",
		// Mixed content
		"# Title\n\nimport X from 'x';\n\n<A />\n\nText {expr}\n\nexport const y = 1;\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got := PreprocessMDX([]byte(input))

		// Determinism: same input must produce same output
		got2 := PreprocessMDX([]byte(input))
		if string(got) != string(got2) {
			t.Errorf("non-deterministic output for input %q", input)
		}

		// Lines present in the input that match importRe must be stripped
		// (unless inside frontmatter). We check exact line presence in
		// output to avoid substring false positives.
		outputLines := make(map[string]bool)
		for _, ol := range strings.Split(string(got), "\n") {
			outputLines[ol] = true
		}
		inFM := false
		for i, line := range strings.Split(input, "\n") {
			trimmed := strings.TrimSpace(line)
			if i == 0 && trimmed == "---" {
				inFM = true
				continue
			}
			if inFM {
				if trimmed == "---" {
					inFM = false
				}
				continue
			}
			if importRe.MatchString(trimmed) && outputLines[line] {
				t.Errorf("import line not stripped from output: %q", line)
			}
			if exportRe.MatchString(trimmed) && outputLines[line] {
				t.Errorf("export line not stripped from output: %q", line)
			}
		}
	})
}

func TestPreprocessMDX_FullDocument(t *testing.T) {
	input := `---
title: Getting Started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Getting Started

Install the package:

<Tabs>
<TabItem label="npm">

` + "```bash\nnpm install foo\n```" + `

</TabItem>
<TabItem label="yarn">

` + "```bash\nyarn add foo\n```" + `

</TabItem>
</Tabs>

<CustomWidget />

That's it! You're ready to go.
`

	got := string(PreprocessMDX([]byte(input)))

	// Should preserve
	if !strings.Contains(got, "# Getting Started") {
		t.Error("heading lost")
	}
	if !strings.Contains(got, "npm install foo") {
		t.Error("npm code block lost")
	}
	if !strings.Contains(got, "yarn add foo") {
		t.Error("yarn code block lost")
	}
	if !strings.Contains(got, "ready to go") {
		t.Error("trailing text lost")
	}

	// Should strip
	if strings.Contains(got, "import Tabs") {
		t.Error("import not stripped")
	}
	if strings.Contains(got, "CustomWidget") {
		t.Error("self-closing JSX not stripped")
	}
}
