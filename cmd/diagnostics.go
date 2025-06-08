package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// diagnosticsCmd shows basic Longhorn diagnostic information.
var diagnosticsCmd = &cobra.Command{
	Use:     "diagnostics",
	Aliases: []string{"diag"},
	Short:   "Show Longhorn diagnostic information",
	RunE:    runDiagnostics,
}

func init() {
	rootCmd.AddCommand(diagnosticsCmd)
}

func runDiagnostics(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	// Fetch Longhorn version from settings if available
	versionSetting, err := c.Settings().Get("current-longhorn-version")
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to get version: %v\n", err)
	}

	engineSetting, err := c.Settings().Get("default-engine-image")
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to get default engine image: %v\n", err)
	}

	fmt.Println("Longhorn Diagnostics")
	if versionSetting != nil {
		fmt.Printf("Version: %s\n", versionSetting.Value)
	}
	if engineSetting != nil {
		fmt.Printf("Default Engine Image: %s\n", engineSetting.Value)
	}

	// List engine images to show active/default
	engineImages, err := c.EngineImages().List()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to list engine images: %v\n", err)
		return nil
	}

	fmt.Println()
	fmt.Println("Engine Images:")
	for _, ei := range engineImages {
		mark := ""
		if ei.Default {
			mark = " (default)"
		}
		fmt.Printf("- %s%s\t%s\n", ei.Name, mark, ei.Image)
	}

	return nil
}
