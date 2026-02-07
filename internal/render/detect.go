package render

import (
	"path"
	"strings"
)

// ContentType represents the detected content type of a file.
type ContentType string

const (
	TypeMarkdown    ContentType = "markdown"
	TypeMDX         ContentType = "mdx"
	TypeCode        ContentType = "code"
	TypePlaintext   ContentType = "plaintext"
	TypeUnsupported ContentType = "unsupported"
)

// FileInfo holds detected information about a file.
type FileInfo struct {
	ContentType ContentType
	Language    string // e.g. "python", "go", "yaml" — empty for markdown/plaintext/unsupported
	Label       string // human-readable label e.g. "Markdown", "Python", "YAML"
}

// markdownExts maps markdown file extensions.
var markdownExts = map[string]bool{
	".md":       true,
	".markdown": true,
	".mdown":    true,
	".mkd":      true,
}

// codeExts maps file extensions to (language, label).
var codeExts = map[string][2]string{
	".py":         {"python", "Python"},
	".go":         {"go", "Go"},
	".js":         {"javascript", "JavaScript"},
	".ts":         {"typescript", "TypeScript"},
	".rs":         {"rust", "Rust"},
	".c":          {"c", "C"},
	".h":          {"c", "C Header"},
	".cpp":        {"cpp", "C++"},
	".hpp":        {"cpp", "C++ Header"},
	".java":       {"java", "Java"},
	".rb":         {"ruby", "Ruby"},
	".lua":        {"lua", "Lua"},
	".pl":         {"perl", "Perl"},
	".sh":         {"bash", "Shell"},
	".bash":       {"bash", "Bash"},
	".zsh":        {"zsh", "Zsh"},
	".fish":       {"fish", "Fish"},
	".yaml":       {"yaml", "YAML"},
	".yml":        {"yaml", "YAML"},
	".json":       {"json", "JSON"},
	".toml":       {"toml", "TOML"},
	".xml":        {"xml", "XML"},
	".csv":        {"csv", "CSV"},
	".sql":        {"sql", "SQL"},
	".graphql":    {"graphql", "GraphQL"},
	".tf":         {"hcl", "Terraform"},
	".hcl":        {"hcl", "HCL"},
	".diff":       {"diff", "Diff"},
	".patch":      {"diff", "Patch"},
	".dockerfile": {"docker", "Dockerfile"},
}

// specialNames maps exact filenames (case-insensitive) to (language, label).
var specialNames = map[string][2]string{
	"dockerfile":  {"docker", "Dockerfile"},
	"makefile":    {"makefile", "Makefile"},
	"jenkinsfile": {"groovy", "Jenkinsfile"},
}

// plaintextExts maps plaintext file extensions.
var plaintextExts = map[string]bool{
	".txt":  true,
	".text": true,
	".log":  true,
	".conf": true,
	".cfg":  true,
	".ini":  true,
	".env":  true,
}

// DetectFile determines the content type and language of a file from its URL path.
func DetectFile(urlPath string) FileInfo {
	// Extract filename from path
	filename := path.Base(urlPath)
	if filename == "." || filename == "/" {
		return FileInfo{ContentType: TypeUnsupported, Label: "Unknown"}
	}

	// Check for special filenames first (case-insensitive)
	lower := strings.ToLower(filename)
	if info, ok := specialNames[lower]; ok {
		return FileInfo{ContentType: TypeCode, Language: info[0], Label: info[1]}
	}

	// Get extension (lowercase)
	ext := strings.ToLower(path.Ext(filename))

	// MDX is a special case — markdown variant
	if ext == ".mdx" {
		return FileInfo{ContentType: TypeMDX, Label: "MDX"}
	}

	// Check markdown extensions
	if markdownExts[ext] {
		return FileInfo{ContentType: TypeMarkdown, Label: "Markdown"}
	}

	// Check code extensions
	if info, ok := codeExts[ext]; ok {
		return FileInfo{ContentType: TypeCode, Language: info[0], Label: info[1]}
	}

	// Check plaintext extensions
	if plaintextExts[ext] {
		return FileInfo{ContentType: TypePlaintext, Label: "Plain Text"}
	}

	return FileInfo{ContentType: TypeUnsupported, Label: "Unknown"}
}

// IsMarkdownLink returns true if the URL path points to a markdown-like file
// (used for relative URL rewriting decisions).
func IsMarkdownLink(urlPath string) bool {
	ext := strings.ToLower(path.Ext(urlPath))
	return markdownExts[ext] || ext == ".mdx"
}
