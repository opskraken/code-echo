package scanner

import (
	"path/filepath"
	"strings"
)

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "jsx",
		".tsx":  "tsx",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".h":    "c",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".css":  "css",
		".html": "html",
		".json": "json",
		".md":   "markdown",
		".yml":  "yaml",
		".yaml": "yaml",
		".toml": "toml",
		".xml":  "xml",
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return ""
}

func isTextFile(path, extension string) bool {
	// Known text extensions
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".rst": true, ".asciidoc": true,
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true,
		".cs": true, ".php": true, ".rb": true, ".rs": true, ".swift": true, ".kt": true,
		".html": true, ".htm": true, ".xml": true, ".xhtml": true,
		".css": true, ".scss": true, ".sass": true, ".less": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true, ".cfg": true, ".conf": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".ps1": true, ".bat": true, ".cmd": true,
		".sql": true, ".graphql": true, ".gql": true,
		".dockerfile": true, ".gitignore": true, ".gitattributes": true,
		".makefile": true, ".cmake": true,
		".r": true, ".rmd": true, ".m": true, ".scala": true, ".clj": true, ".hs": true,
		".vim": true, ".lua": true, ".pl": true, ".tcl": true,
		".tex": true, ".bib": true, ".cls": true, ".sty": true,
		".csv": true, ".tsv": true, ".log": true,
	}

	ext := strings.ToLower(extension)
	if textExtensions[ext] {
		return true
	}

	// Files without extensions but with known names
	fileName := strings.ToLower(filepath.Base(path))
	textFiles := map[string]bool{
		"readme": true, "license": true, "changelog": true, "contributing": true,
		"authors": true, "contributors": true, "copying": true, "install": true,
		"news": true, "thanks": true, "todo": true, "version": true,
		"makefile": true, "dockerfile": true, "jenkinsfile": true,
		"gemfile": true, "rakefile": true, "guardfile": true, "procfile": true,
		".gitignore": true, ".gitattributes": true, ".dockerignore": true,
		".eslintrc": true, ".prettierrc": true, ".babelrc": true,
	}

	return textFiles[fileName]
}