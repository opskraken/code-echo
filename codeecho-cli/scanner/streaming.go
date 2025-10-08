package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/opskraken/codeecho-cli/utils"
)

type StreamingScanner struct {
	rootPath    string
	opts        ScanOptions
	fileHandler func(*FileInfo) error
	treeWriter  func([]string) error

	// NEW: Progress and error tracking
	progressCallback ProgressCallback
	errors           []ScanError

	stats     *StreamingStats
	filePaths []string

	// NEW: Timing
	startTime time.Time
}

// StreamingStats tracks lightweight counters (not full file data)
type StreamingStats struct {
	TotalFiles     int
	TotalSize      int64
	TextFiles      int
	BinaryFiles    int
	LanguageCounts map[string]int
}

// NewStreamingScanner creates a scanner that calls fileHandler for each file
func NewStreamingScanner(rootPath string, opts ScanOptions, fileHandler func(*FileInfo) error) *StreamingScanner {
	return &StreamingScanner{
		rootPath:    rootPath,
		opts:        opts,
		fileHandler: fileHandler,
		stats: &StreamingStats{
			LanguageCounts: make(map[string]int),
		},
		filePaths: []string{},
		errors:    []ScanError{}, // Initialize error slice
	}
}

// NEW: Set progress callback
// Why: Allow external progress monitoring
func (s *StreamingScanner) SetProgressCallback(callback ProgressCallback) {
	s.progressCallback = callback
}

func (s *StreamingScanner) SetTreeWriter(treeWriter func([]string) error) {
	s.treeWriter = treeWriter
}

func (s *StreamingScanner) GetFilePaths() []string {
	return s.filePaths
}

// NEW: Get collected errors
// Why: Return all errors at the end
func (s *StreamingScanner) GetErrors() []ScanError {
	return s.errors
}

// NEW: Report progress
// Why: Centralized progress reporting
func (s *StreamingScanner) reportProgress(phase string, currentFile string) {
	if s.progressCallback == nil {
		return
	}

	progress := ScanProgress{
		Phase:          phase,
		CurrentFile:    currentFile,
		ProcessedFiles: s.stats.TotalFiles,
		TotalFiles:     len(s.filePaths),
		BytesProcessed: s.stats.TotalSize,
	}

	// calculate percentage
	if len(s.filePaths) > 0 {
		progress.Percentage = float64(s.stats.TextFiles) / float64(len(s.filePaths)) * 100
	}

	s.progressCallback(progress)
}

// NEW: Record error
// Why: Collect errors instead of just logging
func (s *StreamingScanner) recordError(path string, phase string, err error, skipped bool) {
	s.errors = append(s.errors, ScanError{
		Path:    path,
		Phase:   phase,
		Error:   err,
		Skipped: skipped,
	})

	// Still log for debugging
	if skipped {
		fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", path, err)
	}
}

// Update: Enhanced with error tracking
func (s *StreamingScanner) collectPaths() error {
	s.reportProgress("collecting", "scanning directories...")

	return filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
			return nil // Continue scanning
		}

		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name(), s.opts.ExcludeDirs) {
			return filepath.SkipDir
		}

		// Collect file paths only
		if !d.IsDir() && shouldIncludeFile(path, s.opts.IncludeExts) {
			relativePath := utils.GetRelativePath(s.rootPath, path)
			s.filePaths = append(s.filePaths, relativePath)
		}

		return nil
	})
}

// Scan walks the directory and calls fileHandler for each file
// This is where streaming happens - we don't accumulate anything!
// Scan - Enhanced with progress and error tracking
func (s *StreamingScanner) Scan() (*StreamingStats, error) {
	s.startTime = time.Now()

	// Phase 1: Collect paths if tree is needed
	if s.opts.IncludeDirectoryTree {
		if err := s.collectPaths(); err != nil {
			return nil, fmt.Errorf("failed to collect paths: %w", err)
		}

		// Write tree immediately after collecting paths
		if s.treeWriter != nil {
			s.reportProgress("tree", "writing directory structure...")
			if err := s.treeWriter(s.filePaths); err != nil {
				return nil, fmt.Errorf("failed to write tree: %w", err)
			}
		}
	}

	// Phase 2: Process files and stream content
	s.reportProgress("scanning", "processing files...")

	err := filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.recordError(path, "scan", err, true)
			return nil // Continue
		}

		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name(), s.opts.ExcludeDirs) {
			return filepath.SkipDir
		}

		// Process files only
		if !d.IsDir() && shouldIncludeFile(path, s.opts.IncludeExts) {
			if err := s.processFile(path, d); err != nil {
				// Error recorded in processFile
				return nil // Continue scanning
			}
		}

		return nil
	})

	return s.stats, err
}

// Update: Separated file processing
// Why: Makes error handling cleaner and more testable
func (s *StreamingScanner) processFile(path string, d fs.DirEntry) error {
	info, err := d.Info()
	if err != nil {
		s.recordError(path, "stat", err, true)
		return err
	}

	relativePath := utils.GetRelativePath(s.rootPath, path)
	s.reportProgress("scanning", relativePath)

	language := detectLanguage(path)
	extension := filepath.Ext(path)

	fileInfo := FileInfo{
		Path:             path,
		RelativePath:     relativePath,
		Size:             info.Size(),
		SizeFormatted:    utils.FormatBytes(info.Size()),
		ModTime:          info.ModTime().Format(time.RFC3339),
		ModTimeFormatted: info.ModTime().Format("2006-01-02 15:04:05"),
		Language:         language,
		Extension:        extension,
		IsText:           isTextFile(path, extension),
	}

	// Read and process content if requested
	if s.opts.IncludeContent && fileInfo.IsText {
		content, err := os.ReadFile(path)
		if err != nil {
			s.recordError(path, "read", err, true)
			// Continue with empty content
		} else {
			// ENHANCED: Try content-based detection if language unknown
			if fileInfo.Language == "" {
				fileInfo.Language = detectLanguageFromContent(path, content)
			}

			// ENHANCED: Re-check if text using content
			if !fileInfo.IsText && isTextContent(content) {
				fileInfo.IsText = true
			}

			processedContent := processFileContent(string(content), fileInfo.Language, s.opts)
			fileInfo.Content = processedContent
			fileInfo.LineCount = utils.CountLines(processedContent)
		}
	}

	// Update statistics
	s.stats.TotalFiles++
	s.stats.TotalSize += info.Size()

	if fileInfo.IsText {
		s.stats.TextFiles++
	} else {
		s.stats.BinaryFiles++
	}

	if fileInfo.Language != "" {
		s.stats.LanguageCounts[fileInfo.Language]++
	}

	// Call handler immediately, then discard from memory
	if err := s.fileHandler(&fileInfo); err != nil {
		s.recordError(path, "write", err, false)
		return fmt.Errorf("error writing file %s: %w", path, err)
	}

	return nil
}
