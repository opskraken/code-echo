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
	stats       *StreamingStats
	filePaths   []string
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
	}
}

func (s *StreamingScanner) SetTreeWriter(treeWriter func([]string) error) {
	s.treeWriter = treeWriter
}

func (s *StreamingScanner) GetFilePaths() []string {
	return s.filePaths
}

func (s *StreamingScanner) collectPaths() error {
	return filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
			return nil
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
func (s *StreamingScanner) Scan() (*StreamingStats, error) {
	{
		// collect paths if tree is needed
		if s.opts.IncludeDirectoryTree {
			if err := s.collectPaths(); err != nil {
				return nil, fmt.Errorf("failed to collect paths: %w", err)
			}
			// Write tree immediately after collecting paths
			if s.treeWriter != nil {
				if err := s.treeWriter(s.filePaths); err != nil {
					return nil, fmt.Errorf("failed to write tree: %w", err)
				}
			}
		}
		//  process files and stream content
		err := filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				//	Log error
				fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
				return nil
			}

			// Skip excluded directories
			if d.IsDir() && shouldExcludeDir(d.Name(), s.opts.ExcludeDirs) {
				return filepath.SkipDir
			}

			// Process files only
			if !d.IsDir() && shouldIncludeFile(path, s.opts.IncludeExts) {
				info, err := d.Info()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: cannot stat %s: %v\n", path, err)
					return nil
				}

				relativePath := utils.GetRelativePath(s.rootPath, path)
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
						// Log but continue
						fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v\n", path, err)
					} else {
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

				//Call handler immediately, then discard from memory
				if err := s.fileHandler(&fileInfo); err != nil {
					return fmt.Errorf("error writing file %s: %w", path, err)
				}
			}

			return nil
		})

		return s.stats, err
	}
}
