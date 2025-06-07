// pkg/client/debug.go
package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// EnableDebug enables debug output
var EnableDebug = os.Getenv("LHCLI_DEBUG") != ""

// debugLog prints debug information if debugging is enabled
func debugLog(format string, args ...interface{}) {
	if EnableDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// debugResponse prints response details for debugging
func debugResponse(resp *http.Response) {
	if !EnableDebug {
		return
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Response Status: %s\n", resp.Status)
	fmt.Fprintf(os.Stderr, "[DEBUG] Response Headers:\n")
	for k, v := range resp.Header {
		fmt.Fprintf(os.Stderr, "[DEBUG]   %s: %v\n", k, v)
	}

	// Read and print body preview
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Failed to read body: %v\n", err)
		return
	}

	// Restore body for later use
	resp.Body = io.NopCloser(bytes.NewReader(body))

	preview := string(body)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Response Body Preview:\n%s\n", preview)
}
