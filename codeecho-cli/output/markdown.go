package output

import (
	"fmt"
	"strings"

	"github.com/opskraken/codeecho-cli/config"
	"github.com/opskraken/codeecho-cli/scanner"
	"github.com/opskraken/codeecho-cli/utils"
)

func GenerateMarkdownOutput(result *scanner.ScanResult, opts config.OutputOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("# CodeEcho Repository Scan\n\n")
	builder.WriteString(fmt.Sprintf("**Repository:** %s  \n", result.RepoPath))
	builder.WriteString(fmt.Sprintf("**Scan Time:** %s  \n", result.ScanTime))
	builder.WriteString(fmt.Sprintf("**Total Files:** %d  \n", result.TotalFiles))
	builder.WriteString(fmt.Sprintf("**Total Size:** %s  \n\n", utils.FormatBytes(result.TotalSize)))

	if opts.IncludeDirectoryTree {
		builder.WriteString("## Directory Structure\n\n")
		builder.WriteString("```\n")
		builder.WriteString(GenerateDirectoryTree(result.Files))
		builder.WriteString("```\n\n")
	}

	builder.WriteString("## Files\n\n")
	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf("### %s\n\n", file.RelativePath))

		// Enhanced metadata display
		builder.WriteString(fmt.Sprintf("**Size:** %s", file.SizeFormatted))
		if file.Language != "" {
			builder.WriteString(fmt.Sprintf(" | **Language:** %s", file.Language))
		}
		if file.LineCount > 0 {
			builder.WriteString(fmt.Sprintf(" | **Lines:** %d", file.LineCount))
		}
		if file.Extension != "" {
			builder.WriteString(fmt.Sprintf(" | **Extension:** %s", file.Extension))
		}
		builder.WriteString(fmt.Sprintf(" | **Modified:** %s", file.ModTimeFormatted))
		builder.WriteString(fmt.Sprintf(" | **Text File:** %t", file.IsText))
		builder.WriteString("\n\n")

		// Content display
		if opts.IncludeContent && file.Content != "" && file.IsText {
			builder.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", strings.ToLower(file.Language), file.Content))
		} else if !file.IsText {
			builder.WriteString("*Binary file - content not displayed*\n\n")
		} else {
			builder.WriteString("*Content not included*\n\n")
		}

		builder.WriteString("---\n\n")
	}

	return builder.String(), nil
}
