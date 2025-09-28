// cmd/scan.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// FileInfo represents information about a scanned file
type FileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	ModTime      string `json:"mod_time"`
	Content      string `json:"content,omitempty"`
	Language     string `json:"language,omitempty"`
	LineCount    int    `json:"line_count,omitempty"`
}

// ScanResult represents the complete scan result
type ScanResult struct {
	RepoPath    string     `json:"repo_path"`
	ScanTime    string     `json:"scan_time"`
	TotalFiles  int        `json:"total_files"`
	TotalSize   int64      `json:"total_size"`
	Files       []FileInfo `json:"files"`
	ProcessedBy string     `json:"processed_by"`
}

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
  codeecho scan .                             # Basic XML scan
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
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	scanCmd.Flags().BoolVar(&includeSummary, "include-summary", true, "Include file summary section")
	scanCmd.Flags().BoolVar(&includeDirectoryTree, "include-tree", true, "Include directory structure")
	scanCmd.Flags().BoolVar(&showLineNumbers, "line-numbers", false, "Show line numbers in code blocks")
	scanCmd.Flags().BoolVar(&outputParsableFormat, "parsable", true, "Use parsable format tags")

	// File processing flags
	scanCmd.Flags().BoolVar(&compressCode, "compress-code", false, "Remove unnecessary whitespace from code")
	scanCmd.Flags().BoolVar(&removeComments, "remove-comments", false, "Strip comments from source files")
	scanCmd.Flags().BoolVar(&removeEmptyLines, "remove-empty-lines", false, "Remove empty lines from files")

	// File filtering flags
	scanCmd.Flags().BoolVar(&includeContent, "content", true, "Include file contents (disable for structure-only)")
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

	// Perform the scan
	result, err := scanRepository(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Generate output based on format
	var output string
	switch strings.ToLower(outputFormat) {
	case "xml":
		output, err = generateXMLOutput(result)
	case "json":
		output, err = generateJSONOutput(result)
	case "markdown", "md":
		output, err = generateMarkdownOutput(result)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: xml, json, markdown)", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Write output
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", outputFile)
	} else {
		// Auto-generate filename if not specified
		autoFile := generateAutoFilename(result.RepoPath, outputFormat)
		err = os.WriteFile(autoFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Output written to %s\n", autoFile)
	}

	fmt.Printf("\nScan Summary: %d files processed, %s total size\n",
		result.TotalFiles,
		formatBytes(result.TotalSize))

	return nil
}

func scanRepository(rootPath string) (*ScanResult, error) {
	result := &ScanResult{
		RepoPath:    rootPath,
		ScanTime:    time.Now().Format(time.RFC3339),
		Files:       []FileInfo{},
		ProcessedBy: "CodeEcho CLI",
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name()) {
			return filepath.SkipDir
		}

		// Process files only
		if !d.IsDir() && shouldIncludeFile(path) {
			info, err := d.Info()
			if err != nil {
				return err
			}

			fileInfo := FileInfo{
				Path:         path,
				RelativePath: getRelativePath(rootPath, path),
				Size:         info.Size(),
				ModTime:      info.ModTime().Format(time.RFC3339),
				Language:     detectLanguage(path),
			}

			// Include content if requested
			if includeContent {
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("Warning: Could not read %s: %v\n", path, err)
				} else {
					// Apply file processing
					processedContent := processFileContent(string(content), fileInfo.Language)
					fileInfo.Content = processedContent
					fileInfo.LineCount = strings.Count(processedContent, "\n") + 1
				}
			}

			result.Files = append(result.Files, fileInfo)
			result.TotalFiles++
			result.TotalSize += info.Size()
		}

		return nil
	})

	// Sort files by path for consistent output
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].RelativePath < result.Files[j].RelativePath
	})

	return result, err
}

