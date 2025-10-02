package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opskraken/codeecho-cli/output"
	"github.com/opskraken/codeecho-cli/scanner"
	"github.com/opskraken/codeecho-cli/utils"
	"github.com/spf13/cobra"
)

var (
	docOutputFile string
	docType       string
)

// ScanResult is an alias for scanner.ScanResult for backward compatibility
type ScanResult = scanner.ScanResult
type FileInfo = scanner.FileInfo

// docCmd represents the doc command
var docCmd = &cobra.Command{
	Use:   "doc [path]",
	Short: "Generate documentation from repository scan",
	Long: `Generate documentation automatically from a repository scan.
This command first scans the repository and then generates different types
of documentation based on the codebase structure and content.

Supported documentation types:
• readme    - Generate a comprehensive README.md
• api       - Generate API documentation (for web projects)
• overview  - Generate project overview documentation`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDoc,
}

func init() {
	rootCmd.AddCommand(docCmd)

	// Add flags
	docCmd.Flags().StringVarP(&docOutputFile, "output", "o", "", "Output file (default: README.md)")
	docCmd.Flags().StringVarP(&docType, "type", "t", "readme", "Documentation type: readme, api, overview")
}

// scanRepository uses AnalysisScanner for full repository analysis
func scanRepository(path string) (*ScanResult, error) {
	opts := scanner.ScanOptions{
		IncludeSummary:       true,
		IncludeDirectoryTree: true,
		ShowLineNumbers:      false,
		OutputParsableFormat: true,
		CompressCode:         false,
		RemoveComments:       false,
		RemoveEmptyLines:     false,
		ExcludeDirs:          []string{".git", "node_modules", "vendor", ".vscode", ".idea", "target", "build", "dist"},
		IncludeExts:          []string{".go", ".js", ".ts", ".jsx", ".tsx", ".json", ".md", ".html", ".css", ".py", ".java", ".cpp", ".c", ".h", ".rs", ".rb", ".php", ".yml", ".yaml", ".toml", ".xml"},
		IncludeContent:       true, // Doc needs content for analysis
	}

	// Use analysis scanner (not streaming) for full in-memory analysis
	analysisScanner := scanner.NewAnalysisScanner(path, opts)
	return analysisScanner.Scan()
}

func generateDirectoryTree(files []FileInfo) string {
	return output.GenerateDirectoryTree(files)
}

func formatBytes(bytes int64) string {
	return utils.FormatBytes(bytes)
}

func runDoc(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Validate path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}

	// Get absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Generating %s documentation for %s...\n", docType, absPath)

	// First, scan the repository using AnalysisScanner
	result, err := scanRepository(absPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Generate documentation based on type
	var doc string
	switch strings.ToLower(docType) {
	case "readme":
		doc, err = generateReadmeDoc(result)
	case "api":
		doc, err = generateAPIDoc(result)
	case "overview":
		doc, err = generateOverviewDoc(result)
	default:
		return fmt.Errorf("unsupported documentation type: %s (supported: readme, api, overview)", docType)
	}

	if err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Determine output file
	outputFile := docOutputFile
	if outputFile == "" {
		switch docType {
		case "readme":
			outputFile = "README.md"
		case "api":
			outputFile = "API.md"
		case "overview":
			outputFile = "OVERVIEW.md"
		}
	}

	// Write documentation
	err = os.WriteFile(outputFile, []byte(doc), 0644)
	if err != nil {
		return fmt.Errorf("failed to write documentation file: %w", err)
	}

	fmt.Printf("Documentation written to %s\n", outputFile)
	fmt.Printf("Documentation Summary: %d files analyzed\n", result.TotalFiles)

	return nil
}

