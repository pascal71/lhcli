// cmd/replica.go
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/formatter"
	"github.com/pascal71/lhcli/pkg/utils"
)

var replicaCmd = &cobra.Command{
	Use:   "replica",
	Short: "Manage Longhorn replicas",
	Long:  `Manage Longhorn volume replicas including list, get, and delete operations.`,
}

var replicaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all replicas",
	Long:  `List all Longhorn replicas in the specified namespace.`,
	RunE:  runReplicaList,
}

var replicaGetCmd = &cobra.Command{
	Use:   "get [replica-name]",
	Short: "Get replica details",
	Long:  `Get detailed information about a specific replica.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runReplicaGet,
}

var replicaDeleteCmd = &cobra.Command{
	Use:   "delete [replica-name]",
	Short: "Delete a replica",
	Long:  `Delete a Longhorn replica. Use with caution as this removes a data copy.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runReplicaDelete,
}

var (
	volumeFilter string
	nodeFilter   string
)

func init() {
	rootCmd.AddCommand(replicaCmd)
	replicaCmd.AddCommand(replicaListCmd)
	replicaCmd.AddCommand(replicaGetCmd)
	replicaCmd.AddCommand(replicaDeleteCmd)

	// Replica list flags
	replicaListCmd.Flags().StringVar(&volumeFilter, "volume", "", "Filter replicas by volume name")
	replicaListCmd.Flags().StringVar(&nodeFilter, "node", "", "Filter replicas by node")
	replicaListCmd.Flags().
		Bool("full-ids", false, "Show full replica names and disk IDs without abbreviation")

	// Replica get flags
	replicaGetCmd.Flags().
		Bool("full-ids", false, "Show full disk IDs without abbreviation")

	// Replica delete flags
	replicaDeleteCmd.Flags().Bool("force", false, "Force delete without confirmation")
}

func runReplicaList(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	replicas, err := c.Replicas().List()
	if err != nil {
		return fmt.Errorf("failed to list replicas: %w", err)
	}

	// Apply filters
	var filteredReplicas []client.Replica
	for _, replica := range replicas {
		if volumeFilter != "" && replica.VolumeName != volumeFilter {
			continue
		}
		if nodeFilter != "" && replica.NodeID != nodeFilter {
			continue
		}
		filteredReplicas = append(filteredReplicas, replica)
	}

	// Get the showFullIDs flag value
	showFullIDs, _ := cmd.Flags().GetBool("full-ids")

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(filteredReplicas)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(filteredReplicas)
	case "wide":
		return printReplicasWide(filteredReplicas, showFullIDs)
	default:
		return printReplicasTable(filteredReplicas, showFullIDs)
	}
}

func runReplicaGet(cmd *cobra.Command, args []string) error {
	replicaName := args[0]

	c, err := getClient()
	if err != nil {
		return err
	}

	replica, err := c.Replicas().Get(replicaName)
	if err != nil {
		return fmt.Errorf("failed to get replica: %w", err)
	}

	// Get the showFullIDs flag value
	showFullIDs, _ := cmd.Flags().GetBool("full-ids")

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(replica)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(replica)
	default:
		return printReplicaDetails(replica, showFullIDs)
	}
}