// processFileContent applies file processing options
func processFileContent(content, language string) string {
	processed := content

	// Remove comments based on language
	if removeComments {
		processed = stripComments(processed, language)
	}

	// Remove empty lines
	if removeEmptyLines {
		processed = stripEmptyLines(processed)
	}

	// Compress code (remove unnecessary whitespace)
	if compressCode {
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

// generateXMLOutput creates Repomix-style XML output
func generateXMLOutput(result *ScanResult) (string, error) {
	var builder strings.Builder

	// XML declaration and root element
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	builder.WriteString("\n")
	builder.WriteString("<!-- This file is a merged representation of the entire codebase, combined into a single document by CodeEcho CLI. -->\n")
	builder.WriteString("<!-- The content has been processed with the following options: ")

	var options []string
	if removeComments {
		options = append(options, "comments removed")
	}
	if removeEmptyLines {
		options = append(options, "empty lines removed")
	}
	if compressCode {
		options = append(options, "code compressed")
	}
	if len(options) > 0 {
		builder.WriteString(strings.Join(options, ", "))
	} else {
		builder.WriteString("no processing applied")
	}
	builder.WriteString(" -->\n\n")

	// File summary section
	if includeSummary {
		builder.WriteString("<file_summary>\n")
		builder.WriteString("This section contains a summary of this file.\n\n")

		builder.WriteString("<purpose>\n")
		builder.WriteString("This file contains a packed representation of the entire repository's contents.\n")
		builder.WriteString("It is designed to be easily consumable by AI systems for analysis, code review,\n")
		builder.WriteString("or other automated processes.\n")
		builder.WriteString("</purpose>\n\n")

		builder.WriteString("<file_format>\n")
		builder.WriteString("The content is organized as follows:\n")
		builder.WriteString("1. This summary section\n")
		builder.WriteString("2. Repository information\n")
		if includeDirectoryTree {
			builder.WriteString("3. Directory structure\n")
			builder.WriteString("4. Multiple file entries, each consisting of:\n")
		} else {
			builder.WriteString("3. Multiple file entries, each consisting of:\n")
		}
		builder.WriteString("  - File path as an attribute\n")
		builder.WriteString("  - Full contents of the file\n")
		builder.WriteString("</file_format>\n\n")

		builder.WriteString("<usage_guidelines>\n")
		builder.WriteString("- This file should be treated as read-only. Any changes should be made to the\n")
		builder.WriteString("  original repository files, not this packed version.\n")
		builder.WriteString("- When processing this file, use the file path to distinguish\n")
		builder.WriteString("  between different files in the repository.\n")
		builder.WriteString("- Be aware that this file may contain sensitive information. Handle it with\n")
		builder.WriteString("  the same level of security as you would the original repository.\n")
		builder.WriteString("</usage_guidelines>\n\n")

		builder.WriteString("<notes>\n")
		builder.WriteString("- Some files may have been excluded based on .gitignore rules and CodeEcho's configuration\n")
		builder.WriteString("- Binary files are not included in this packed representation\n")
		builder.WriteString("- Files matching default ignore patterns are excluded\n")
		if removeComments || removeEmptyLines || compressCode {
			builder.WriteString("- File processing has been applied - content may differ from original files\n")
		}
		builder.WriteString(fmt.Sprintf("- Generated by CodeEcho CLI on %s\n", result.ScanTime))
		builder.WriteString("</notes>\n\n")

		builder.WriteString("</file_summary>\n\n")
	}

	// Repository information
	builder.WriteString("<repository_info>\n")
	builder.WriteString(fmt.Sprintf("<repo_path>%s</repo_path>\n", escapeXML(result.RepoPath)))
	builder.WriteString(fmt.Sprintf("<scan_time>%s</scan_time>\n", result.ScanTime))
	builder.WriteString(fmt.Sprintf("<total_files>%d</total_files>\n", result.TotalFiles))
	builder.WriteString(fmt.Sprintf("<total_size>%s</total_size>\n", formatBytes(result.TotalSize)))
	builder.WriteString(fmt.Sprintf("<processed_by>%s</processed_by>\n", result.ProcessedBy))
	builder.WriteString("</repository_info>\n\n")

	// Directory structure
	if includeDirectoryTree {
		builder.WriteString("<directory_structure>\n")
		builder.WriteString(generateDirectoryTree(result.Files))
		builder.WriteString("</directory_structure>\n\n")
	}

	// Files section
	builder.WriteString("<files>\n")
	builder.WriteString("This section contains the contents of the repository's files.\n\n")

	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf(`<file path="%s"`, escapeXML(file.RelativePath)))
		if file.Language != "" {
			builder.WriteString(fmt.Sprintf(` language="%s"`, file.Language))
		}
		if file.LineCount > 0 {
			builder.WriteString(fmt.Sprintf(` lines="%d"`, file.LineCount))
		}
		builder.WriteString(fmt.Sprintf(` size="%s"`, formatBytes(file.Size)))
		builder.WriteString(">\n")

		if includeContent && file.Content != "" {
			if showLineNumbers {
				builder.WriteString(addLineNumbers(file.Content))
			} else {
				builder.WriteString(escapeXML(file.Content))
			}
		} else {
			builder.WriteString("<!-- Content not included -->")
		}

		builder.WriteString("\n</file>\n\n")
	}

	builder.WriteString("</files>\n")

	return builder.String(), nil
}

