// cmd/node.go
package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/formatter"
	"github.com/pascal71/lhcli/pkg/utils"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage Longhorn nodes",
	Long:  `Manage Longhorn nodes including scheduling, disk management, and tagging.`,
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes",
	Long:  `List all Longhorn nodes in the cluster.`,
	RunE:  runNodeList,
}

var nodeGetCmd = &cobra.Command{
	Use:   "get [node-name]",
	Short: "Get node details",
	Long:  `Get detailed information about a specific node.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runNodeGet,
}

var nodeSchedulingCmd = &cobra.Command{
	Use:   "scheduling",
	Short: "Manage node scheduling",
	Long:  `Enable or disable scheduling on a node.`,
}

var nodeSchedulingEnableCmd = &cobra.Command{
	Use:   "enable [node-name]",
	Short: "Enable scheduling on a node",
	Args:  cobra.ExactArgs(1),
	RunE:  runNodeSchedulingEnable,
}

var nodeSchedulingDisableCmd = &cobra.Command{
	Use:   "disable [node-name]",
	Short: "Disable scheduling on a node",
	Args:  cobra.ExactArgs(1),
	RunE:  runNodeSchedulingDisable,
}

var nodeEvictCmd = &cobra.Command{
	Use:   "evict [node-name]",
	Short: "Evict all replicas from a node",
	Long:  `Request eviction of all replicas from a node. This will move all replicas to other nodes.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runNodeEvict,
}

var nodeTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage node tags",
	Long:  `Add or remove tags from nodes.`,
}

var nodeTagAddCmd = &cobra.Command{
	Use:   "add [node-name] [tag]",
	Short: "Add a tag to a node",
	Args:  cobra.ExactArgs(2),
	RunE:  runNodeTagAdd,
}

var nodeTagRemoveCmd = &cobra.Command{
	Use:   "remove [node-name] [tag]",
	Short: "Remove a tag from a node",
	Args:  cobra.ExactArgs(2),
	RunE:  runNodeTagRemove,
}

var nodeDiskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Manage node disks",
	Long:  `Add, remove, or update disks on nodes.`,
}

var nodeDiskAddCmd = &cobra.Command{
	Use:   "add [node-name]",
	Short: "Add a disk to a node",
	Args:  cobra.ExactArgs(1),
	RunE:  runNodeDiskAdd,
}

var nodeDiskRemoveCmd = &cobra.Command{
	Use:   "remove [node-name] [disk-id]",
	Short: "Remove a disk from a node",
	Args:  cobra.ExactArgs(2),
	RunE:  runNodeDiskRemove,
}

var nodeDiskUpdateCmd = &cobra.Command{
	Use:   "update [node-name] [disk-id]",
	Short: "Update disk configuration",
	Args:  cobra.ExactArgs(2),
	RunE:  runNodeDiskUpdate,
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// Add subcommands
	nodeCmd.AddCommand(nodeListCmd)
	nodeCmd.AddCommand(nodeGetCmd)
	nodeCmd.AddCommand(nodeSchedulingCmd)
	nodeCmd.AddCommand(nodeEvictCmd)
	nodeCmd.AddCommand(nodeTagCmd)
	nodeCmd.AddCommand(nodeDiskCmd)

	// Scheduling subcommands
	nodeSchedulingCmd.AddCommand(nodeSchedulingEnableCmd)
	nodeSchedulingCmd.AddCommand(nodeSchedulingDisableCmd)

	// Tag subcommands
	nodeTagCmd.AddCommand(nodeTagAddCmd)
	nodeTagCmd.AddCommand(nodeTagRemoveCmd)

	// Disk subcommands
	nodeDiskCmd.AddCommand(nodeDiskAddCmd)
	nodeDiskCmd.AddCommand(nodeDiskRemoveCmd)
	nodeDiskCmd.AddCommand(nodeDiskUpdateCmd)

	// Node evict flags
	nodeEvictCmd.Flags().Bool("force", false, "Force eviction without confirmation")

	// Disk add flags
	nodeDiskAddCmd.Flags().String("path", "", "Disk mount path")
	nodeDiskAddCmd.Flags().String("storage-reserved", "0", "Storage reserved (e.g., 10Gi)")
	nodeDiskAddCmd.Flags().StringSlice("tags", []string{}, "Disk tags")
	nodeDiskAddCmd.MarkFlagRequired("path")

	// Disk update flags
	nodeDiskUpdateCmd.Flags().StringSlice("tags", []string{}, "Update disk tags")
	nodeDiskUpdateCmd.Flags().Bool("allow-scheduling", true, "Allow scheduling on this disk")
	nodeDiskUpdateCmd.Flags().String("storage-reserved", "", "Update storage reserved")
}