func generateReadmeDoc(result *ScanResult) (string, error) {
	var builder strings.Builder

	// Extract project name from path
	projectName := filepath.Base(result.RepoPath)

	// Header
	builder.WriteString(fmt.Sprintf("# %s\n\n", strings.Title(projectName)))
	builder.WriteString("Generated documentation by CodeEcho\n\n")

	// Project Overview
	builder.WriteString("## Overview\n\n")
	builder.WriteString("This project contains ")
	builder.WriteString(fmt.Sprintf("%d files ", result.TotalFiles))
	builder.WriteString(fmt.Sprintf("with a total size of %s.\n\n", formatBytes(result.TotalSize)))

	// Technology Stack
	builder.WriteString("## Technology Stack\n\n")
	languages := analyzeTechStack(result.Files)
	for lang, count := range languages {
		builder.WriteString(fmt.Sprintf("- **%s**: %d files\n", lang, count))
	}
	builder.WriteString("\n")

	// Project Structure
	builder.WriteString("## Project Structure\n\n")
	builder.WriteString("```\n")
	builder.WriteString(generateDirectoryTree(result.Files))
	builder.WriteString("```\n\n")

	// Key Files
	builder.WriteString("## Key Files\n\n")
	keyFiles := identifyKeyFiles(result.Files)
	for _, file := range keyFiles {
		builder.WriteString(fmt.Sprintf("- **%s**: %s\n", file.RelativePath, describeFile(file)))
	}
	builder.WriteString("\n")

	// Getting Started (if applicable)
	if hasConfigFiles(result.Files) {
		builder.WriteString("## Getting Started\n\n")
		builder.WriteString(generateGettingStarted(result.Files))
	}

	// Footer
	builder.WriteString("---\n\n")
	builder.WriteString(fmt.Sprintf("*Documentation generated by CodeEcho on %s*\n",
		time.Now().Format("January 2, 2006")))

	return builder.String(), nil
}

func generateAPIDoc(result *ScanResult) (string, error) {
	var builder strings.Builder

	projectName := filepath.Base(result.RepoPath)

	builder.WriteString(fmt.Sprintf("# %s API Documentation\n\n", strings.Title(projectName)))

	// Look for API-related files
	apiFiles := findAPIFiles(result.Files)
	if len(apiFiles) == 0 {
		builder.WriteString("No API endpoints detected in this project.\n\n")
		builder.WriteString("This documentation type is best suited for web applications with API endpoints.\n")
		return builder.String(), nil
	}

	builder.WriteString("## API Endpoints\n\n")

	for _, file := range apiFiles {
		builder.WriteString(fmt.Sprintf("### %s\n\n", file.RelativePath))

		// Basic analysis of the file
		if strings.Contains(strings.ToLower(file.Content), "router") ||
			strings.Contains(strings.ToLower(file.Content), "endpoint") ||
			strings.Contains(strings.ToLower(file.Content), "handler") {
			builder.WriteString("Contains API route definitions.\n\n")
		}
	}

	return builder.String(), nil
}

func generateOverviewDoc(result *ScanResult) (string, error) {
	var builder strings.Builder

	projectName := filepath.Base(result.RepoPath)

	builder.WriteString(fmt.Sprintf("# %s - Project Overview\n\n", strings.Title(projectName)))

	// Statistics
	builder.WriteString("## Project Statistics\n\n")
	builder.WriteString(fmt.Sprintf("- **Total Files**: %d\n", result.TotalFiles))
	builder.WriteString(fmt.Sprintf("- **Total Size**: %s\n", formatBytes(result.TotalSize)))
	builder.WriteString(fmt.Sprintf("- **Last Scanned**: %s\n\n", result.ScanTime))

	// File Distribution
	builder.WriteString("## File Distribution\n\n")
	languages := analyzeTechStack(result.Files)
	for lang, count := range languages {
		percentage := float64(count) / float64(result.TotalFiles) * 100
		builder.WriteString(fmt.Sprintf("- %s: %d files (%.1f%%)\n", lang, count, percentage))
	}
	builder.WriteString("\n")

	// Directory Analysis
	builder.WriteString("## Directory Analysis\n\n")
	dirCounts := analyzeDirectories(result.Files)
	for dir, count := range dirCounts {
		if count > 1 { // Only show directories with multiple files
			builder.WriteString(fmt.Sprintf("- `%s/`: %d files\n", dir, count))
		}
	}

	return builder.String(), nil
}

// Helper functions
func analyzeTechStack(files []FileInfo) map[string]int {
	languages := make(map[string]int)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.RelativePath))
		switch ext {
		case ".go":
			languages["Go"]++
		case ".js":
			languages["JavaScript"]++
		case ".ts":
			languages["TypeScript"]++
		case ".py":
			languages["Python"]++
		case ".java":
			languages["Java"]++
		case ".cpp", ".cc":
			languages["C++"]++
		case ".c":
			languages["C"]++
		case ".rs":
			languages["Rust"]++
		case ".rb":
			languages["Ruby"]++
		case ".php":
			languages["PHP"]++
		case ".html":
			languages["HTML"]++
		case ".css":
			languages["CSS"]++
		case ".json":
			languages["JSON"]++
		case ".md":
			languages["Markdown"]++
		case ".yml", ".yaml":
			languages["YAML"]++
		default:
			if ext != "" {
				languages["Other"]++
			}
		}
	}

	return languages
}

