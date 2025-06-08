package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/formatter"
	"github.com/pascal71/lhcli/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	showReplicas bool // Flag to show replica locations
	showFullIDs  bool // Flag to show full IDs without abbreviation
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

var volumeUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update volume configuration",
	Long:  `Update a Longhorn volume's configuration such as replica count, access mode, etc.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVolumeUpdate,
}

var volumeMapCmd = &cobra.Command{
	Use:   "map [volume-name]",
	Short: "Map Longhorn volumes to Kubernetes PVs",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runVolumeMap,
}

func init() {
	rootCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeCreateCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)
	volumeCmd.AddCommand(volumeGetCmd)
	volumeCmd.AddCommand(volumeUpdateCmd)
	volumeCmd.AddCommand(volumeMapCmd)

	// Volume list flags
	volumeListCmd.Flags().
		BoolVarP(&showReplicas, "show-replicas", "r", false, "Show replica locations (nodes and disk paths)")
	volumeListCmd.Flags().
		BoolVar(&showFullIDs, "full-ids", false, "Show full disk IDs and replica names without abbreviation")

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
	volumeGetCmd.Flags().
		BoolVar(&showFullIDs, "full-ids", false, "Show full disk IDs and replica names without abbreviation")

	// Volume update flags
	volumeUpdateCmd.Flags().Int("replicas", 0, "Number of replicas (0 means no change)")
	volumeUpdateCmd.Flags().String("access-mode", "", "Access mode (rwo|rwx)")
	volumeUpdateCmd.Flags().
		String("data-locality", "", "Data locality (disabled|best-effort|strict-local)")
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

	// If show-replicas flag is set, fetch node information
	if showReplicas && (output == "table" || output == "wide" || output == "") {
		// Fetch all nodes to get disk path information
		nodes, err := c.Nodes().List()
		if err != nil {
			// Don't fail completely, just warn
			fmt.Fprintf(
				cmd.ErrOrStderr(),
				"Warning: failed to fetch node information for disk paths: %v\n",
				err,
			)
		} else {
			// Create a map for quick lookup of disk paths by node and disk ID
			diskPathMap := make(map[string]map[string]string) // nodeName -> diskID -> path
			for _, node := range nodes {
				if diskPathMap[node.Name] == nil {
					diskPathMap[node.Name] = make(map[string]string)
				}
				for diskID, disk := range node.Disks {
					diskPathMap[node.Name][diskID] = disk.Path
				}
			}

			// Enrich volume data with disk paths
			for i := range volumes {
				for j := range volumes[i].Replicas {
					if paths, ok := diskPathMap[volumes[i].Replicas[j].NodeID]; ok {
						if path, ok := paths[volumes[i].Replicas[j].DiskID]; ok {
							volumes[i].Replicas[j].DiskPath = path
						}
					}
				}
			}
		}
	}

	// Handle output format
	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(volumes)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(volumes)
	case "wide":
		if showReplicas {
			return printVolumesWideWithReplicas(volumes)
		}
		return printVolumesWide(volumes)
	default:
		if showReplicas {
			return printVolumesWithReplicas(volumes)
		}
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

func runVolumeUpdate(cmd *cobra.Command, args []string) error {
	volumeName := args[0]

	replicas, _ := cmd.Flags().GetInt("replicas")
	accessMode, _ := cmd.Flags().GetString("access-mode")
	dataLocality, _ := cmd.Flags().GetString("data-locality")

	// Check if any updates were specified
	if replicas == 0 && accessMode == "" && dataLocality == "" {
		return fmt.Errorf("no updates specified. Use --replicas, --access-mode, or --data-locality")
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	// Build update input
	update := &client.VolumeUpdateInput{}

	if replicas > 0 {
		update.NumberOfReplicas = &replicas
	}
	if accessMode != "" {
		update.AccessMode = accessMode
	}
	if dataLocality != "" {
		update.DataLocality = dataLocality
	}

	// Perform the update
	volume, err := c.Volumes().Update(volumeName, update)
	if err != nil {
		return fmt.Errorf("failed to update volume: %w", err)
	}

	fmt.Printf("✓ Volume %s updated successfully\n", volume.Name)

	// Show what was updated
	if replicas > 0 {
		fmt.Printf("  Replicas: %d\n", replicas)
	}
	if accessMode != "" {
		fmt.Printf("  Access Mode: %s\n", accessMode)
	}
	if dataLocality != "" {
		fmt.Printf("  Data Locality: %s\n", dataLocality)
	}

	// Optionally show the current state
	if !quiet {
		fmt.Println("\nCurrent volume state:")
		return printVolumeDetails(volume, false)
	}

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

func runVolumeMap(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	kubeClient, err := getKubeClient()
	if err != nil {
		return err
	}

	volumes, err := c.Volumes().List()
	if err != nil {
		return fmt.Errorf("failed to list volumes: %w", err)
	}

	pvs, err := kubeClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	filter := ""
	if len(args) == 1 {
		filter = args[0]
	}

	pvMap := make(map[string]*v1.PersistentVolume)
	for _, pv := range pvs.Items {
		if pv.Spec.CSI != nil && pv.Spec.CSI.Driver == "driver.longhorn.io" {
			pvMap[pv.Spec.CSI.VolumeHandle] = &pv
		}
	}

	type mapping struct {
		Volume    string `json:"volume"`
		PV        string `json:"pv,omitempty"`
		PVC       string `json:"pvc,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}

	var result []mapping
	for _, v := range volumes {
		if filter != "" && v.Name != filter {
			continue
		}
		m := mapping{Volume: v.Name}
		if pv, ok := pvMap[v.Name]; ok {
			m.PV = pv.Name
			if pv.Spec.ClaimRef != nil {
				m.PVC = pv.Spec.ClaimRef.Name
				m.Namespace = pv.Spec.ClaimRef.Namespace
			}
		}
		result = append(result, m)
	}

	switch output {
	case "json":
		return formatter.NewJSONFormatter(true).Format(result)
	case "yaml":
		return formatter.NewYAMLFormatter().Format(result)
	default:
		headers := []string{"VOLUME", "PV", "PVC", "NAMESPACE"}
		table := formatter.NewTableFormatter(headers)
		for _, m := range result {
			table.AddRow([]string{m.Volume, m.PV, m.PVC, m.Namespace})
		}
		return table.Format(nil)
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

		// Convert size to human readable if it's a number
		size := volume.Size
		if sizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}

		formatter.AddRow([]string{
			volume.Name,
			size,
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

		// Convert size to human readable if it's a number
		size := volume.Size
		if sizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}

		formatter.AddRow([]string{
			volume.Name,
			size,
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

// Print volumes with detailed replica information
func printVolumesWithReplicas(volumes []client.Volume) error {
	for i, volume := range volumes {
		if i > 0 {
			fmt.Println() // Add blank line between volumes
		}

		// Print volume header
		fmt.Printf("Volume: %s\n", volume.Name)

		// Convert size to human readable if it's a number
		size := volume.Size
		if sizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}

		fmt.Printf("  Size: %s, Replicas: %d, State: %s, Robustness: %s\n",
			size,
			volume.NumberOfReplicas,
			getVolumeState(volume),
			volume.Robustness)

		// Print replica locations
		if len(volume.Replicas) > 0 {
			fmt.Println("  Replica Locations:")

			// Create a formatter for the replica table
			headers := []string{"REPLICA NAME", "NODE", "DISK ID", "DISK PATH", "MODE"}
			replicaFormatter := formatter.NewTableFormatter(headers)

			for _, replica := range volume.Replicas {
				diskPath := replica.DiskPath
				if diskPath == "" {
					diskPath = "<unknown>"
				}

				replicaFormatter.AddRow([]string{
					shortenReplicaName(replica.Name),
					replica.NodeID,
					shortenDiskID(replica.DiskID),
					diskPath,
					replica.Mode,
				})
			}

			// Print with indent
			// Use a string builder to capture the formatter output
			var sb strings.Builder
			replicaFormatter.Format(&sb)
			output := sb.String()
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line != "" && strings.TrimSpace(line) != "" {
					fmt.Printf("  %s\n", line)
				}
			}
		} else {
			fmt.Println("  Replica Locations: <none>")
		}
	}

	return nil
}

// Print volumes in wide format with replica information
func printVolumesWideWithReplicas(volumes []client.Volume) error {
	headers := []string{
		"NAME",
		"SIZE",
		"REPLICAS",
		"STATE",
		"ROBUSTNESS",
		"FRONTEND",
		"ACCESS",
		"REPLICA LOCATIONS",
		"CREATED",
	}
	formatter := formatter.NewTableFormatter(headers)

	for _, volume := range volumes {
		// Build replica locations string
		replicaLocations := ""
		if len(volume.Replicas) > 0 {
			locations := []string{}
			for _, replica := range volume.Replicas {
				location := fmt.Sprintf("%s:%s", replica.NodeID, replica.DiskPath)
				if replica.DiskPath == "" {
					diskID := replica.DiskID
					if !showFullIDs {
						diskID = shortenDiskID(replica.DiskID)
					}
					location = fmt.Sprintf("%s:%s", replica.NodeID, diskID)
				}
				locations = append(locations, location)
			}
			replicaLocations = strings.Join(locations, ", ")
			// Truncate if too long and not showing full IDs
			if !showFullIDs && len(replicaLocations) > 50 {
				replicaLocations = replicaLocations[:47] + "..."
			}
		} else {
			replicaLocations = "<none>"
		}

		// Convert size to human readable if it's a number
		size := volume.Size
		if sizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
			size = utils.FormatSize(sizeInt)
		}

		formatter.AddRow([]string{
			volume.Name,
			size,
			fmt.Sprintf("%d", volume.NumberOfReplicas),
			getVolumeState(volume),
			volume.Robustness,
			volume.Frontend,
			volume.AccessMode,
			replicaLocations,
			formatTime(volume.Created),
		})
	}

	return formatter.Format(nil)
}

