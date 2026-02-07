package sanitize

import (
	"regexp"
	"strings"
)

// dangerousTags are HTML tags that must be stripped from upstream content.
var dangerousTags = []string{"script", "iframe", "object", "embed", "form", "input"}

// blockPatterns matches full element blocks (opening tag + content + closing tag).
// These are applied first to remove both tags and their content.
var blockPatterns []*regexp.Regexp

// selfClosingPatterns matches void/self-closing elements that have no closing tag.
var selfClosingPatterns []*regexp.Regexp

// eventHandlerRe matches on* event handler attributes.
var eventHandlerRe = regexp.MustCompile(`(?i)\s+on\w+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]*)`)

func init() {
	for _, tag := range dangerousTags {
		// Full block: <tag ...>content</tag>
		block := regexp.MustCompile(`(?is)<` + tag + `\b[^>]*>.*?</` + tag + `\s*>`)
		blockPatterns = append(blockPatterns, block)

		// Self-closing or orphaned opening tags: <tag ...> or <tag .../>
		selfClosing := regexp.MustCompile(`(?is)<` + tag + `\b[^>]*/?>`)
		selfClosingPatterns = append(selfClosingPatterns, selfClosing)
	}
}

// HTML strips dangerous elements and event handlers from HTML content.
// This processes the HTML after goldmark rendering but BEFORE cooked's own
// scripts are injected (so cooked's scripts are never stripped).
func HTML(input []byte) []byte {
	s := string(input)

	// First pass: remove full blocks (tag + content + closing tag)
	for _, re := range blockPatterns {
		s = re.ReplaceAllString(s, "")
	}

	// Second pass: remove any remaining self-closing or orphaned tags
	for _, re := range selfClosingPatterns {
		s = re.ReplaceAllString(s, "")
	}

	// Remove on* event handler attributes
	s = eventHandlerRe.ReplaceAllString(s, "")

	return []byte(s)
}

// ContainsDangerousContent checks if HTML has any dangerous elements.
// Useful for testing.
func ContainsDangerousContent(html string) bool {
	lower := strings.ToLower(html)
	for _, tag := range dangerousTags {
		if strings.Contains(lower, "<"+tag) {
			return true
		}
	}
	if eventHandlerRe.MatchString(html) {
		return true
	}
	return false
}
