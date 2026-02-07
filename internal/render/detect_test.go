package render

import "testing"

func TestDetectFile_Markdown(t *testing.T) {
	tests := []struct {
		path string
		want ContentType
	}{
		{"/README.md", TypeMarkdown},
		{"/docs/guide.markdown", TypeMarkdown},
		{"/file.mdown", TypeMarkdown},
		{"/file.mkd", TypeMarkdown},
		{"/FILE.MD", TypeMarkdown},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			info := DetectFile(tc.path)
			if info.ContentType != tc.want {
				t.Errorf("DetectFile(%q).ContentType = %q, want %q", tc.path, info.ContentType, tc.want)
			}
			if info.Label != "Markdown" {
				t.Errorf("DetectFile(%q).Label = %q, want Markdown", tc.path, info.Label)
			}
		})
	}
}

func TestDetectFile_MDX(t *testing.T) {
	info := DetectFile("/docs/intro.mdx")
	if info.ContentType != TypeMDX {
		t.Errorf("ContentType = %q, want %q", info.ContentType, TypeMDX)
	}
	if info.Label != "MDX" {
		t.Errorf("Label = %q, want MDX", info.Label)
	}
}

func TestDetectFile_Code(t *testing.T) {
	tests := []struct {
		path      string
		wantLang  string
		wantLabel string
	}{
		{"/main.py", "python", "Python"},
		{"/main.go", "go", "Go"},
		{"/index.js", "javascript", "JavaScript"},
		{"/index.ts", "typescript", "TypeScript"},
		{"/lib.rs", "rust", "Rust"},
		{"/hello.c", "c", "C"},
		{"/hello.h", "c", "C Header"},
		{"/hello.cpp", "cpp", "C++"},
		{"/hello.hpp", "cpp", "C++ Header"},
		{"/App.java", "java", "Java"},
		{"/script.rb", "ruby", "Ruby"},
		{"/script.lua", "lua", "Lua"},
		{"/script.pl", "perl", "Perl"},
		{"/install.sh", "bash", "Shell"},
		{"/install.bash", "bash", "Bash"},
		{"/config.zsh", "zsh", "Zsh"},
		{"/config.fish", "fish", "Fish"},
		{"/config.yaml", "yaml", "YAML"},
		{"/config.yml", "yaml", "YAML"},
		{"/data.json", "json", "JSON"},
		{"/config.toml", "toml", "TOML"},
		{"/data.xml", "xml", "XML"},
		{"/data.csv", "csv", "CSV"},
		{"/query.sql", "sql", "SQL"},
		{"/schema.graphql", "graphql", "GraphQL"},
		{"/main.tf", "hcl", "Terraform"},
		{"/config.hcl", "hcl", "HCL"},
		{"/changes.diff", "diff", "Diff"},
		{"/fix.patch", "diff", "Patch"},
		{"/app.dockerfile", "docker", "Dockerfile"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			info := DetectFile(tc.path)
			if info.ContentType != TypeCode {
				t.Errorf("DetectFile(%q).ContentType = %q, want %q", tc.path, info.ContentType, TypeCode)
			}
			if info.Language != tc.wantLang {
				t.Errorf("DetectFile(%q).Language = %q, want %q", tc.path, info.Language, tc.wantLang)
			}
			if info.Label != tc.wantLabel {
				t.Errorf("DetectFile(%q).Label = %q, want %q", tc.path, info.Label, tc.wantLabel)
			}
		})
	}
}

func TestDetectFile_SpecialNames(t *testing.T) {
	tests := []struct {
		path      string
		wantLang  string
		wantLabel string
	}{
		{"/Dockerfile", "docker", "Dockerfile"},
		{"/Makefile", "makefile", "Makefile"},
		{"/Jenkinsfile", "groovy", "Jenkinsfile"},
		{"/path/to/Dockerfile", "docker", "Dockerfile"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			info := DetectFile(tc.path)
			if info.ContentType != TypeCode {
				t.Errorf("DetectFile(%q).ContentType = %q, want %q", tc.path, info.ContentType, TypeCode)
			}
			if info.Language != tc.wantLang {
				t.Errorf("DetectFile(%q).Language = %q, want %q", tc.path, info.Language, tc.wantLang)
			}
			if info.Label != tc.wantLabel {
				t.Errorf("DetectFile(%q).Label = %q, want %q", tc.path, info.Label, tc.wantLabel)
			}
		})
	}
}

func TestDetectFile_Plaintext(t *testing.T) {
	tests := []string{
		"/readme.txt",
		"/notes.text",
		"/output.log",
		"/server.conf",
		"/settings.cfg",
		"/settings.ini",
		"/local.env",
	}

	for _, p := range tests {
		t.Run(p, func(t *testing.T) {
			info := DetectFile(p)
			if info.ContentType != TypePlaintext {
				t.Errorf("DetectFile(%q).ContentType = %q, want %q", p, info.ContentType, TypePlaintext)
			}
			if info.Label != "Plain Text" {
				t.Errorf("DetectFile(%q).Label = %q, want Plain Text", p, info.Label)
			}
		})
	}
}

func TestDetectFile_Unsupported(t *testing.T) {
	tests := []string{
		"/image.png",
		"/archive.tar.gz",
		"/binary.exe",
		"/",
		"",
	}

	for _, p := range tests {
		t.Run(p, func(t *testing.T) {
			info := DetectFile(p)
			if info.ContentType != TypeUnsupported {
				t.Errorf("DetectFile(%q).ContentType = %q, want %q", p, info.ContentType, TypeUnsupported)
			}
		})
	}
}

func TestDetectFile_CaseInsensitive(t *testing.T) {
	info := DetectFile("/README.MD")
	if info.ContentType != TypeMarkdown {
		t.Errorf("DetectFile(README.MD).ContentType = %q, want %q", info.ContentType, TypeMarkdown)
	}

	info = DetectFile("/SCRIPT.PY")
	if info.ContentType != TypeCode {
		t.Errorf("DetectFile(SCRIPT.PY).ContentType = %q, want %q", info.ContentType, TypeCode)
	}
}

func TestIsMarkdownLink(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"CONTRIBUTING.md", true},
		{"guide.markdown", true},
		{"docs.mdx", true},
		{"image.png", false},
		{"script.py", false},
		{"readme.txt", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			if got := IsMarkdownLink(tc.path); got != tc.want {
				t.Errorf("IsMarkdownLink(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
