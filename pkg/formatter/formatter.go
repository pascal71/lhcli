// pkg/formatter/formatter.go
package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Formatter interface for different output formats
type Formatter interface {
	Format(data interface{}) error
}

// FormatType represents the output format type
type FormatType string

const (
	FormatTable FormatType = "table"
	FormatJSON  FormatType = "json"
	FormatYAML  FormatType = "yaml"
	FormatWide  FormatType = "wide"
)

// TableFormatter formats output as a table
type TableFormatter struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(headers []string) *TableFormatter {
	return &TableFormatter{
		headers: headers,
		writer:  os.Stdout,
	}
}

// NewTableFormatterWithWriter creates a new table formatter with custom writer
func NewTableFormatterWithWriter(headers []string, writer io.Writer) *TableFormatter {
	return &TableFormatter{
		headers: headers,
		writer:  writer,
	}
}

// AddRow adds a row to the table
func (t *TableFormatter) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// AddRows adds multiple rows to the table
func (t *TableFormatter) AddRows(rows [][]string) {
	t.rows = append(t.rows, rows...)
}

// Format outputs the table
func (t *TableFormatter) Format(data interface{}) error {
	// If data is provided, try to extract rows from it
	if data != nil && len(t.rows) == 0 {
		if err := t.extractRows(data); err != nil {
			return err
		}
	}

	// Use standard library's tabwriter for consistent column alignment
	w := tabwriter.NewWriter(t.writer, 0, 0, 3, ' ', 0)

	// Print headers
	if len(t.headers) > 0 {
		fmt.Fprintln(w, strings.Join(t.headers, "\t"))
	}

	// Print rows
	for _, row := range t.rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

// extractRows attempts to extract rows from the provided data
func (t *TableFormatter) extractRows(data interface{}) error {
	// Use reflection to handle different data types
	val := reflect.ValueOf(data)

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			row := t.extractRowFromStruct(item)
			if row != nil {
				t.rows = append(t.rows, row)
			}
		}
	case reflect.Struct:
		row := t.extractRowFromStruct(val)
		if row != nil {
			t.rows = append(t.rows, row)
		}
	default:
		return fmt.Errorf("unsupported data type for table formatting: %T", data)
	}

	return nil
}

// extractRowFromStruct extracts a row from a struct based on headers
func (t *TableFormatter) extractRowFromStruct(val reflect.Value) []string {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	row := make([]string, len(t.headers))
	typ := val.Type()

	for i, header := range t.headers {
		// Try to find field by name (case-insensitive)
		fieldName := t.findFieldName(typ, header)
		if fieldName != "" {
			field := val.FieldByName(fieldName)
			if field.IsValid() {
				row[i] = t.formatValue(field)
			}
		}
	}

	return row
}

// findFieldName finds the actual field name in a struct that matches the header
func (t *TableFormatter) findFieldName(typ reflect.Type, header string) string {
	header = strings.ToLower(strings.ReplaceAll(header, " ", ""))

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)

		if fieldName == header {
			return field.Name
		}

		// Check JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if strings.ToLower(tagName) == header {
				return field.Name
			}
		}
	}

	return ""
}

// formatValue formats a reflect.Value to string
func (t *TableFormatter) formatValue(val reflect.Value) string {
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", val.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", val.Bool())
	case reflect.Slice, reflect.Array:
		// Handle string slices specially
		if val.Type().Elem().Kind() == reflect.String {
			items := make([]string, val.Len())
			for i := 0; i < val.Len(); i++ {
				items[i] = val.Index(i).String()
			}
			return strings.Join(items, ",")
		}
		return fmt.Sprintf("[%d items]", val.Len())
	case reflect.Map:
		return fmt.Sprintf("[%d items]", val.Len())
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	pretty bool
	writer io.Writer
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{
		pretty: pretty,
		writer: os.Stdout,
	}
}

// NewJSONFormatterWithWriter creates a new JSON formatter with custom writer
func NewJSONFormatterWithWriter(pretty bool, writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		pretty: pretty,
		writer: writer,
	}
}

// Format outputs as JSON
func (j *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(j.writer)
	if j.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// YAMLFormatter formats output as YAML
type YAMLFormatter struct {
	writer io.Writer
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{
		writer: os.Stdout,
	}
}

// NewYAMLFormatterWithWriter creates a new YAML formatter with custom writer
func NewYAMLFormatterWithWriter(writer io.Writer) *YAMLFormatter {
	return &YAMLFormatter{
		writer: writer,
	}
}

// Format outputs as YAML
func (y *YAMLFormatter) Format(data interface{}) error {
	encoder := yaml.NewEncoder(y.writer)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

// GetFormatter returns the appropriate formatter based on format string
func GetFormatter(format string) (Formatter, error) {
	switch FormatType(strings.ToLower(format)) {
	case FormatJSON:
		return NewJSONFormatter(true), nil
	case FormatYAML:
		return NewYAMLFormatter(), nil
	case FormatTable:
		return nil, fmt.Errorf("table formatter requires headers to be specified")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetFormatterWithHeaders returns a formatter with headers for table format
func GetFormatterWithHeaders(format string, headers []string) (Formatter, error) {
	switch FormatType(strings.ToLower(format)) {
	case FormatJSON:
		return NewJSONFormatter(true), nil
	case FormatYAML:
		return NewYAMLFormatter(), nil
	case FormatTable, FormatWide:
		return NewTableFormatter(headers), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// PrintError prints an error message in a formatted way
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("⚠ %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("ℹ %s\n", message)
}
