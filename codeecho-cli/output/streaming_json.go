package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/opskraken/codeecho-cli/config"
	"github.com/opskraken/codeecho-cli/scanner"
	"github.com/opskraken/codeecho-cli/utils"
)

// StreamingJSONWriter writes JSON output incrementally
type StreamingJSONWriter struct {
	writer    *bufio.Writer
	opts      config.OutputOptions
	stats     *scanner.StreamingStats
	firstFile bool // Track if this is the first file (for comma handling)
}

func NewStreamingJSONWriter(w io.Writer, opts config.OutputOptions) *StreamingJSONWriter {
	return &StreamingJSONWriter{
		writer: bufio.NewWriterSize(w, 65536),
		opts:   opts,
		stats: &scanner.StreamingStats{
			LanguageCounts: make(map[string]int),
		},
		firstFile: true,
	}
}

func (w *StreamingJSONWriter) WriteHeader(repoPath string, scanTime string) error {
	// Start JSON object
	if _, err := w.writer.WriteString("{\n"); err != nil {
		return err
	}

	// Write repo metadata
	repoInfo := fmt.Sprintf(`  "repo_path": %s,
  "scan_time": %s,
  "processed_by": "CodeEcho CLI",
  "files": [
`, jsonString(repoPath), jsonString(scanTime))

	if _, err := w.writer.WriteString(repoInfo); err != nil {
		return err
	}

	return nil
}

func (w *StreamingJSONWriter) WriteTree(paths []string) error {
	if !w.opts.IncludeDirectoryTree || len(paths) == 0 {
		return nil
	}

	// Convert paths to FileInfo structs (minimal data needed for tree)
	fileInfos := make([]scanner.FileInfo, len(paths))
	for i, path := range paths {
		fileInfos[i] = scanner.FileInfo{RelativePath: path}
	}

	tree := GenerateDirectoryTree(fileInfos)

	// Add tree field before files array
	treeField := fmt.Sprintf(`  "directory_tree": %s,
`, jsonString(tree))

	if _, err := w.writer.WriteString(treeField); err != nil {
		return err
	}

	return nil
}

func (w *StreamingJSONWriter) WriteFile(file *scanner.FileInfo) error {
	// Update stats
	w.stats.TotalFiles++
	w.stats.TotalSize += file.Size

	if file.IsText {
		w.stats.TextFiles++
	} else {
		w.stats.BinaryFiles++
	}

	if file.Language != "" {
		w.stats.LanguageCounts[file.Language]++
	}

	// Add comma before all files except the first
	// This is why we need firstFile flag
	if !w.firstFile {
		if _, err := w.writer.WriteString(",\n"); err != nil {
			return err
		}
	}
	w.firstFile = false

	// Marshal file to JSON (Go does this automatically)
	fileJSON, err := json.MarshalIndent(file, "    ", "  ")
	if err != nil {
		return err
	}

	// Write with proper indentation
	if _, err := w.writer.WriteString("    "); err != nil {
		return err
	}
	if _, err := w.writer.Write(fileJSON); err != nil {
		return err
	}

	return nil
}

func (w *StreamingJSONWriter) WriteFooter(stats *scanner.StreamingStats) error {
	// Close files array
	if _, err := w.writer.WriteString("\n  ],\n"); err != nil {
		return err
	}

	// Write statistics
	statsJSON := fmt.Sprintf(`  "statistics": {
    "total_files": %d,
    "total_size": %s,
    "text_files": %d,
    "binary_files": %d
  }
}
`, stats.TotalFiles, jsonString(utils.FormatBytes(stats.TotalSize)), stats.TextFiles, stats.BinaryFiles)

	if _, err := w.writer.WriteString(statsJSON); err != nil {
		return err
	}

	return nil
}

func (w *StreamingJSONWriter) Close() error {
	return w.writer.Flush()
}

// jsonString properly escapes a string for JSON
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
