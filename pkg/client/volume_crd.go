// pkg/client/volume_crd.go
package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/pascal71/lhcli/pkg/utils"
)

// volumeClient implementation for CRDs
type crdVolumeClient struct {
	crdClient *LonghornCRDClient
}

// Volumes returns the volume interface for CRD operations
func (c *Client) VolumesCRD() VolumeInterface {
	if c.crdClient != nil {
		return &crdVolumeClient{crdClient: c.crdClient}
	}
	return nil
}

// List returns all Longhorn volumes with their replicas
func (c *crdVolumeClient) List() ([]Volume, error) {
	debugLog("Listing Longhorn volumes via CRD")

	list, err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	volumes := make([]Volume, 0, len(list.Items))
	for _, item := range list.Items {
		volume, err := unstructuredToVolume(&item)
		if err != nil {
			debugLog("Failed to convert volume %s: %v", item.GetName(), err)
			continue
		}
		volumes = append(volumes, *volume)
	}

	// Fetch all replicas and map them to volumes
	debugLog("Fetching replicas for volumes")
	replicaList, err := c.crdClient.dynamicClient.Resource(replicaGVR).
		Namespace(c.crdClient.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		debugLog("Warning: failed to list replicas: %v", err)
		// Don't fail completely if we can't get replicas
		return volumes, nil
	}

	debugLog("Found %d replicas", len(replicaList.Items))

	// Create a map of volume name to replicas
	volumeReplicas := make(map[string][]Replica)
	for _, item := range replicaList.Items {
		replica, volumeName, err := unstructuredToReplica(&item)
		if err != nil {
			debugLog("Failed to convert replica %s: %v", item.GetName(), err)
			continue
		}
		if volumeName != "" {
			debugLog("Adding replica %s to volume %s", replica.Name, volumeName)
			volumeReplicas[volumeName] = append(volumeReplicas[volumeName], *replica)
		}
	}

	// Assign replicas to volumes
	for i := range volumes {
		if replicas, ok := volumeReplicas[volumes[i].Name]; ok {
			debugLog("Assigning %d replicas to volume %s", len(replicas), volumes[i].Name)
			volumes[i].Replicas = replicas
		}
	}

	return volumes, nil
}

// Get returns a specific volume with its replicas
func (c *crdVolumeClient) Get(name string) (*Volume, error) {
	debugLog("Getting Longhorn volume %s via CRD", name)

	unstructuredVolume, err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get volume %s: %w", name, err)
	}

	volume, err := unstructuredToVolume(unstructuredVolume)
	if err != nil {
		return nil, err
	}

	// Fetch replicas for this volume
	debugLog("Fetching replicas for volume %s", name)
	replicaList, err := c.crdClient.dynamicClient.Resource(replicaGVR).
		Namespace(c.crdClient.namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("longhornvolume=%s", name),
		})
	if err != nil {
		debugLog("Warning: failed to list replicas for volume %s: %v", name, err)
		// Don't fail completely if we can't get replicas
		return volume, nil
	}

	debugLog("Found %d replicas for volume %s", len(replicaList.Items), name)

	volume.Replicas = make([]Replica, 0)
	for _, item := range replicaList.Items {
		replica, _, err := unstructuredToReplica(&item)
		if err != nil {
			debugLog("Failed to convert replica %s: %v", item.GetName(), err)
			continue
		}
		volume.Replicas = append(volume.Replicas, *replica)
	}

	return volume, nil
}

// Create creates a new volume
func (c *crdVolumeClient) Create(input *VolumeCreateInput) (*Volume, error) {
	debugLog("Creating Longhorn volume %s via CRD", input.Name)

	// Parse size string to bytes
	sizeBytes, err := utils.ParseSize(input.Size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	// Create unstructured volume
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "longhorn.io",
		Version: "v1beta2",
		Kind:    "Volume",
	})
	u.SetName(input.Name)
	u.SetNamespace(c.crdClient.namespace)

	// Set spec
	spec := map[string]interface{}{
		"size": fmt.Sprintf(
			"%d",
			sizeBytes,
		), // Convert to string representation of bytes
		"numberOfReplicas": int64(input.NumberOfReplicas), // Convert to int64
	}

	if input.Frontend != "" {
		spec["frontend"] = input.Frontend
	}
	if input.DataLocality != "" {
		spec["dataLocality"] = input.DataLocality
	}
	if input.AccessMode != "" {
		spec["accessMode"] = input.AccessMode
	}
	if input.Migratable {
		spec["migratable"] = input.Migratable
	}
	if input.Encrypted {
		spec["encrypted"] = input.Encrypted
	}
	if len(input.NodeSelector) > 0 {
		selectors := make([]interface{}, len(input.NodeSelector))
		for i, s := range input.NodeSelector {
			selectors[i] = s
		}
		spec["nodeSelector"] = selectors
	}
	if len(input.DiskSelector) > 0 {
		selectors := make([]interface{}, len(input.DiskSelector))
		for i, s := range input.DiskSelector {
			selectors[i] = s
		}
		spec["diskSelector"] = selectors
	}

	if err := unstructured.SetNestedMap(u.Object, spec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set spec: %w", err)
	}

	// Set labels if provided
	if len(input.Labels) > 0 {
		u.SetLabels(input.Labels)
	}

	// Create the volume
	created, err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		Create(context.TODO(), u, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	return unstructuredToVolume(created)
}