func runReplicaDelete(cmd *cobra.Command, args []string) error {
	replicaName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force &&
		!utils.Confirm(fmt.Sprintf("Are you sure you want to delete replica %s?", replicaName)) {
		fmt.Println("Deletion cancelled")
		return nil
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	// Get replica details first to show what we're deleting
	replica, err := c.Replicas().Get(replicaName)
	if err != nil {
		return fmt.Errorf("failed to get replica details: %w", err)
	}

	fmt.Printf("Deleting replica %s from volume %s on node %s...\n",
		replicaName, replica.VolumeName, replica.NodeID)

	if err := c.Replicas().Delete(replicaName); err != nil {
		return fmt.Errorf("failed to delete replica: %w", err)
	}

	fmt.Printf("✓ Replica %s deleted successfully\n", replicaName)
	fmt.Println(
		"Note: The volume will rebuild a new replica if numberOfReplicas is higher than remaining replicas",
	)

	return nil
}

// Helper functions

// getReplicaSize returns the size from either VolumeSize or SpecSize fields
func getReplicaSize(replica *client.Replica) string {
	// Try VolumeSize first (Kubernetes CRD field)
	if replica.VolumeSize != "" {
		return replica.VolumeSize
	}
	// Fall back to SpecSize (potential API field)
	if replica.SpecSize != "" {
		return replica.SpecSize
	}
	return ""
}

// getReplicaState returns the state from either CurrentState or Mode fields
func getReplicaState(replica *client.Replica) string {
	// Try CurrentState first (Kubernetes CRD field)
	if replica.CurrentState != "" {
		return replica.CurrentState
	}
	// Fall back to Mode (potential API field)
	if replica.Mode != "" {
		return replica.Mode
	}
	return "unknown"
}

// Printing functions

func printReplicasTable(replicas []client.Replica, showFullIDs bool) error {
	headers := []string{"NAME", "VOLUME", "NODE", "DISK PATH", "SIZE", "STATE"}
	formatter := formatter.NewTableFormatter(headers)

	for _, replica := range replicas {
		name := replica.Name
		if !showFullIDs && len(name) > 40 {
			name = name[:37] + "..."
		}

		volumeName := replica.VolumeName
		if !showFullIDs && len(volumeName) > 30 {
			volumeName = volumeName[:27] + "..."
		}

		size := getReplicaSize(&replica)
		if size != "" && size != "0" {
			if sizeInt, err := strconv.ParseInt(size, 10, 64); err == nil {
				size = utils.FormatSize(sizeInt)
			}
		}

		formatter.AddRow([]string{
			name,
			volumeName,
			replica.NodeID,
			replica.DiskPath,
			size,
			getReplicaState(&replica),
		})
	}

	return formatter.Format(nil)
}

func printReplicasWide(replicas []client.Replica, showFullIDs bool) error {
	headers := []string{
		"NAME",
		"VOLUME",
		"NODE",
		"DISK ID",
		"DISK PATH",
		"SIZE",
		"STATE",
		"RUNNING",
		"IP",
	}
	formatter := formatter.NewTableFormatter(headers)

	for _, replica := range replicas {
		size := getReplicaSize(&replica)
		if size != "" && size != "0" {
			if sizeInt, err := strconv.ParseInt(size, 10, 64); err == nil {
				size = utils.FormatSize(sizeInt)
			}
		}

		diskID := replica.DiskID
		if !showFullIDs && len(diskID) > 20 {
			diskID = diskID[:17] + "..."
		}

		formatter.AddRow([]string{
			replica.Name,
			replica.VolumeName,
			replica.NodeID,
			diskID,
			replica.DiskPath,
			size,
			getReplicaState(&replica),
			fmt.Sprintf("%v", replica.Running),
			replica.IP,
		})
	}

	return formatter.Format(nil)
}

func printReplicaDetails(replica *client.Replica, showFullIDs bool) error {
	fmt.Printf("Name:              %s\n", replica.Name)
	fmt.Printf("Volume:            %s\n", replica.VolumeName)
	fmt.Printf("Node:              %s\n", replica.NodeID)

	diskID := replica.DiskID
	if !showFullIDs && len(diskID) > 60 {
		diskID = diskID[:57] + "..."
	}
	fmt.Printf("Disk ID:           %s\n", diskID)

	fmt.Printf("Disk Path:         %s\n", replica.DiskPath)
	fmt.Printf("Data Path:         %s\n", replica.DataPath)
	fmt.Printf("State:             %s\n", getReplicaState(replica))
	fmt.Printf("Running:           %v\n", replica.Running)

	// Display size information
	replicaSize := getReplicaSize(replica)
	if replicaSize != "" && replicaSize != "0" {
		size := replicaSize
		if sizeInt, err := strconv.ParseInt(replicaSize, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}
		fmt.Printf("Size:              %s\n", size)
	}

	// Display actual size if available
	if replica.ActualSize != "" && replica.ActualSize != "0" {
		size := replica.ActualSize
		if sizeInt, err := strconv.ParseInt(replica.ActualSize, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}
		fmt.Printf("Size (Actual):     %s\n", size)
	}

	if replica.IP != "" {
		fmt.Printf("IP:                %s\n", replica.IP)
		fmt.Printf("Port:              %d\n", replica.Port)
	}

	if replica.InstanceManager != "" {
		fmt.Printf("Instance Manager:  %s\n", replica.InstanceManager)
	}

	if replica.Image != "" {
		fmt.Printf("Image:             %s\n", replica.Image)
	}

	if replica.CurrentImage != "" && replica.CurrentImage != replica.Image {
		fmt.Printf("Current Image:     %s\n", replica.CurrentImage)
	}

	if replica.FailedAt != "" {
		fmt.Printf("Failed At:         %s\n", replica.FailedAt)
	}

	if replica.DataEngine != "" {
		fmt.Printf("Data Engine:       %s\n", replica.DataEngine)
	}

	return nil
}
