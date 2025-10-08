package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CountLines(content string) int {
	if content == "" {
		return 0
	}

	lines := 1
	for _, char := range content {
		if char == '\n' {
			lines++
		}
	}

	// file ends with newline
	if strings.HasSuffix(content, "\n") {
		lines--
	}

	return lines
}

// NEW: Format duration human-readable
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dh %dm", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// NEW: Create a simple progress bar
func CreateProgressBar(current, total int, width int) string {
	if total == 0 {
		return "[" + strings.Repeat("?", width) + "]"
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d%%", bar, int(percentage*100))
}

// NEW: Format ETA based on progress
func EstimateTimeRemaining(processed, total int, elapsed time.Duration) string {
	if processed == 0 || total == 0 {
		return "calculating..."
	}

	percentage := float64(processed) / float64(total)
	if percentage == 0 {
		return "calculating..."
	}

	totalEstimated := time.Duration(float64(elapsed) / percentage)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		return "finishing..."
	}

	return FormatDuration((remaining))
}
