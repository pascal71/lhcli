// pkg/utils/size.go
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseSize parses a size string (e.g., "10Gi", "1Ti") and returns bytes
func ParseSize(size string) (int64, error) {
	if size == "" || size == "0" {
		return 0, nil
	}

	// Remove spaces and convert to uppercase
	size = strings.TrimSpace(strings.ToUpper(size))

	// Pattern: number followed by optional unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGTPE]?I?B?)?$`)
	matches := re.FindStringSubmatch(size)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	// Parse the numeric value
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", matches[1])
	}

	// Parse the unit
	unit := matches[2]
	if unit == "" {
		// No unit means bytes
		return int64(value), nil
	}

	// Normalize unit (remove trailing B if present)
	unit = strings.TrimSuffix(unit, "B")

	// Calculate multiplier based on unit
	var multiplier float64 = 1
	switch unit {
	case "K", "KI":
		multiplier = 1024
	case "M", "MI":
		multiplier = 1024 * 1024
	case "G", "GI":
		multiplier = 1024 * 1024 * 1024
	case "T", "TI":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "P", "PI":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case "E", "EI":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	case "":
		multiplier = 1
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(value * multiplier), nil
}

// FormatSizeInGi formats bytes to Gi with 2 decimal places
func FormatSizeInGi(bytes int64) string {
	gi := float64(bytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.2fGi", gi)
}
