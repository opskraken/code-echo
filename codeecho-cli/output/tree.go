package output

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/opskraken/codeecho-cli/scanner"
)

func GenerateDirectoryTree(files []scanner.FileInfo) string {
	if len(files) == 0 {
		return ""
	}

	// Determine project root name
	projectRoot := "project"
	if len(files) > 0 && files[0].Path != "" {
		dir := filepath.Dir(files[0].Path)
		if dir != "." && dir != "/" {
			projectRoot = filepath.Base(dir)
		}
	}

	// Build directory structure
	var result strings.Builder
	result.WriteString(projectRoot + "/\n")

	// Track processed paths to avoid duplicates
	processed := make(map[string]bool)

	for _, file := range files {
		parts := strings.Split(file.RelativePath, string(filepath.Separator))

		// Build each level of the path
		for i := range parts {
			pathSoFar := strings.Join(parts[:i+1], "/")

			if !processed[pathSoFar] {
				processed[pathSoFar] = true
				indent := strings.Repeat("  ", i+1)

				if i == len(parts)-1 {
					// It's a file
					result.WriteString(fmt.Sprintf("%s%s\n", indent, parts[i]))
				} else {
					// It's a directory
					result.WriteString(fmt.Sprintf("%s%s/\n", indent, parts[i]))
				}
			}
		}
	}

	return result.String()
}
