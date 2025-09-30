package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opskraken/codeecho-cli/config"
	"github.com/opskraken/codeecho-cli/output"
	"github.com/opskraken/codeecho-cli/scanner"
	"github.com/opskraken/codeecho-cli/utils"
	"github.com/spf13/cobra"
)

var (
	// Output format flags
	outputFormat         string
	outputFile           string
	includeSummary       bool
	includeDirectoryTree bool
	showLineNumbers      bool
	outputParsableFormat bool

	// File processing flags
	compressCode     bool
	removeComments   bool
	removeEmptyLines bool

	// File filtering flags
	excludeDirs    []string
	includeExts    []string
	includeContent bool
	excludeContent bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan repository and generate AI-ready context",
	Long: `Scan a repository and generate structured output for AI consumption.
Similar to Repomix, this command creates a single file containing your entire
codebase in a format optimized for AI tools.

Output Formats:
  xml        - Structured XML format (recommended for AI)
  json       - JSON format for programmatic use
  markdown   - Human-readable markdown format

Examples:
  codeecho scan .                              # Basic XML scan
  codeecho scan . --format json               # JSON output
  codeecho scan . --remove-comments           # Strip comments
  codeecho scan . --compress-code             # Minify code
  codeecho scan . --no-summary                # Skip file summary
  codeecho scan . --output packed-repo.xml    # Save to file`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Output format flags
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "xml", "Output format: xml, json, markdown")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: auto-generated)")
	scanCmd.Flags().BoolVar(&includeSummary, "include-summary", true, "Include file summary section")
	scanCmd.Flags().BoolVar(&includeDirectoryTree, "include-tree", true, "Include directory structure")
	scanCmd.Flags().BoolVar(&showLineNumbers, "line-numbers", false, "Show line numbers in code blocks")
	scanCmd.Flags().BoolVar(&outputParsableFormat, "parsable", true, "Use parsable format tags")

	// File processing flags
	scanCmd.Flags().BoolVar(&compressCode, "compress-code", false, "Remove unnecessary whitespace from code")
	scanCmd.Flags().BoolVar(&removeComments, "remove-comments", false, "Strip comments from source files")
	scanCmd.Flags().BoolVar(&removeEmptyLines, "remove-empty-lines", false, "Remove empty lines from files")

	// File filtering flags
	scanCmd.Flags().BoolVar(&includeContent, "content", true, "Include file contents")
	scanCmd.Flags().BoolVar(&excludeContent, "no-content", false, "Exclude file contents (structure only)")
	scanCmd.Flags().StringSliceVar(&excludeDirs, "exclude-dirs",
		[]string{".git", "node_modules", "vendor", ".vscode", ".idea", "target", "build", "dist"},
		"Directories to exclude")
	scanCmd.Flags().StringSliceVar(&includeExts, "include-exts",
		[]string{".go", ".js", ".ts", ".jsx", ".tsx", ".json", ".md", ".html", ".css", ".py", ".java", ".cpp", ".c", ".h", ".rs", ".rb", ".php", ".yml", ".yaml", ".toml", ".xml"},
		"File extensions to include")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Validate path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}

	// Get absolute path for cleaner output
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Scanning repository at %s...\n", absPath)

	if excludeContent {
		includeContent = false
	}

	if compressCode || removeComments || removeEmptyLines {
		fmt.Println("File processing enabled:")
		if compressCode {
			fmt.Println("  - Code compression")
		}
		if removeComments {
			fmt.Println("  - Comment removal")
		}
		if removeEmptyLines {
			fmt.Println("  - Empty line removal")
		}
	}

	opts := scanner.ScanOptions{
		IncludeSummary:       includeSummary,
		IncludeDirectoryTree: includeDirectoryTree,
		ShowLineNumbers:      showLineNumbers,
		OutputParsableFormat: outputParsableFormat,
		CompressCode:         compressCode,
		RemoveComments:       removeComments,
		RemoveEmptyLines:     removeEmptyLines,
		ExcludeDirs:          excludeDirs,
		IncludeExts:          includeExts,
		IncludeContent:       includeContent,
	}
	// Perform the scan
	result, err := scanner.ScanRepository(absPath, opts)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	outputOpts := config.OutputOptions{
		IncludeSummary:       includeSummary,
		IncludeDirectoryTree: includeDirectoryTree,
		ShowLineNumbers:      showLineNumbers,
		IncludeContent:       includeContent,
		RemoveComments:       removeComments,
		RemoveEmptyLines:     removeEmptyLines,
		CompressCode:         compressCode,
	}
	// Generate output based on format
	var outputContent string
	switch strings.ToLower(outputFormat) {
	case "xml":
		outputContent, err = output.GenerateXMLOutput(result, outputOpts)
	case "json":
		outputContent, err = output.GenerateJSONOutput(result)
	case "markdown", "md":
		outputContent, err = output.GenerateMarkdownOutput(result, outputOpts)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: xml, json, markdown)", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Write output
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(outputContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", outputFile)
	} else {
		// Auto-generate filename if not specified
		autoFile := utils.GenerateAutoFilename(result.RepoPath, outputFormat, outputOpts)
		err = os.WriteFile(autoFile, []byte(outputContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", autoFile)
	}

	// Enhanced scan summary
	fmt.Printf("\nScan Summary:\n")
	fmt.Printf("  Files processed: %d\n", result.TotalFiles)
	fmt.Printf("  Total size: %s\n", utils.FormatBytes(result.TotalSize))

	// Show file type breakdown
	fileTypes := make(map[string]int)
	textFiles := 0
	binaryFiles := 0

	for _, file := range result.Files {
		if file.IsText {
			textFiles++
		} else {
			binaryFiles++
		}

		if file.Language != "" {
			fileTypes[file.Language]++
		} else if file.Extension != "" {
			fileTypes[file.Extension]++
		} else {
			fileTypes["no extension"]++
		}
	}

	fmt.Printf("  Text files: %d, Binary files: %d\n", textFiles, binaryFiles)

	// Show top file types
	if len(fileTypes) > 0 {
		fmt.Printf("  Top file types: ")
		type kv struct {
			Key   string
			Value int
		}
		var sortedTypes []kv
		for k, v := range fileTypes {
			sortedTypes = append(sortedTypes, kv{k, v})
		}
		sort.Slice(sortedTypes, func(i, j int) bool {
			return sortedTypes[i].Value > sortedTypes[j].Value
		})

		for i, kv := range sortedTypes {
			if i > 2 { // Show top 3
				break
			}
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s (%d)", kv.Key, kv.Value)
		}
		fmt.Printf("\n")
	}

	return nil
}
