package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/opskraken/codeecho-cli/utils"
)
func ScanRepository(rootPath string, opts ScanOptions) (*ScanResult, error) {
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
		if d.IsDir() && shouldExcludeDir(d.Name(), opts.ExcludeDirs) {
			return filepath.SkipDir
		}

		// Process files only
		if !d.IsDir() && shouldIncludeFile(path, opts.IncludeExts) {
			info, err := d.Info()
			if err != nil {
				return err
			}

			relativePath := utils.GetRelativePath(rootPath, path)
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
			if opts.IncludeContent && fileInfo.IsText {
				content, err := os.ReadFile(path)
				if err == nil {
					processedContent := processFileContent(string(content), fileInfo.Language, opts)
					fileInfo.Content = processedContent
					fileInfo.LineCount = utils.CountLines(processedContent)
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
