package validation

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

// ValidateVolumeName validates a volume name
func ValidateVolumeName(name string) error {
    if name == "" {
        return fmt.Errorf("volume name cannot be empty")
    }
    
    // Kubernetes name validation
    if len(name) > 253 {
        return fmt.Errorf("volume name must be no more than 253 characters")
    }
    
    validName := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
    if !validName.MatchString(name) {
        return fmt.Errorf("volume name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
    }
    
    return nil
}

// ValidateSize validates a size string (e.g., "10Gi", "1Ti")
func ValidateSize(size string) error {
    if size == "" {
        return fmt.Errorf("size cannot be empty")
    }
    
    // Parse size with unit
    re := regexp.MustCompile(`^(\d+)([KMGT]i?)$`)
    matches := re.FindStringSubmatch(size)
    if len(matches) != 3 {
        return fmt.Errorf("invalid size format: %s (expected format: 10Gi, 1Ti, etc.)", size)
    }
    
    value, err := strconv.Atoi(matches[1])
    if err != nil {
        return fmt.Errorf("invalid size value: %s", matches[1])
    }
    
    if value <= 0 {
        return fmt.Errorf("size must be positive")
    }
    
    return nil
}

// ValidateReplicaCount validates replica count
func ValidateReplicaCount(count int) error {
    if count < 1 || count > 10 {
        return fmt.Errorf("replica count must be between 1 and 10")
    }
    return nil
}

// ValidateFrontend validates frontend type
func ValidateFrontend(frontend string) error {
    validFrontends := []string{"blockdev", "iscsi"}
    for _, valid := range validFrontends {
        if frontend == valid {
            return nil
        }
    }
    return fmt.Errorf("invalid frontend type: %s (valid types: %s)", frontend, strings.Join(validFrontends, ", "))
}

// ValidateLabels validates label format
func ValidateLabels(labels map[string]string) error {
    for key, value := range labels {
        if err := ValidateLabelKey(key); err != nil {
            return err
        }
        if err := ValidateLabelValue(value); err != nil {
            return err
        }
    }
    return nil
}

// ValidateLabelKey validates a label key
func ValidateLabelKey(key string) error {
    if key == "" {
        return fmt.Errorf("label key cannot be empty")
    }
    
    if len(key) > 63 {
        return fmt.Errorf("label key must be no more than 63 characters")
    }
    
    validKey := regexp.MustCompile(`^[a-zA-Z0-9]([-_.a-zA-Z0-9]*[a-zA-Z0-9])?$`)
    if !validKey.MatchString(key) {
        return fmt.Errorf("invalid label key: %s", key)
    }
    
    return nil
}

// ValidateLabelValue validates a label value
func ValidateLabelValue(value string) error {
    if len(value) > 63 {
        return fmt.Errorf("label value must be no more than 63 characters")
    }
    
    if value == "" {
        return nil // Empty value is allowed
    }
    
    validValue := regexp.MustCompile(`^[a-zA-Z0-9]([-_.a-zA-Z0-9]*[a-zA-Z0-9])?$`)
    if !validValue.MatchString(value) {
        return fmt.Errorf("invalid label value: %s", value)
    }
    
    return nil
}