// Helper functions for XML processing
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&#39;")
	return s
}

func addLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var numberedLines []string

	for i, line := range lines {
		numberedLines = append(numberedLines, fmt.Sprintf("%4d: %s", i+1, line))
	}

	return strings.Join(numberedLines, "\n")
}

// Keep existing helper functions but update them for the new structure
func shouldExcludeDir(dirName string) bool {
	for _, excluded := range excludeDirs {
		if dirName == excluded {
			return true
		}
	}
	return false
}

func shouldIncludeFile(path string) bool {
	if len(includeExts) == 0 {
		return true
	}

	for _, ext := range includeExts {
		if strings.HasSuffix(strings.ToLower(path), strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

func getRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func generateJSONOutput(result *ScanResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func generateMarkdownOutput(result *ScanResult) (string, error) {
	// Keep the existing markdown generation for compatibility
	var builder strings.Builder

	builder.WriteString("# CodeEcho Repository Scan\n\n")
	builder.WriteString(fmt.Sprintf("**Repository:** %s  \n", result.RepoPath))
	builder.WriteString(fmt.Sprintf("**Scan Time:** %s  \n", result.ScanTime))
	builder.WriteString(fmt.Sprintf("**Total Files:** %d  \n", result.TotalFiles))
	builder.WriteString(fmt.Sprintf("**Total Size:** %s  \n\n", formatBytes(result.TotalSize)))

	if includeDirectoryTree {
		builder.WriteString("## Directory Structure\n\n")
		builder.WriteString("```\n")
		builder.WriteString(generateDirectoryTree(result.Files))
		builder.WriteString("```\n\n")
	}

	builder.WriteString("## Files\n\n")
	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf("### %s\n\n", file.RelativePath))
		builder.WriteString(fmt.Sprintf("**Size:** %s", formatBytes(file.Size)))
		if file.Language != "" {
			builder.WriteString(fmt.Sprintf(" | **Language:** %s", file.Language))
		}
		if file.LineCount > 0 {
			builder.WriteString(fmt.Sprintf(" | **Lines:** %d", file.LineCount))
		}
		builder.WriteString("\n\n")

		if includeContent && file.Content != "" {
			builder.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", strings.ToLower(file.Language), file.Content))
		} else {
			builder.WriteString("*Content not included*\n\n")
		}

		builder.WriteString("---\n\n")
	}

	return builder.String(), nil
}

func generateDirectoryTree(files []FileInfo) string {
	var builder strings.Builder
	processed := make(map[string]bool)

	// Extract project root name
	if len(files) > 0 {
		firstFile := files[0].RelativePath
		parts := strings.Split(firstFile, string(filepath.Separator))
		if len(parts) > 0 {
			// Determine common root
			rootName := filepath.Base(files[0].Path)
			if rootName == "." {
				rootName = "project"
			}
			builder.WriteString(fmt.Sprintf("%s/\n", rootName))
		}
	}

	for _, file := range files {
		parts := strings.Split(file.RelativePath, string(filepath.Separator))

		for i := range parts {
			path := strings.Join(parts[:i+1], "/")
			if !processed[path] {
				indent := strings.Repeat("  ", i+1)
				if i == len(parts)-1 {
					// It's a file
					builder.WriteString(fmt.Sprintf("%s%s\n", indent, parts[i]))
				} else {
					// It's a directory
					builder.WriteString(fmt.Sprintf("%s%s/\n", indent, parts[i]))
				}
				processed[path] = true
			}
		}
	}

	return builder.String()
}

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

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// generateAutoFilename creates a filename based on project and options
func generateAutoFilename(repoPath, format string) string {
	// Get project name
	projectName := filepath.Base(repoPath)
	if projectName == "." || projectName == "/" {
		projectName = "codeecho-scan"
	}

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("20060102-150405")

	// Determine file extension
	var ext string
	switch format {
	case "json":
		ext = ".json"
	case "markdown", "md":
		ext = ".md"
	default:
		ext = ".xml"
	}

	// Build filename with processing indicators
	var suffix []string
	if removeComments {
		suffix = append(suffix, "no-comments")
	}
	if removeEmptyLines {
		suffix = append(suffix, "no-empty-lines")
	}
	if compressCode {
		suffix = append(suffix, "compressed")
	}
	if !includeContent {
		suffix = append(suffix, "structure-only")
	}

	filename := projectName
	if len(suffix) > 0 {
		filename += "-" + strings.Join(suffix, "-")
	}
	filename += "-" + timestamp + ext

	return filename
}
