package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

	// Determine output file
	var outputFilePath string
	if outputFile != "" {
		outputFilePath = outputFile
	} else {
		// Generate auto filename
		outputOpts := config.OutputOptions{
			IncludeSummary:       includeSummary,
			IncludeDirectoryTree: includeDirectoryTree,
			ShowLineNumbers:      showLineNumbers,
			IncludeContent:       includeContent,
			RemoveComments:       removeComments,
			RemoveEmptyLines:     removeEmptyLines,
			CompressCode:         compressCode,
		}
		outputFilePath = utils.GenerateAutoFilename(absPath, outputFormat, outputOpts)
	}

	// Create output file
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create output options
	outputOpts := config.OutputOptions{
		IncludeSummary:       includeSummary,
		IncludeDirectoryTree: includeDirectoryTree,
		ShowLineNumbers:      showLineNumbers,
		IncludeContent:       includeContent,
		RemoveComments:       removeComments,
		RemoveEmptyLines:     removeEmptyLines,
		CompressCode:         compressCode,
	}

	// Create streaming writer based on format
	writer, err := output.NewStreamingWriter(outFile, outputFormat, outputOpts)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Write header
	scanTime := time.Now().Format(time.RFC3339)
	if err := writer.WriteHeader(absPath, scanTime); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create scanner with streaming handler
	scanOpts := scanner.ScanOptions{
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

	// Each file gets written immediately, then discarded
	streamingScanner := scanner.NewStreamingScanner(absPath, scanOpts, writer.WriteFile)
	// Set tree writer callback
	streamingScanner.SetTreeWriter(writer.WriteTree)

	// Perform the scan (streaming mode!)
	fmt.Println("Streaming scan in progress...")
	stats, err := streamingScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Write footer with final statistics
	if err := writer.WriteFooter(stats); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

	fmt.Printf("\nOutput written to %s\n", outputFilePath)

	// Enhanced scan summary
	fmt.Printf("\nScan Summary:\n")
	fmt.Printf("  Files processed: %d\n", stats.TotalFiles)
	fmt.Printf("  Total size: %s\n", utils.FormatBytes(stats.TotalSize))
	fmt.Printf("  Text files: %d, Binary files: %d\n", stats.TextFiles, stats.BinaryFiles)

	// Show top file types
	if len(stats.LanguageCounts) > 0 {
		fmt.Printf("  Languages detected: ")
		count := 0
		for lang, num := range stats.LanguageCounts {
			if count > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s (%d)", lang, num)
			count++
			if count >= 5 { // Show top 5
				break
			}
		}
		fmt.Printf("\n")
	}

	return nil
}