func runNodeList(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	nodes, err := c.Nodes().List()
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(nodes)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(nodes)
	case "wide":
		return printNodesWide(nodes)
	default:
		return printNodesTable(nodes)
	}
}

func runNodeGet(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	c, err := getClient()
	if err != nil {
		return err
	}

	node, err := c.Nodes().Get(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(node)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(node)
	default:
		return printNodeDetails(node)
	}
}

func runNodeSchedulingEnable(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().EnableScheduling(nodeName); err != nil {
		return fmt.Errorf("failed to enable scheduling: %w", err)
	}

	fmt.Printf("✓ Scheduling enabled on node %s\n", nodeName)
	return nil
}

func runNodeSchedulingDisable(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().DisableScheduling(nodeName); err != nil {
		return fmt.Errorf("failed to disable scheduling: %w", err)
	}

	fmt.Printf("✓ Scheduling disabled on node %s\n", nodeName)
	return nil
}

func runNodeEvict(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force &&
		!utils.Confirm(
			fmt.Sprintf("Are you sure you want to evict all replicas from node %s?", nodeName),
		) {
		fmt.Println("Eviction cancelled")
		return nil
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().EvictNode(nodeName); err != nil {
		return fmt.Errorf("failed to evict node: %w", err)
	}

	fmt.Printf("✓ Eviction requested for node %s\n", nodeName)
	return nil
}

func runNodeTagAdd(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	tag := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().AddNodeTag(nodeName, tag); err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}

	fmt.Printf("✓ Tag '%s' added to node %s\n", tag, nodeName)
	return nil
}

func runNodeTagRemove(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	tag := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().RemoveNodeTag(nodeName, tag); err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	fmt.Printf("✓ Tag '%s' removed from node %s\n", tag, nodeName)
	return nil
}

