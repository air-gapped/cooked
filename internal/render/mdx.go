package render

import (
	"regexp"
	"strings"
)

// PreprocessMDX transforms MDX content into standard markdown that goldmark can render.
// It strips JSX imports/exports, handles JSX component tags, and preserves frontmatter.
func PreprocessMDX(source []byte) []byte {
	lines := strings.Split(string(source), "\n")
	var result []string
	inFrontmatter := false
	frontmatterDone := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle frontmatter
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			result = append(result, line)
			continue
		}
		if inFrontmatter {
			result = append(result, line)
			if trimmed == "---" {
				inFrontmatter = false
				frontmatterDone = true
			}
			continue
		}
		_ = frontmatterDone

		// Strip import statements
		if importRe.MatchString(trimmed) {
			continue
		}

		// Strip export statements
		if exportRe.MatchString(trimmed) {
			continue
		}

		result = append(result, line)
	}

	content := strings.Join(result, "\n")

	// Strip self-closing JSX tags: <Component />
	content = selfClosingJSXRe.ReplaceAllString(content, "")

	// Handle JSX container tags: extract content, use label/title as heading
	content = processContainerTags(content)

	// Strip JSX expressions: {props.foo}, {<Component />}
	content = jsxExprRe.ReplaceAllString(content, "")

	// Final pass: strip import/export lines exposed by JSX removal
	lines2 := strings.Split(content, "\n")
	var cleaned []string
	for _, line := range lines2 {
		trimmed := strings.TrimSpace(line)
		if importRe.MatchString(trimmed) || exportRe.MatchString(trimmed) {
			continue
		}
		cleaned = append(cleaned, line)
	}
	content = strings.Join(cleaned, "\n")

	return []byte(content)
}

// importRe matches import statements.
var importRe = regexp.MustCompile(`^import\s+(?:(?:\{[^}]*\}|\w+|\*\s+as\s+\w+)\s+from\s+)?['"][^'"]+['"];?\s*$`)

// exportRe matches export statements.
var exportRe = regexp.MustCompile(`^export\s+(?:default\s+|const\s+|let\s+|function\s+|class\s+)`)

// selfClosingJSXRe matches self-closing JSX tags like <Component />.
var selfClosingJSXRe = regexp.MustCompile(`<[A-Z]\w*\s*[^>]*/\s*>`)

// jsxExprRe matches JSX expressions like {props.foo}.
var jsxExprRe = regexp.MustCompile(`\{[^{}]*\}`)

// containerTagRe matches opening JSX container tags with optional attributes.
var containerTagOpenRe = regexp.MustCompile(`<([A-Z]\w*)\s*([^>]*)>`)
var containerTagCloseRe = regexp.MustCompile(`</[A-Z]\w*>`)

// labelAttrRe extracts label, title, or value attributes.
var labelAttrRe = regexp.MustCompile(`(?:label|title|value)\s*=\s*(?:"([^"]*)"|'([^']*)'|{[^}]*})`)

// processContainerTags strips JSX container tags, keeps inner content.
// If a tag has a label/title/value attribute, render it as a bold heading.
func processContainerTags(content string) string {
	// Extract labels from opening tags before stripping
	content = containerTagOpenRe.ReplaceAllStringFunc(content, func(match string) string {
		labelMatch := labelAttrRe.FindStringSubmatch(match)
		if labelMatch != nil {
			label := labelMatch[1]
			if label == "" {
				label = labelMatch[2]
			}
			if label != "" {
				return "\n**" + label + "**\n"
			}
		}
		return ""
	})

	// Remove closing tags
	content = containerTagCloseRe.ReplaceAllString(content, "")

	return content
}
