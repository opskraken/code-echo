package scanner

import (
	"bytes"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "jsx",
		".tsx":  "tsx",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".h":    "c",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".css":  "css",
		".html": "html",
		".json": "json",
		".md":   "markdown",
		".yml":  "yaml",
		".yaml": "yaml",
		".toml": "toml",
		".xml":  "xml",
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return ""
}

// ENHANCED: Now checks content for unknown types
// Files without extensions or misnamed files need content inspection
func isTextFile(path, extension string) bool {
	// Step 1: Try fast extension-based detection first
	ext := strings.ToLower(extension)
	if isTextExtension(ext) {
		return true
	}

	// Step 2: Try filename-based detection (no extension files)
	fileName := strings.ToLower(filepath.Base(path))
	if isTextFilename(fileName) {
		return true
	}

	// Step 3: If still unknown, we'll need content sampling
	// (This happens in the scanner when we read the file)
	return false
}

// Separated extension checking for clarity
// Makes the logic clearer and easier to maintain
func isTextExtension(ext string) bool {
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".rst": true, ".asciidoc": true,
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true,
		".cs": true, ".php": true, ".rb": true, ".rs": true, ".swift": true, ".kt": true,
		".html": true, ".htm": true, ".xml": true, ".xhtml": true,
		".css": true, ".scss": true, ".sass": true, ".less": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true, ".cfg": true, ".conf": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".ps1": true, ".bat": true, ".cmd": true,
		".sql": true, ".graphql": true, ".gql": true,
		".dockerfile": true, ".gitignore": true, ".gitattributes": true,
		".makefile": true, ".cmake": true,
		".r": true, ".rmd": true, ".m": true, ".scala": true, ".clj": true, ".hs": true,
		".vim": true, ".lua": true, ".pl": true, ".tcl": true,
		".tex": true, ".bib": true, ".cls": true, ".sty": true,
		".csv": true, ".tsv": true, ".log": true,
	}
	return textExtensions[ext]
}

// Separated filename checking for clarity
// Keeps the logic organized and testable
func isTextFilename(fileName string) bool {
	textFiles := map[string]bool{
		"readme": true, "license": true, "changelog": true, "contributing": true,
		"authors": true, "contributors": true, "copying": true, "install": true,
		"news": true, "thanks": true, "todo": true, "version": true,
		"makefile": true, "dockerfile": true, "jenkinsfile": true,
		"gemfile": true, "rakefile": true, "guardfile": true, "procfile": true,
		".gitignore": true, ".gitattributes": true, ".dockerignore": true,
		".eslintrc": true, ".prettierrc": true, ".babelrc": true,
	}
	return textFiles[fileName]
}

// Content-based text detection
// Last resort for files we can't identify by name/extension
// Algorithm:
//   1. Check for null bytes (binary indicator)
//   2. Validate UTF-8 encoding
//   3. Check printable character ratio

func isTextContent(data []byte) bool {

	if len(data) == 0 {
		return true // Empty file is technically text
	}

	// Sample size: Check first 8KB (enough to identify most files)
	// 8KB? Balance between speed and accuracy
	sampleSize := 8192
	if len(data) < sampleSize {
		sampleSize = len(data)
	}
	sample := data[:sampleSize]

	// Rule 1: Binary files often contain null bytes
	// Why: Text files rarely have \x00 characters
	if bytes.Contains(sample, []byte{0}) {
		return false
	}

	// Rule 2: Text files must be valid UTF-8
	// Why: If it's not valid UTF-8, it's likely binary
	if !utf8.Valid(sample) {
		// Allow some invalid UTF-8 for legacy encodings
		// If more than 10% is invalid, it's likely binary
		invalidCount := 0
		for len(sample) > 0 {
			r, size := utf8.DecodeRune(sample)
			if r == utf8.RuneError && size == 1 {
				invalidCount++
			}
			sample = sample[size:]
		}
		invalidRatio := float64(invalidCount) / float64(sampleSize)
		if invalidRatio > 0.1 {
			return false
		}
	}

	// Rule 3: Check printable character ratio
	// Text files should be mostly printable
	printableCount := 0
	for _, b := range sample {
		// Printable ASCII: 0x20-0x7E, plus common whitespace
		if (b >= 0x20 && b <= 0x7E) || b == '\r' || b == '\t' {
			printableCount++
		}
	}
	printableRatio := float64(printableCount) / float64(len(sample))
	// If 80%+ is printable, it's probably text
	// 80%? Allows for some special characters in UTF-8
	return printableRatio >= 0.8
}

// Detect language from file content (shebang, patterns)
// Files without extensions need content analysis
func detectLanguageFromContent(path string, content []byte) string {
	// try shebang for scripts
	if lang := detectFromShebang(content); lang != "" {
		return lang
	}

	// try content patterns
	if lang := detectFromPatterns(content); lang != "" {
		return lang
	}

	return ""
}

// Shebang detection
// Script files often lack extensions but have shebangs
func detectFromShebang(content []byte) string {
	if len(content) < 3 || !bytes.HasPrefix(content, []byte("#!")) {
		return ""
	}

	// Read first line
	firstLine := content
	if idx := bytes.IndexByte(content, '\n'); idx > 0 {
		firstLine = content[:idx]
	}

	shebang := string(firstLine)

	// Common shebang patterns
	patterns := map[string]string{
		"python":    "python",
		"node":      "javascript",
		"ruby":      "ruby",
		"perl":      "perl",
		"bash":      "bash",
		"sh":        "shell",
		"/bin/sh":   "shell",
		"/bin/bash": "bash",
		"php":       "php",
	}

	for pattern, lang := range patterns {
		if strings.Contains(strings.ToLower(shebang), pattern) {
			return lang
		}
	}

	return ""
}

// Pattern-based detection
// Some file types have distinctive patterns
func detectFromPatterns(content []byte) string {
	// Sample first 1KB for pattern matching
	// Why 1KB? Most file signatures appear early
	sampleSize := 1024
	if len(content) < sampleSize {
		sampleSize = len(content)
	}
	sample := strings.ToLower(string(content[:sampleSize]))

	// Check for distinctive patterns
	patterns := []struct {
		pattern string
		lang    string
	}{
		{"<?php", "php"},
		{"<?xml", "xml"},
		{"<!doctype html", "html"},
		{"<html", "html"},
		{"import react", "jsx"},
		{"from react", "jsx"},
		{"package main", "go"},
		{"#!/usr/bin/env python", "python"},
	}

	for _, p := range patterns {
		if strings.Contains(sample, p.pattern) {
			return p.lang
		}
	}

	return ""
}