// Delete deletes a volume
func (c *crdVolumeClient) Delete(name string) error {
	debugLog("Deleting Longhorn volume %s via CRD", name)

	err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete volume %s: %w", name, err)
	}

	return nil
}

// Update updates a volume
func (c *crdVolumeClient) Update(name string, update *VolumeUpdateInput) (*Volume, error) {
	debugLog("Updating Longhorn volume %s via CRD", name)

	// Get current volume
	current, err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get current volume: %w", err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(current.Object, "spec")
	if err != nil || !found {
		spec = make(map[string]interface{})
	}

	// Apply updates
	if update.NumberOfReplicas != nil {
		spec["numberOfReplicas"] = *update.NumberOfReplicas
	}
	if update.DataLocality != "" {
		spec["dataLocality"] = update.DataLocality
	}
	if update.AccessMode != "" {
		spec["accessMode"] = update.AccessMode
	}

	// Set the updated spec
	if err := unstructured.SetNestedMap(current.Object, spec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to update spec: %w", err)
	}

	// Update labels if provided
	if len(update.Labels) > 0 {
		current.SetLabels(update.Labels)
	}

	// Update the volume
	updated, err := c.crdClient.dynamicClient.Resource(volumeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), current, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}

	return unstructuredToVolume(updated)
}

// Attach attaches a volume to a node
func (c *crdVolumeClient) Attach(name string, input *VolumeAttachInput) (*Volume, error) {
	debugLog("Attaching Longhorn volume %s to node %s via CRD", name, input.HostID)

	// Create an engine CRD for attachment
	engineName := fmt.Sprintf("%s-e-0", name)

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "longhorn.io",
		Version: "v1beta2",
		Kind:    "Engine",
	})
	u.SetName(engineName)
	u.SetNamespace(c.crdClient.namespace)

	// Set owner reference to the volume
	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "longhorn.io/v1beta2",
			Kind:       "Volume",
			Name:       name,
			UID:        "", // Would need to get from volume
		},
	})

	spec := map[string]interface{}{
		"volumeName":      name,
		"nodeID":          input.HostID,
		"disableFrontend": input.DisableFrontend,
	}

	if err := unstructured.SetNestedMap(u.Object, spec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set engine spec: %w", err)
	}

	// For now, return an error as this needs more complex orchestration
	return nil, fmt.Errorf(
		"attach operation requires more complex orchestration - use kubectl or Longhorn UI",
	)
}

// Detach detaches a volume
func (c *crdVolumeClient) Detach(name string) error {
	// For now, return an error as this needs more complex orchestration
	return fmt.Errorf(
		"detach operation requires more complex orchestration - use kubectl or Longhorn UI",
	)
}

