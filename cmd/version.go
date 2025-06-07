package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   string
	buildDate string
)

// SetBuildInfo sets the build information for the version command.
func SetBuildInfo(v, b string) {
	version = v
	buildDate = b
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version)
		if buildDate != "" {
			fmt.Printf("Build Date: %s\n", buildDate)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
