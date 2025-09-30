package scanner

import "strings"

func shouldExcludeDir(dirName string, excludeDirs []string) bool {
	for _, excluded := range excludeDirs {
		if dirName == excluded {
			return true
		}
	}
	return false
}

func shouldIncludeFile(path string, includeExts []string) bool {
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