// Helper function to convert unstructured to Volume
func unstructuredToVolume(u *unstructured.Unstructured) (*Volume, error) {
	volume := &Volume{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
		Created:   u.GetCreationTimestamp().Format("2006-01-02T15:04:05Z"),
		Labels:    u.GetLabels(),
	}

	// Get spec
	if spec, found, err := unstructured.NestedMap(u.Object, "spec"); err == nil && found {
		if v, ok := spec["size"].(string); ok {
			volume.Size = v
		}
		if v, ok := spec["numberOfReplicas"].(int64); ok {
			volume.NumberOfReplicas = int(v)
		} else if v, ok := spec["numberOfReplicas"].(float64); ok {
			volume.NumberOfReplicas = int(v)
		}
		if v, ok := spec["frontend"].(string); ok {
			volume.Frontend = v
		}
		if v, ok := spec["dataLocality"].(string); ok {
			volume.DataLocality = v
		}
		if v, ok := spec["accessMode"].(string); ok {
			volume.AccessMode = v
		}
		if v, ok := spec["migratable"].(bool); ok {
			volume.Migratable = v
		}
		if v, ok := spec["encrypted"].(bool); ok {
			volume.Encrypted = v
		}
		if v, ok := spec["image"].(string); ok {
			volume.Image = v
		}
	}

	// Get status
	if status, found, err := unstructured.NestedMap(u.Object, "status"); err == nil && found {
		if v, ok := status["state"].(string); ok {
			volume.State = v
		}
		if v, ok := status["robustness"].(string); ok {
			volume.Robustness = v
		}
		if v, ok := status["lastBackup"].(string); ok {
			volume.LastBackup = v
		}
		if v, ok := status["lastBackupAt"].(string); ok {
			volume.LastBackupAt = v
		}
		// Get actual size from status
		if v, ok := status["actualSize"].(int64); ok {
			volume.ActualSize = v
		} else if v, ok := status["actualSize"].(float64); ok {
			volume.ActualSize = int64(v)
		}

		// Get conditions
		if conditions, ok := status["conditions"].([]interface{}); ok {
			volume.Conditions = make(map[string]Status)
			for _, condData := range conditions {
				if condMap, ok := condData.(map[string]interface{}); ok {
					condition := Status{}
					condType := ""

					if v, ok := condMap["type"].(string); ok {
						condType = v
						condition.Type = v
					}
					if v, ok := condMap["status"].(string); ok {
						condition.Status = v
					}
					if v, ok := condMap["message"].(string); ok {
						condition.Message = v
					}
					if v, ok := condMap["reason"].(string); ok {
						condition.Reason = v
					}

					if condType != "" {
						volume.Conditions[condType] = condition
					}
				}
			}
		}
	}

	// Replicas will be fetched separately
	volume.Replicas = make([]Replica, 0)

	return volume, nil
}

// Helper function to convert unstructured to Replica
func unstructuredToReplica(u *unstructured.Unstructured) (*Replica, string, error) {
	replica := &Replica{
		Name: u.GetName(),
	}

	var volumeName string

	// Get spec
	if spec, found, err := unstructured.NestedMap(u.Object, "spec"); err == nil && found {
		// Get volume name from spec
		if v, ok := spec["volumeName"].(string); ok {
			volumeName = v
		}
		// Get node ID from spec
		if v, ok := spec["nodeID"].(string); ok {
			replica.NodeID = v
		}
		// Get disk ID from spec
		if v, ok := spec["diskID"].(string); ok {
			replica.DiskID = v
		}
		// Get disk path from spec
		if v, ok := spec["diskPath"].(string); ok {
			replica.DiskPath = v
		}
		// Get data directory name
		if v, ok := spec["dataDirectoryName"].(string); ok {
			replica.DataPath = v
		}
		// Get failed at
		if v, ok := spec["failedAt"].(string); ok {
			replica.FailedAt = v
		}
		// Get data engine
		if v, ok := spec["dataEngine"].(string); ok {
			replica.DataEngine = v
		}
		// Get image
		if v, ok := spec["image"].(string); ok {
			replica.Image = v
		}
	}

	// Get status
	if status, found, err := unstructured.NestedMap(u.Object, "status"); err == nil && found {
		// Get instance manager name
		if v, ok := status["instanceManagerName"].(string); ok {
			replica.InstanceManager = v
		}
		// Get current state (this is the mode)
		if v, ok := status["currentState"].(string); ok {
			replica.Mode = v
		}
		// Get IP
		if v, ok := status["ip"].(string); ok {
			replica.IP = v
		}
		// Get port
		if v, ok := status["port"].(int64); ok {
			replica.Port = int(v)
		} else if v, ok := status["port"].(float64); ok {
			replica.Port = int(v)
		}
		// Get started status (running)
		if v, ok := status["started"].(bool); ok {
			replica.Running = v
		}
		// Get storage IP
		if v, ok := status["storageIP"].(string); ok {
			replica.StorageIP = v
		}
		// Get current image
		if v, ok := status["currentImage"].(string); ok {
			// Override with current image if available
			replica.Image = v
		}
	}

	// Get labels to find the volume name if not in spec
	if volumeName == "" {
		labels := u.GetLabels()
		if v, ok := labels["longhornvolume"]; ok {
			volumeName = v
		}
	}

	debugLog("Converted replica %s for volume %s", replica.Name, volumeName)

	return replica, volumeName, nil
}
