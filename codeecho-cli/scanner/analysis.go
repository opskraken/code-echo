package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/opskraken/codeecho-cli/utils"
)

// AnalysisScanner collects full data in memory for documentation analysis
// This is separate from StreamingScanner because doc generation needs
// to analyze the entire codebase at once
type AnalysisScanner struct {
	rootPath string
	opts     ScanOptions
}

// NewAnalysisScanner creates a scanner for full repository analysis
func NewAnalysisScanner(rootPath string, opts ScanOptions) *AnalysisScanner {
	return &AnalysisScanner{
		rootPath: rootPath,
		opts:     opts,
	}
}

// Scan performs a full repository scan and returns complete results
// Unlike StreamingScanner, this keeps all data in memory
func (a *AnalysisScanner) Scan() (*ScanResult, error) {
	result := &ScanResult{
		RepoPath:       a.rootPath,
		ScanTime:       time.Now().Format(time.RFC3339),
		Files:          []FileInfo{},
		ProcessedBy:    "CodeEcho CLI",
		LanguageCounts: make(map[string]int),
	}

	err := filepath.WalkDir(a.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name(), a.opts.ExcludeDirs) {
			return filepath.SkipDir
		}

		// Process files only
		if !d.IsDir() && shouldIncludeFile(path, a.opts.IncludeExts) {
			info, err := d.Info()
			if err != nil {
				return err
			}

			relativePath := utils.GetRelativePath(a.rootPath, path)
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

			// Include content if requested and it's a text file
			if a.opts.IncludeContent && fileInfo.IsText {
				content, err := os.ReadFile(path)
				if err == nil {
					processedContent := processFileContent(string(content), fileInfo.Language, a.opts)
					fileInfo.Content = processedContent
					fileInfo.LineCount = utils.CountLines(processedContent)
				}
			}

			result.Files = append(result.Files, fileInfo)
			result.TotalFiles++
			result.TotalSize += info.Size()

			if fileInfo.IsText {
				result.TextFiles++
			} else {
				result.BinaryFiles++
			}

			if fileInfo.Language != "" {
				result.LanguageCounts[fileInfo.Language]++
			}
		}

		return nil
	})

	// Sort files by path for consistent output
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].RelativePath < result.Files[j].RelativePath
	})

	return result, err
}
