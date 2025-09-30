package config

type OutputOptions struct {
	IncludeSummary       bool
	IncludeDirectoryTree bool
	ShowLineNumbers      bool
	IncludeContent       bool
	RemoveComments       bool
	RemoveEmptyLines     bool
	CompressCode         bool
}