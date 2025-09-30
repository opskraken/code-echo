package scanner

import (
	"encoding/json"
	"regexp"
	"strings"
)

func processFileContent(content, language string, opts ScanOptions) string {
	processed := content

	if opts.RemoveComments {
		processed = stripComments(processed, language)
	}
	if opts.RemoveEmptyLines {
		processed = stripEmptyLines(processed)
	}
	if opts.CompressCode {
		processed = compressWhitespace(processed, language)
	}

	return processed
}

// stripComments removes comments based on file language
func stripComments(content, language string) string {
	switch language {
	case "go", "javascript", "typescript", "java", "cpp", "c", "rust", "php":
		// Remove single-line comments //
		re1 := regexp.MustCompile(`//.*$`)
		content = re1.ReplaceAllString(content, "")

		// Remove multi-line comments /* */
		re2 := regexp.MustCompile(`/\*[\s\S]*?\*/`)
		content = re2.ReplaceAllString(content, "")

	case "python", "ruby":
		// Remove # comments
		re := regexp.MustCompile(`#.*$`)
		content = re.ReplaceAllString(content, "")

	case "html", "xml":
		// Remove HTML/XML comments <!-- -->
		re := regexp.MustCompile(`<!--[\s\S]*?-->`)
		content = re.ReplaceAllString(content, "")

	case "css":
		// Remove CSS comments /* */
		re := regexp.MustCompile(`/\*[\s\S]*?\*/`)
		content = re.ReplaceAllString(content, "")
	}

	return content
}

// stripEmptyLines removes empty lines from content
func stripEmptyLines(content string) string {
	lines := strings.Split(content, "\n")
	var nonEmptyLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}

// compressWhitespace removes unnecessary whitespace
func compressWhitespace(content, language string) string {
	switch language {
	case "json":
		// For JSON, we can actually minify it properly
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(content), &jsonObj); err == nil {
			if minified, err := json.Marshal(jsonObj); err == nil {
				return string(minified)
			}
		}
	case "javascript", "css":
		// Basic whitespace compression for JS/CSS
		// Remove extra spaces and tabs (but preserve single spaces)
		re := regexp.MustCompile(`[ \t]+`)
		content = re.ReplaceAllString(content, " ")
	}

	// Generic whitespace compression
	// Remove trailing whitespace from each line
	re := regexp.MustCompile(`[ \t]+$`)
	content = re.ReplaceAllString(content, "")

	return content
}
