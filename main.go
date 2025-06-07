package main

import (
	"os"

	"github.com/pascal71/lhcli/cmd"
)

func main() {
	cmd.SetBuildInfo(Version, BuildDate)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