func printVolumeDetails(volume *client.Volume, detailed bool) error {
	fmt.Printf("Name:              %s\n", volume.Name)

	// Convert size to human readable if it's a number
	size := volume.Size
	if sizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
		size = utils.FormatSize(sizeInt)
	}
	fmt.Printf("Size:              %s\n", size)

	// Show actual size if available
	if volume.ActualSize > 0 {
		fmt.Printf("Actual Size:       %s\n", utils.FormatSize(volume.ActualSize))
	}

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

		// Calculate per-replica size if we have actual size
		var perReplicaSize int64
		if volume.ActualSize > 0 && len(volume.Replicas) > 0 {
			perReplicaSize = volume.ActualSize
		}

		for _, replica := range volume.Replicas {
			replicaName := replica.Name
			diskID := replica.DiskID

			// Apply abbreviation unless --full-ids is set
			if !showFullIDs {
				replicaName = shortenReplicaName(replica.Name)
				if len(diskID) > 20 {
					diskID = shortenDiskID(replica.DiskID)
				}
			}

			fmt.Printf("  %s:\n", replicaName)
			fmt.Printf("    Node:   %s\n", replica.NodeID)
			fmt.Printf("    Mode:   %s\n", replica.Mode)
			if replica.DiskID != "" {
				fmt.Printf("    Disk:   %s\n", diskID)
			}
			if replica.DiskPath != "" {
				fmt.Printf("    Path:   %s\n", replica.DiskPath)
			}

			// Show size information
			if replica.SpecSize != "" {
				specSize := replica.SpecSize
				// Convert to human readable if it's a number
				if specSizeInt, err := strconv.ParseInt(replica.SpecSize, 10, 64); err == nil {
					specSize = utils.FormatSize(specSizeInt)
				}
				fmt.Printf("    Size (Allocated): %s\n", specSize)
			}

			// Show estimated actual size per replica
			if perReplicaSize > 0 {
				fmt.Printf("    Size (Estimated): %s\n", utils.FormatSize(perReplicaSize))
			}
		}

		// Show total consumed size
		if volume.ActualSize > 0 && len(volume.Replicas) > 0 {
			totalConsumed := volume.ActualSize * int64(len(volume.Replicas))
			fmt.Printf("\nStorage Summary:\n")
			fmt.Printf("  Volume Actual Size: %s\n", utils.FormatSize(volume.ActualSize))
			fmt.Printf("  Total Consumed:     %s (across %d replicas)\n",
				utils.FormatSize(totalConsumed), len(volume.Replicas))

			// Show efficiency if we have spec size
			if specSizeInt, err := strconv.ParseInt(volume.Size, 10, 64); err == nil {
				efficiency := float64(volume.ActualSize) / float64(specSizeInt) * 100
				fmt.Printf("  Space Efficiency:   %.1f%% of allocated\n", efficiency)
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

// Helper function to shorten replica names for better display
func shortenReplicaName(name string) string {
	// If the name is very long (like pvc-xxx-r-xxx), shorten it
	if len(name) > 40 {
		parts := strings.Split(name, "-")
		if len(parts) > 3 {
			// Keep first part and last 2 parts
			return fmt.Sprintf("%s...%s-%s", parts[0], parts[len(parts)-2], parts[len(parts)-1])
		}
	}
	return name
}

// Helper function to shorten disk IDs for better display
func shortenDiskID(diskID string) string {
	// If it's a UUID-like string, shorten it
	if len(diskID) > 20 && strings.Count(diskID, "-") >= 4 {
		parts := strings.Split(diskID, "-")
		if len(parts) >= 5 {
			// Show first and last part of UUID
			return fmt.Sprintf("%s...%s", parts[0], parts[len(parts)-1])
		}
	}
	return diskID
}

// Helper function to format time strings
func formatTime(timeStr string) string {
	// For now, just return the time string
	// TODO: Parse and format nicely (e.g., "2d ago")
	if len(timeStr) > 19 {
		return timeStr[:19] // Return just YYYY-MM-DDTHH:MM:SS
	}
	return timeStr
}