func identifyKeyFiles(files []FileInfo) []FileInfo {
	var keyFiles []FileInfo

	keyPatterns := []string{
		"main.go", "main.js", "index.js", "app.js",
		"package.json", "go.mod", "requirements.txt",
		"dockerfile", "docker-compose.yml",
		"readme.md", "license",
	}

	for _, file := range files {
		fileName := strings.ToLower(filepath.Base(file.RelativePath))
		for _, pattern := range keyPatterns {
			if fileName == pattern {
				keyFiles = append(keyFiles, file)
				break
			}
		}
	}

	return keyFiles
}

func describeFile(file FileInfo) string {
	fileName := strings.ToLower(filepath.Base(file.RelativePath))

	descriptions := map[string]string{
		"main.go":            "Main application entry point",
		"main.js":            "Main JavaScript file",
		"index.js":           "Application entry point",
		"package.json":       "Node.js project configuration",
		"go.mod":             "Go module definition",
		"dockerfile":         "Docker container configuration",
		"docker-compose.yml": "Docker services configuration",
		"readme.md":          "Project documentation",
	}

	if desc, exists := descriptions[fileName]; exists {
		return desc
	}

	return fmt.Sprintf("Project file (%s)", formatBytes(file.Size))
}

func hasConfigFiles(files []FileInfo) bool {
	configPatterns := []string{"package.json", "go.mod", "requirements.txt", "dockerfile"}

	for _, file := range files {
		fileName := strings.ToLower(filepath.Base(file.RelativePath))
		for _, pattern := range configPatterns {
			if fileName == pattern {
				return true
			}
		}
	}
	return false
}

func generateGettingStarted(files []FileInfo) string {
	var builder strings.Builder

	// Check for different project types
	hasPackageJSON := false
	hasGoMod := false
	hasDockerfile := false

	for _, file := range files {
		fileName := strings.ToLower(filepath.Base(file.RelativePath))
		switch fileName {
		case "package.json":
			hasPackageJSON = true
		case "go.mod":
			hasGoMod = true
		case "dockerfile":
			hasDockerfile = true
		}
	}

	if hasPackageJSON {
		builder.WriteString("### Node.js Project\n")
		builder.WriteString("```bash\n")
		builder.WriteString("npm install\n")
		builder.WriteString("npm start\n")
		builder.WriteString("```\n\n")
	}

	if hasGoMod {
		builder.WriteString("### Go Project\n")
		builder.WriteString("```bash\n")
		builder.WriteString("go mod tidy\n")
		builder.WriteString("go run main.go\n")
		builder.WriteString("```\n\n")
	}

	if hasDockerfile {
		builder.WriteString("### Docker\n")
		builder.WriteString("```bash\n")
		builder.WriteString("docker build -t app .\n")
		builder.WriteString("docker run -p 8080:8080 app\n")
		builder.WriteString("```\n\n")
	}

	return builder.String()
}

func findAPIFiles(files []FileInfo) []FileInfo {
	var apiFiles []FileInfo

	apiPatterns := []string{"router", "route", "handler", "controller", "api", "endpoint"}

	for _, file := range files {
		fileName := strings.ToLower(file.RelativePath)
		content := strings.ToLower(file.Content)

		// Check filename
		for _, pattern := range apiPatterns {
			if strings.Contains(fileName, pattern) {
				apiFiles = append(apiFiles, file)
				break
			}
		}

		// Check content for API-related keywords
		if strings.Contains(content, "http.") ||
			strings.Contains(content, "express") ||
			strings.Contains(content, "@requestmapping") ||
			strings.Contains(content, "@getmapping") {
			apiFiles = append(apiFiles, file)
		}
	}

	return apiFiles
}

func analyzeDirectories(files []FileInfo) map[string]int {
	dirCounts := make(map[string]int)

	for _, file := range files {
		dir := filepath.Dir(file.RelativePath)
		if dir != "." {
			dirCounts[dir]++
		}
	}

	return dirCounts
}
