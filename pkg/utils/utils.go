package utils

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

// Confirm asks for user confirmation
func Confirm(prompt string) bool {
    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("%s [y/N]: ", prompt)
    response, err := reader.ReadString('\n')
    if err != nil {
        return false
    }
    
    response = strings.ToLower(strings.TrimSpace(response))
    return response == "y" || response == "yes"
}

// FormatSize formats bytes to human readable format
func FormatSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    
    return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TruncateString truncates a string to specified length
func TruncateString(str string, length int) string {
    if len(str) <= length {
        return str
    }
    
    if length <= 3 {
        return str[:length]
    }
    
    return str[:length-3] + "..."
}

// ParseLabels parses key=value label pairs
func ParseLabels(labels []string) (map[string]string, error) {
    result := make(map[string]string)
    
    for _, label := range labels {
        parts := strings.SplitN(label, "=", 2)
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid label format: %s", label)
        }
        
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        
        if key == "" {
            return nil, fmt.Errorf("empty label key in: %s", label)
        }
        
        result[key] = value
    }
    
    return result, nil
}
