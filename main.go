package main

import (
	"os"

	"github.com/pascal71/lhcli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