func runNodeDiskAdd(cmd *cobra.Command, args []string) error {
	nodeName := args[0]

	path, _ := cmd.Flags().GetString("path")
	storageReserved, _ := cmd.Flags().GetString("storage-reserved")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Convert storage reserved to bytes
	reservedBytes, err := utils.ParseSize(storageReserved)
	if err != nil {
		return fmt.Errorf("invalid storage-reserved: %w", err)
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	disk := client.DiskUpdate{
		Path:            path,
		StorageReserved: reservedBytes,
		Tags:            tags,
	}

	if err := c.Nodes().AddDisk(nodeName, disk); err != nil {
		return fmt.Errorf("failed to add disk: %w", err)
	}

	fmt.Printf("✓ Disk %s added to node %s\n", path, nodeName)
	return nil
}

func runNodeDiskRemove(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	diskID := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Nodes().RemoveDisk(nodeName, diskID); err != nil {
		return fmt.Errorf("failed to remove disk: %w", err)
	}

	fmt.Printf("✓ Disk %s removed from node %s\n", diskID, nodeName)
	return nil
}

func runNodeDiskUpdate(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	diskID := args[1]

	tags, _ := cmd.Flags().GetStringSlice("tags")

	c, err := getClient()
	if err != nil {
		return err
	}

	if len(tags) > 0 {
		if err := c.Nodes().UpdateDiskTags(nodeName, diskID, tags); err != nil {
			return fmt.Errorf("failed to update disk tags: %w", err)
		}
	}

	fmt.Printf("✓ Disk %s updated on node %s\n", diskID, nodeName)
	return nil
}

// Helper functions for printing

func printNodesTable(nodes []client.Node) error {
	// Use formatter package instead
	headers := []string{"NAME", "STATUS", "SCHEDULABLE", "AGE", "REGION", "ZONE"}
	formatter := formatter.NewTableFormatter(headers)

	for _, node := range nodes {
		status := getNodeStatus(node)
		schedulable := "true"
		if !node.AllowScheduling {
			schedulable = "false"
		}

		formatter.AddRow([]string{
			node.Name,
			status,
			schedulable,
			"-", // TODO: Calculate age
			node.Region,
			node.Zone,
		})
	}

	return formatter.Format(nil)
}

func printNodesWide(nodes []client.Node) error {
	headers := []string{"NAME", "STATUS", "SCHEDULABLE", "DISKS", "REPLICAS", "TAGS", "AGE"}
	formatter := formatter.NewTableFormatter(headers)

	for _, node := range nodes {
		status := getNodeStatus(node)
		schedulable := "true"
		if !node.AllowScheduling {
			schedulable = "false"
		}

		diskCount := len(node.Disks)
		replicaCount := 0
		for _, disk := range node.Disks {
			replicaCount += len(disk.ScheduledReplica)
		}

		tags := strings.Join(node.Tags, ",")
		if tags == "" {
			tags = "<none>"
		}

		formatter.AddRow([]string{
			node.Name,
			status,
			schedulable,
			fmt.Sprintf("%d", diskCount),
			fmt.Sprintf("%d", replicaCount),
			tags,
			"-", // TODO: Calculate age
		})
	}

	return formatter.Format(nil)
}

func printNodeDetails(node *client.Node) error {
	fmt.Printf("Name:              %s\n", node.Name)
	fmt.Printf("Address:           %s\n", node.Address)
	fmt.Printf("Status:            %s\n", getNodeStatus(*node))
	fmt.Printf("Schedulable:       %v\n", node.AllowScheduling)
	fmt.Printf("Eviction Requested: %v\n", node.EvictionRequested)
	fmt.Printf("Region:            %s\n", node.Region)
	fmt.Printf("Zone:              %s\n", node.Zone)

	if len(node.Tags) > 0 {
		fmt.Printf("Tags:              %s\n", strings.Join(node.Tags, ", "))
	} else {
		fmt.Printf("Tags:              <none>\n")
	}

	fmt.Println("\nConditions:")
	for name, condition := range node.Conditions {
		statusColor := color.New(color.FgGreen)
		if condition.Status != "True" {
			statusColor = color.New(color.FgRed)
		}
		fmt.Printf("  %s: %s\n", name, statusColor.Sprint(condition.Status))
		if condition.Message != "" {
			fmt.Printf("    Message: %s\n", condition.Message)
		}
	}

	fmt.Println("\nDisks:")
	if len(node.Disks) == 0 {
		fmt.Println("  <none>")
	} else {
		for diskID, disk := range node.Disks {
			fmt.Printf("  %s:\n", diskID)
			fmt.Printf("    Path:            %s\n", disk.Path)
			fmt.Printf("    Type:            %s\n", disk.DiskType)
			fmt.Printf("    Schedulable:     %v\n", disk.AllowScheduling)
			fmt.Printf("    Storage Maximum: %s\n", utils.FormatSize(disk.StorageMaximum))
			fmt.Printf("    Storage Available: %s\n", utils.FormatSize(disk.StorageAvailable))
			fmt.Printf("    Storage Reserved: %s\n", utils.FormatSize(disk.StorageReserved))
			fmt.Printf("    Storage Scheduled: %s\n", utils.FormatSize(disk.StorageScheduled))
			if len(disk.Tags) > 0 {
				fmt.Printf("    Tags: %s\n", strings.Join(disk.Tags, ", "))
			}
		}
	}

	return nil
}

func getNodeStatus(node client.Node) string {
	// Check if node is ready based on conditions
	if readyCondition, exists := node.Conditions["Ready"]; exists {
		if readyCondition.Status == "True" {
			return "Ready"
		}
		return "NotReady"
	}
	return "Unknown"
}
