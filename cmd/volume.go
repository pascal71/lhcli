// cmd/volume.go
package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/config"
	"github.com/pascal71/lhcli/pkg/formatter"
	"github.com/pascal71/lhcli/pkg/utils"
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage Longhorn volumes",
	Long:  `Manage Longhorn volumes including create, delete, list, and update operations.`,
}

var volumeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all volumes",
	Long:  `List all Longhorn volumes in the specified namespace.`,
	RunE:  runVolumeList,
}

var volumeCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new volume",
	Long:  `Create a new Longhorn volume with the specified configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeCreate,
}

var volumeDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a volume",
	Long:  `Delete a Longhorn volume.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeDelete,
}

var volumeGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get volume details",
	Long:  `Get detailed information about a specific volume.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeGet,
}

func init() {
	rootCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeCreateCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)
	volumeCmd.AddCommand(volumeGetCmd)

	// Volume create flags
	volumeCreateCmd.Flags().String("size", "10Gi", "Volume size")
	volumeCreateCmd.Flags().Int("replicas", 3, "Number of replicas")
	volumeCreateCmd.Flags().String("frontend", "blockdev", "Frontend type (blockdev|iscsi)")
	volumeCreateCmd.Flags().String("access-mode", "rwo", "Access mode (rwo|rwx)")
	volumeCreateCmd.Flags().StringSlice("node-selector", []string{}, "Node selector tags")
	volumeCreateCmd.Flags().StringSlice("disk-selector", []string{}, "Disk selector tags")
	volumeCreateCmd.Flags().StringToString("labels", nil, "Labels for the volume")

	// Volume delete flags
	volumeDeleteCmd.Flags().Bool("force", false, "Force delete")

	// Volume get flags
	volumeGetCmd.Flags().Bool("detailed", false, "Show detailed information")
}

func getClient() (*client.Client, error) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get current context
	ctx, err := cfg.GetContext(context)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Determine namespace to use
	ns := namespace
	if ns == "" {
		// Use namespace from context if available
		if ctx.Namespace != "" {
			ns = ctx.Namespace
		} else {
			// Default to longhorn-system
			ns = "longhorn-system"
		}
	}

	// Check auth type
	switch ctx.Auth.Type {
	case "kubeconfig":
		// Use kubeconfig for authentication
		kubeConfig := &client.KubeConfig{
			ConfigPath: ctx.Auth.Path,
			Context:    ctx.Auth.Context,
			Namespace:  ns,
		}
		return client.NewClientFromKubeconfig(kubeConfig)

	case "token":
		// Use direct connection with token
		clientConfig := &client.Config{
			Endpoint:  ctx.Endpoint,
			Namespace: ns,
			Token:     ctx.Auth.Token,
			Timeout:   30 * time.Second,
		}
		return client.NewClient(clientConfig)

	case "none", "":
		// Direct connection without auth
		clientConfig := &client.Config{
			Endpoint:  ctx.Endpoint,
			Namespace: ns,
			Timeout:   30 * time.Second,
		}
		return client.NewClient(clientConfig)

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", ctx.Auth.Type)
	}
}

func runVolumeList(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	volumes, err := c.Volumes().List()
	if err != nil {
		return fmt.Errorf("failed to list volumes: %w", err)
	}

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(volumes)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(volumes)
	case "wide":
		return printVolumesWide(volumes)
	default:
		return printVolumesTable(volumes)
	}
}

func runVolumeCreate(cmd *cobra.Command, args []string) error {
	volumeName := args[0]

	size, _ := cmd.Flags().GetString("size")
	replicas, _ := cmd.Flags().GetInt("replicas")
	frontend, _ := cmd.Flags().GetString("frontend")
	accessMode, _ := cmd.Flags().GetString("access-mode")
	nodeSelector, _ := cmd.Flags().GetStringSlice("node-selector")
	diskSelector, _ := cmd.Flags().GetStringSlice("disk-selector")
	labels, _ := cmd.Flags().GetStringToString("labels")

	c, err := getClient()
	if err != nil {
		return err
	}

	input := &client.VolumeCreateInput{
		Name:             volumeName,
		Size:             size,
		NumberOfReplicas: replicas,
		Frontend:         frontend,
		AccessMode:       accessMode,
		NodeSelector:     nodeSelector,
		DiskSelector:     diskSelector,
		Labels:           labels,
	}

	volume, err := c.Volumes().Create(input)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	fmt.Printf("✓ Volume %s created successfully\n", volume.Name)
	return nil
}

func runVolumeDelete(cmd *cobra.Command, args []string) error {
	volumeName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force &&
		!utils.Confirm(fmt.Sprintf("Are you sure you want to delete volume %s?", volumeName)) {
		fmt.Println("Deletion cancelled")
		return nil
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	if err := c.Volumes().Delete(volumeName); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	fmt.Printf("✓ Volume %s deleted successfully\n", volumeName)
	return nil
}

func runVolumeGet(cmd *cobra.Command, args []string) error {
	volumeName := args[0]
	detailed, _ := cmd.Flags().GetBool("detailed")

	c, err := getClient()
	if err != nil {
		return err
	}

	volume, err := c.Volumes().Get(volumeName)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(volume)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(volume)
	default:
		return printVolumeDetails(volume, detailed)
	}
}

// Helper functions for printing

func printVolumesTable(volumes []client.Volume) error {
	headers := []string{"NAME", "SIZE", "REPLICAS", "STATE", "ROBUSTNESS", "AGE"}
	formatter := formatter.NewTableFormatter(headers)

	for _, volume := range volumes {
		state := getVolumeState(volume)
		robustness := volume.Robustness
		if robustness == "" {
			robustness = "Unknown"
		}

		formatter.AddRow([]string{
			volume.Name,
			volume.Size,
			fmt.Sprintf("%d", volume.NumberOfReplicas),
			state,
			robustness,
			"-", // TODO: Calculate age
		})
	}

	return formatter.Format(nil)
}

func printVolumesWide(volumes []client.Volume) error {
	headers := []string{
		"NAME",
		"SIZE",
		"REPLICAS",
		"STATE",
		"ROBUSTNESS",
		"FRONTEND",
		"ACCESS",
		"CREATED",
	}
	formatter := formatter.NewTableFormatter(headers)

	for _, volume := range volumes {
		state := getVolumeState(volume)
		robustness := volume.Robustness
		if robustness == "" {
			robustness = "Unknown"
		}

		formatter.AddRow([]string{
			volume.Name,
			volume.Size,
			fmt.Sprintf("%d", volume.NumberOfReplicas),
			state,
			robustness,
			volume.Frontend,
			volume.AccessMode,
			volume.Created,
		})
	}

	return formatter.Format(nil)
}

func printVolumeDetails(volume *client.Volume, detailed bool) error {
	fmt.Printf("Name:              %s\n", volume.Name)
	fmt.Printf("Size:              %s\n", volume.Size)
	fmt.Printf("Number of Replicas: %d\n", volume.NumberOfReplicas)
	fmt.Printf("State:             %s\n", getVolumeState(*volume))
	fmt.Printf("Robustness:        %s\n", volume.Robustness)
	fmt.Printf("Frontend:          %s\n", volume.Frontend)
	fmt.Printf("Access Mode:       %s\n", volume.AccessMode)
	fmt.Printf("Migratable:        %v\n", volume.Migratable)
	fmt.Printf("Encrypted:         %v\n", volume.Encrypted)
	fmt.Printf("Created:           %s\n", volume.Created)

	if volume.LastBackup != "" {
		fmt.Printf("Last Backup:       %s at %s\n", volume.LastBackup, volume.LastBackupAt)
	}

	if len(volume.Labels) > 0 {
		fmt.Printf("Labels:            %s\n", formatter.FormatMap(volume.Labels))
	} else {
		fmt.Printf("Labels:            <none>\n")
	}

	if detailed || len(volume.Conditions) > 0 {
		fmt.Println("\nConditions:")
		for name, condition := range volume.Conditions {
			statusColor := color.New(color.FgGreen)
			if condition.Status != "True" {
				statusColor = color.New(color.FgRed)
			}
			fmt.Printf("  %s: %s\n", name, statusColor.Sprint(condition.Status))
			if condition.Message != "" {
				fmt.Printf("    Message: %s\n", condition.Message)
			}
		}
	}

	if detailed && len(volume.Replicas) > 0 {
		fmt.Println("\nReplicas:")
		for _, replica := range volume.Replicas {
			fmt.Printf("  %s:\n", replica.Name)
			fmt.Printf("    Node:   %s\n", replica.NodeID)
			fmt.Printf("    Mode:   %s\n", replica.Mode)
			if replica.DiskID != "" {
				fmt.Printf("    Disk:   %s\n", replica.DiskID)
			}
		}
	}

	return nil
}

func getVolumeState(volume client.Volume) string {
	// Check state field first
	if volume.State != "" {
		return volume.State
	}

	// Check conditions
	if scheduledCondition, exists := volume.Conditions["Scheduled"]; exists {
		if scheduledCondition.Status == "True" {
			return "Attached"
		}
	}

	return "Detached"
}
