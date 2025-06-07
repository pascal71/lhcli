// pkg/formatter/helpers.go
package formatter

import (
	"fmt"
	"strings"
	"time"
)

// Column represents a table column configuration
type Column struct {
	Name   string
	Width  int
	Format func(interface{}) string
}

// FormatTime formats a time value for display
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	return t.Format(time.RFC3339)
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	// Format as human-readable duration
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, "")
}

// FormatAge formats a time as age (e.g., "5d", "2h", "30m")
func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}

	age := time.Since(t)
	return FormatDuration(age)
}

// FormatBool formats a boolean value
func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// FormatList formats a slice of strings
func FormatList(items []string) string {
	if len(items) == 0 {
		return "<none>"
	}
	return strings.Join(items, ", ")
}

// FormatMap formats a map for display
func FormatMap(m map[string]string) string {
	if len(m) == 0 {
		return "<none>"
	}

	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(pairs, ", ")
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

// FormatPercent formats a percentage value
func FormatPercent(value, total float64) string {
	if total == 0 {
		return "0%"
	}

	percent := (value / total) * 100
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatStatus formats a status string with color codes (for terminal output)
func FormatStatus(status string, useColor bool) string {
	if !useColor {
		return status
	}

	// Define color codes
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorGray   = "\033[90m"
	)

	// Map status to colors
	statusLower := strings.ToLower(status)
	switch {
	case strings.Contains(statusLower, "ready"),
		strings.Contains(statusLower, "running"),
		strings.Contains(statusLower, "active"),
		strings.Contains(statusLower, "healthy"):
		return colorGreen + status + colorReset

	case strings.Contains(statusLower, "error"),
		strings.Contains(statusLower, "failed"),
		strings.Contains(statusLower, "unhealthy"):
		return colorRed + status + colorReset

	case strings.Contains(statusLower, "pending"),
		strings.Contains(statusLower, "creating"),
		strings.Contains(statusLower, "updating"):
		return colorYellow + status + colorReset

	case strings.Contains(statusLower, "unknown"),
		strings.Contains(statusLower, "terminating"):
		return colorGray + status + colorReset

	default:
		return status
	}
}

// PadRight pads a string with spaces on the right
func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// PadLeft pads a string with spaces on the left
func PadLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(" ", length-len(s)) + s
}

// Center centers a string within a given width
func Center(s string, width int) string {
	if len(s) >= width {
		return s
	}

	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
