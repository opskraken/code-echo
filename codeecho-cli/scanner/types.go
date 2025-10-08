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

// Progress tracking
type ScanProgress struct {
	Phase          string  // "collecting", "scanning", "writing"
	CurrentFile    string  // File currently being processed
	ProcessedFiles int     // Files completed
	TotalFiles     int     // Total files to process (0 if unknown)
	BytesProcessed int64   // Total bytes processed
	Percentage     float64 // 0-100
}

// NEW: Error tracking
type ScanError struct {
	Path    string // File path that caused error
	Phase   string // "read", "parse", "write"
	Error   error  // The actual error
	Skipped bool   // Was the file skipped or did scan fail?
}

// NEW: Complete scan report
type ScanReport struct {
	Stats        *StreamingStats
	Errors       []ScanError
	SkippedFiles int
	WarningCount int
	Duration     string
	Success      bool
}

// NEW: Progress callback
type ProgressCallback func(progress ScanProgress)
