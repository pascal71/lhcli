package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pascal71/lhcli/pkg/utils"
)

// troubleshootCmd represents the troubleshoot command
var troubleshootCmd = &cobra.Command{
	Use:   "troubleshoot",
	Short: "Discover common Longhorn issues",
	Long:  `Run a set of checks to find potential problems like orphaned replicas and resource shortages.`,
	RunE:  runTroubleshoot,
}

func init() {
	rootCmd.AddCommand(troubleshootCmd)
}

func runTroubleshoot(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	issues := []string{}

	// Map of volumes for quick lookup
	volumes, err := c.Volumes().List()
	if err == nil {
		volMap := make(map[string]struct{})
		for _, v := range volumes {
			volMap[v.Name] = struct{}{}
		}

		replicas, rerr := c.Replicas().List()
		if rerr == nil {
			for _, r := range replicas {
				if _, exists := volMap[r.VolumeName]; !exists {
					issues = append(issues, fmt.Sprintf("orphaned replica %s on node %s", r.Name, r.NodeID))
				}
			}
		} else {
			issues = append(issues, fmt.Sprintf("failed to list replicas: %v", rerr))
		}
	} else {
		issues = append(issues, fmt.Sprintf("failed to list volumes: %v", err))
	}

	// Check node and disk status
	nodes, nerr := c.Nodes().List()
	if nerr == nil {
		for _, n := range nodes {
			if getNodeStatus(n) != "Ready" {
				issues = append(issues, fmt.Sprintf("node %s is not ready", n.Name))
			}
			if !n.AllowScheduling {
				issues = append(issues, fmt.Sprintf("node %s scheduling disabled", n.Name))
			}
			for id, d := range n.Disks {
				if d.StorageMaximum > 0 && d.StorageAvailable*10 < d.StorageMaximum {
					issues = append(issues,
						fmt.Sprintf(
							"low disk space on %s[%s]: %s free of %s",
							n.Name,
							id,
							utils.FormatSize(d.StorageAvailable),
							utils.FormatSize(d.StorageMaximum),
						))
				}
			}
		}
	} else {
		issues = append(issues, fmt.Sprintf("failed to list nodes: %v", nerr))
	}

	if len(issues) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No issues detected")
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Potential issues:")
	for _, issue := range issues {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", issue)
	}
	return nil
}
