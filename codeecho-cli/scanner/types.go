package scanner

type FileInfo struct {
	Path             string `json:"path"`
	RelativePath     string `json:"relative_path"`
	Size             int64  `json:"size"`
	SizeFormatted    string `json:"size_formatted"`
	ModTime          string `json:"mod_time"`
	ModTimeFormatted string `json:"mod_time_formatted"`
	Content          string `json:"content,omitempty"`
	Language         string `json:"language,omitempty"`
	LineCount        int    `json:"line_count,omitempty"`
	Extension        string `json:"extension,omitempty"`
	IsText           bool   `json:"is_text"`
}

type ScanResult struct {
	RepoPath       string         `json:"repo_path"`
	ScanTime       string         `json:"scan_time"`
	TotalFiles     int            `json:"total_files"`
	TotalSize      int64          `json:"total_size"`
	Files          []FileInfo     `json:"files,omitempty"`
	ProcessedBy    string         `json:"processed_by"`
	TextFiles      int            `json:"text_files"`
	BinaryFiles    int            `json:"binary_files"`
	LanguageCounts map[string]int `json:"language_counts"`
}

type ScanOptions struct {
	IncludeSummary       bool
	IncludeDirectoryTree bool
	ShowLineNumbers      bool
	OutputParsableFormat bool

	CompressCode     bool
	RemoveComments   bool
	RemoveEmptyLines bool

	ExcludeDirs    []string
	IncludeExts    []string
	IncludeContent bool
}
