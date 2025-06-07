// pkg/client/node_crd.go
package client

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// AddDisk adds a new disk to a Longhorn node via CRD
func (c *crdNodeClient) AddDisk(nodeName string, disk DiskUpdate) error {
	// Get current node
	currentUnstructured, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(currentUnstructured.Object, "spec")
	if err != nil {
		return fmt.Errorf("failed to get spec: %w", err)
	}
	if !found {
		spec = make(map[string]interface{})
	}

	// Get existing disks or create new map
	disks, found, err := unstructured.NestedMap(spec, "disks")
	if err != nil {
		return fmt.Errorf("failed to get disks: %w", err)
	}
	if !found || disks == nil {
		disks = make(map[string]interface{})
	}

	// Generate disk ID based on path
	diskID := fmt.Sprintf(
		"disk-%s",
		strings.ReplaceAll(strings.TrimPrefix(disk.Path, "/"), "/", "-"),
	)

	// Check if disk already exists
	if _, exists := disks[diskID]; exists {
		return fmt.Errorf("disk with path %s already exists", disk.Path)
	}

	// Create new disk entry
	newDisk := map[string]interface{}{
		"path":            disk.Path,
		"allowScheduling": true, // Default to true for new disks
		"storageReserved": disk.StorageReserved,
	}

	// Add tags if provided
	if len(disk.Tags) > 0 {
		tags := make([]interface{}, len(disk.Tags))
		for i, tag := range disk.Tags {
			tags[i] = tag
		}
		newDisk["tags"] = tags
	}

	// Add the new disk to the disks map
	disks[diskID] = newDisk

	// Update the spec with the new disks map
	if err := unstructured.SetNestedMap(spec, disks, "disks"); err != nil {
		return fmt.Errorf("failed to set disks: %w", err)
	}

	// Set the updated spec back to the node
	if err := unstructured.SetNestedMap(currentUnstructured.Object, spec, "spec"); err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	// Update the node via CRD
	_, err = c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), currentUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	debugLog("Successfully added disk %s to node %s", disk.Path, nodeName)
	return nil
}

// RemoveDisk removes a disk from a Longhorn node via CRD
func (c *crdNodeClient) RemoveDisk(nodeName, diskID string) error {
	// Get current node
	currentUnstructured, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(currentUnstructured.Object, "spec")
	if err != nil {
		return fmt.Errorf("failed to get spec: %w", err)
	}
	if !found {
		return fmt.Errorf("spec not found in node %s", nodeName)
	}

	// Get existing disks
	disks, found, err := unstructured.NestedMap(spec, "disks")
	if err != nil {
		return fmt.Errorf("failed to get disks: %w", err)
	}
	if !found || disks == nil {
		return fmt.Errorf("no disks found on node %s", nodeName)
	}

	// Check if disk exists
	if _, exists := disks[diskID]; !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	// Remove the disk
	delete(disks, diskID)

	// Update the spec with the modified disks map
	if err := unstructured.SetNestedMap(spec, disks, "disks"); err != nil {
		return fmt.Errorf("failed to set disks: %w", err)
	}

	// Set the updated spec back to the node
	if err := unstructured.SetNestedMap(currentUnstructured.Object, spec, "spec"); err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	// Update the node via CRD
	_, err = c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), currentUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	debugLog("Successfully removed disk %s from node %s", diskID, nodeName)
	return nil
}

// UpdateDiskTags updates tags for a specific disk on a Longhorn node via CRD
func (c *crdNodeClient) UpdateDiskTags(nodeName, diskID string, tags []string) error {
	// Get current node
	currentUnstructured, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(currentUnstructured.Object, "spec")
	if err != nil {
		return fmt.Errorf("failed to get spec: %w", err)
	}
	if !found {
		return fmt.Errorf("spec not found in node %s", nodeName)
	}

	// Get existing disks
	disks, found, err := unstructured.NestedMap(spec, "disks")
	if err != nil {
		return fmt.Errorf("failed to get disks: %w", err)
	}
	if !found || disks == nil {
		return fmt.Errorf("no disks found on node %s", nodeName)
	}

	// Get the specific disk
	diskData, exists := disks[diskID]
	if !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	// Convert disk data to map
	disk, ok := diskData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid disk data format for disk %s", diskID)
	}

	// Update tags
	if len(tags) > 0 {
		tagList := make([]interface{}, len(tags))
		for i, tag := range tags {
			tagList[i] = tag
		}
		disk["tags"] = tagList
	} else {
		// Remove tags if empty list provided
		delete(disk, "tags")
	}

	// Update the disk in the disks map
	disks[diskID] = disk

	// Update the spec with the modified disks map
	if err := unstructured.SetNestedMap(spec, disks, "disks"); err != nil {
		return fmt.Errorf("failed to set disks: %w", err)
	}

	// Set the updated spec back to the node
	if err := unstructured.SetNestedMap(currentUnstructured.Object, spec, "spec"); err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	// Update the node via CRD
	_, err = c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), currentUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	debugLog("Successfully updated tags for disk %s on node %s", diskID, nodeName)
	return nil
}

// EnableDiskScheduling enables scheduling for a specific disk on a Longhorn node
func (c *crdNodeClient) EnableDiskScheduling(nodeName, diskID string) error {
	return c.updateDiskScheduling(nodeName, diskID, true)
}

// DisableDiskScheduling disables scheduling for a specific disk on a Longhorn node
func (c *crdNodeClient) DisableDiskScheduling(nodeName, diskID string) error {
	return c.updateDiskScheduling(nodeName, diskID, false)
}

// updateDiskScheduling is a helper function to update disk scheduling
func (c *crdNodeClient) updateDiskScheduling(nodeName, diskID string, allowScheduling bool) error {
	// Get current node
	currentUnstructured, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(currentUnstructured.Object, "spec")
	if err != nil {
		return fmt.Errorf("failed to get spec: %w", err)
	}
	if !found {
		return fmt.Errorf("spec not found in node %s", nodeName)
	}

	// Get existing disks
	disks, found, err := unstructured.NestedMap(spec, "disks")
	if err != nil {
		return fmt.Errorf("failed to get disks: %w", err)
	}
	if !found || disks == nil {
		return fmt.Errorf("no disks found on node %s", nodeName)
	}

	// Get the specific disk
	diskData, exists := disks[diskID]
	if !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	// Convert disk data to map
	disk, ok := diskData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid disk data format for disk %s", diskID)
	}

	// Update allowScheduling
	disk["allowScheduling"] = allowScheduling

	// Update the disk in the disks map
	disks[diskID] = disk

	// Update the spec with the modified disks map
	if err := unstructured.SetNestedMap(spec, disks, "disks"); err != nil {
		return fmt.Errorf("failed to set disks: %w", err)
	}

	// Set the updated spec back to the node
	if err := unstructured.SetNestedMap(currentUnstructured.Object, spec, "spec"); err != nil {
		return fmt.Errorf("failed to update spec: %w", err)
	}

	// Update the node via CRD
	_, err = c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), currentUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	action := "enabled"
	if !allowScheduling {
		action = "disabled"
	}
	debugLog("Successfully %s scheduling for disk %s on node %s", action, diskID, nodeName)
	return nil
}
