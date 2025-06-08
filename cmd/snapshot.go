package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/formatter"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage volume snapshots",
	Long:  `Manage Longhorn volume snapshots including create, delete, and revert operations.`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [volume-name]",
	Short: "Create a snapshot",
	Long:  `Create a snapshot of the specified volume.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotCreate,
}

var snapshotListCmd = &cobra.Command{
	Use:   "list [volume-name]",
	Short: "List snapshots",
	Long:  `List all snapshots for a specific volume.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshotList,
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [volume-name] [snapshot-name]",
	Short: "Delete a snapshot",
	Args:  cobra.ExactArgs(2),
	RunE:  runSnapshotDelete,
}

func runSnapshotCreate(cmd *cobra.Command, args []string) error {
	volumeName := args[0]
	name, _ := cmd.Flags().GetString("name")
	labels, _ := cmd.Flags().GetStringToString("labels")

	c, err := getClient()
	if err != nil {
		return err
	}

	input := &client.SnapshotCreateInput{
		Name:   name,
		Labels: labels,
	}

	if _, err := c.Snapshots().Create(volumeName, input); err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	fmt.Printf("✓ Snapshot %s created for volume %s\n", name, volumeName)
	return nil
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	volumeName := args[0]

	c, err := getClient()
	if err != nil {
		return err
	}

	snapshots, err := c.Snapshots().List(volumeName)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(snapshots)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(snapshots)
	default:
		for _, s := range snapshots {
			fmt.Printf("%s\t%s\n", s.Name, s.Created)
		}
		return nil
	}
}

func runSnapshotDelete(cmd *cobra.Command, args []string) error {
	volumeName := args[0]
	snapshotName := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Snapshots().Delete(volumeName, snapshotName); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	fmt.Printf("✓ Snapshot %s deleted from volume %s\n", snapshotName, volumeName)
	return nil
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)

	// Snapshot create flags
	snapshotCreateCmd.Flags().String("name", "", "Snapshot name")
	snapshotCreateCmd.MarkFlagRequired("name")
	snapshotCreateCmd.Flags().StringToString("labels", nil, "Labels for the snapshot")
}
