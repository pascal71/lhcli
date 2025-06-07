// pkg/formatter/formatter_test.go
package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pascal71/lhcli/pkg/utils"
)

// Test data structures
type TestNode struct {
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	AllowScheduling bool      `json:"allowScheduling"`
	Created         time.Time `json:"created"`
	Tags            []string  `json:"tags"`
}

func TestTableFormatter(t *testing.T) {
	// Create test data
	nodes := []TestNode{
		{
			Name:            "node-1",
			Status:          "Ready",
			AllowScheduling: true,
			Created:         time.Now().Add(-24 * time.Hour),
			Tags:            []string{"ssd", "fast"},
		},
		{
			Name:            "node-2",
			Status:          "NotReady",
			AllowScheduling: false,
			Created:         time.Now().Add(-48 * time.Hour),
			Tags:            []string{"hdd"},
		},
	}

	// Test with manual row addition
	t.Run("ManualRows", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewTableFormatterWithWriter(
			[]string{"NAME", "STATUS", "SCHEDULABLE", "TAGS"},
			&buf,
		)

		for _, node := range nodes {
			formatter.AddRow([]string{
				node.Name,
				node.Status,
				FormatBool(node.AllowScheduling),
				FormatList(node.Tags),
			})
		}

		err := formatter.Format(nil)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "node-1") || !strings.Contains(output, "Ready") {
			t.Errorf("Expected output to contain node-1 and Ready, got: %s", output)
		}
	})

	// Test with automatic extraction
	t.Run("AutoExtraction", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewTableFormatterWithWriter(
			[]string{"name", "status", "allowscheduling"},
			&buf,
		)

		err := formatter.Format(nodes)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "node-1") || !strings.Contains(output, "Ready") {
			t.Errorf("Expected output to contain node-1 and Ready, got: %s", output)
		}
	})
}

func TestJSONFormatter(t *testing.T) {
	node := TestNode{
		Name:            "test-node",
		Status:          "Ready",
		AllowScheduling: true,
		Created:         time.Now(),
		Tags:            []string{"test"},
	}

	var buf bytes.Buffer
	formatter := NewJSONFormatterWithWriter(true, &buf)

	err := formatter.Format(node)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify it's valid JSON
	var result TestNode
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if result.Name != node.Name {
		t.Errorf("Expected name %s, got %s", node.Name, result.Name)
	}
}

func TestYAMLFormatter(t *testing.T) {
	node := TestNode{
		Name:            "test-node",
		Status:          "Ready",
		AllowScheduling: true,
		Tags:            []string{"test"},
	}

	var buf bytes.Buffer
	formatter := NewYAMLFormatterWithWriter(&buf)

	err := formatter.Format(node)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name: test-node") {
		t.Errorf("Expected YAML to contain 'name: test-node', got: %s", output)
	}
}

func TestHelpers(t *testing.T) {
	// Test FormatDuration
	t.Run("FormatDuration", func(t *testing.T) {
		tests := []struct {
			duration time.Duration
			expected string
		}{
			{0, "0s"},
			{30 * time.Second, "30s"},
			{90 * time.Second, "1m30s"},
			{25 * time.Hour, "1d1h"},
			{49 * time.Hour, "2d1h"},
		}

		for _, test := range tests {
			result := FormatDuration(test.duration)
			if result != test.expected {
				t.Errorf("FormatDuration(%v) = %s, want %s", test.duration, result, test.expected)
			}
		}
	})

	// Test TruncateString
	t.Run("TruncateString", func(t *testing.T) {
		tests := []struct {
			input    string
			maxLen   int
			expected string
		}{
			{"hello", 10, "hello"},
			{"hello world", 8, "hello..."},
			{"hi", 2, "hi"},
			{"hello", 3, "hel"},
		}

		for _, test := range tests {
			result := utils.TruncateString(test.input, test.maxLen)
			if result != test.expected {
				t.Errorf("TruncateString(%s, %d) = %s, want %s",
					test.input, test.maxLen, result, test.expected)
			}
		}
	})
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status   string
		useColor bool
	}{
		{"Ready", true},
		{"Error", true},
		{"Pending", true},
		{"Unknown", true},
		{"Ready", false},
	}

	for _, test := range tests {
		result := FormatStatus(test.status, test.useColor)
		if test.useColor {
			// Should contain color codes
			if !strings.Contains(result, "\033[") {
				t.Errorf("Expected colored output for status %s", test.status)
			}
		} else {
			// Should not contain color codes
			if result != test.status {
				t.Errorf("Expected plain status %s, got %s", test.status, result)
			}
		}
	}
}
