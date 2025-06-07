// pkg/client/node.go - Complete file with all disk operations
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// nodeClient implements NodeInterface using HTTP API
type nodeClient struct {
	client *Client
}

// List returns all nodes
func (c *nodeClient) List() ([]Node, error) {
	debugLog("Listing nodes")

	resp, err := c.client.doRequest("GET", "/nodes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []Node `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// Get returns a specific node
func (c *nodeClient) Get(name string) (*Node, error) {
	debugLog("Getting node: %s", name)

	resp, err := c.client.doRequest("GET", fmt.Sprintf("/nodes/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("node %s not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// Update updates a node
func (c *nodeClient) Update(name string, update *NodeUpdate) (*Node, error) {
	debugLog("Updating node: %s", name)

	resp, err := c.client.doRequest("PUT", fmt.Sprintf("/nodes/%s", name), update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update node: %s", string(body))
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// EnableScheduling enables scheduling on a node
func (c *nodeClient) EnableScheduling(name string) error {
	update := &NodeUpdate{
		AllowScheduling: &[]bool{true}[0],
	}
	_, err := c.Update(name, update)
	return err
}

// DisableScheduling disables scheduling on a node
func (c *nodeClient) DisableScheduling(name string) error {
	update := &NodeUpdate{
		AllowScheduling: &[]bool{false}[0],
	}
	_, err := c.Update(name, update)
	return err
}

// EvictNode requests eviction of all replicas from a node
func (c *nodeClient) EvictNode(name string) error {
	update := &NodeUpdate{
		EvictionRequested: &[]bool{true}[0],
	}
	_, err := c.Update(name, update)
	return err
}

// AddDisk adds a disk to a node
func (c *nodeClient) AddDisk(nodeName string, disk DiskUpdate) error {
	debugLog("Adding disk to node %s: %s", nodeName, disk.Path)

	// Get current node
	node, err := c.Get(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Generate disk ID
	diskID := fmt.Sprintf("disk-%s", sanitizeDiskPath(disk.Path))

	// Check if disk already exists
	if _, exists := node.Disks[diskID]; exists {
		return fmt.Errorf("disk with path %s already exists", disk.Path)
	}

	// Prepare the disk data
	diskData := map[string]interface{}{
		"path":            disk.Path,
		"allowScheduling": true,
		"storageReserved": disk.StorageReserved,
		"tags":            disk.Tags,
	}

	// Add disk via API
	path := fmt.Sprintf("/nodes/%s/disks", nodeName)
	resp, err := c.client.doRequest("POST", path, diskData)
	if err != nil {
		return fmt.Errorf("failed to add disk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add disk: %s", string(body))
	}

	return nil
}

// RemoveDisk removes a disk from a node
func (c *nodeClient) RemoveDisk(nodeName, diskID string) error {
	debugLog("Removing disk %s from node %s", diskID, nodeName)

	path := fmt.Sprintf("/nodes/%s/disks/%s", nodeName, diskID)
	resp, err := c.client.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to remove disk: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove disk: %s", string(body))
	}

	return nil
}

// UpdateDiskTags updates tags for a disk
func (c *nodeClient) UpdateDiskTags(nodeName, diskID string, tags []string) error {
	debugLog("Updating disk tags for %s on node %s", diskID, nodeName)

	updatePayload := map[string]interface{}{
		"tags": tags,
	}

	path := fmt.Sprintf("/nodes/%s/disks/%s", nodeName, diskID)
	resp, err := c.client.doRequest("PATCH", path, updatePayload)
	if err != nil {
		return fmt.Errorf("failed to update disk tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update disk tags: %s", string(body))
	}

	return nil
}

// EnableDiskScheduling enables scheduling for a specific disk
func (c *nodeClient) EnableDiskScheduling(nodeName, diskID string) error {
	return c.updateDiskScheduling(nodeName, diskID, true)
}

// DisableDiskScheduling disables scheduling for a specific disk
func (c *nodeClient) DisableDiskScheduling(nodeName, diskID string) error {
	return c.updateDiskScheduling(nodeName, diskID, false)
}

// updateDiskScheduling is a helper that updates disk scheduling
func (c *nodeClient) updateDiskScheduling(nodeName, diskID string, allowScheduling bool) error {
	debugLog("Updating disk scheduling for %s on node %s to %v", diskID, nodeName, allowScheduling)

	updatePayload := map[string]interface{}{
		"allowScheduling": allowScheduling,
	}

	// Try the direct disk update endpoint first
	path := fmt.Sprintf("/nodes/%s/disks/%s", nodeName, diskID)
	resp, err := c.client.doRequest("PATCH", path, updatePayload)
	if err != nil {
		return fmt.Errorf("failed to update disk: %w", err)
	}
	defer resp.Body.Close()

	// If the API doesn't support PATCH on individual disks, fall back to updating the entire node
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotFound {
		return c.updateDiskSchedulingViaNode(nodeName, diskID, allowScheduling)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update disk scheduling: %s", string(body))
	}

	action := "enabled"
	if !allowScheduling {
		action = "disabled"
	}
	debugLog("Successfully %s scheduling for disk %s on node %s", action, diskID, nodeName)
	return nil
}

// updateDiskSchedulingViaNode updates disk scheduling by updating the entire node
func (c *nodeClient) updateDiskSchedulingViaNode(
	nodeName, diskID string,
	allowScheduling bool,
) error {
	// Get the current node
	node, err := c.Get(nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Find and update the disk
	disk, exists := node.Disks[diskID]
	if !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	// Update the disk's scheduling status
	disk.AllowScheduling = allowScheduling
	node.Disks[diskID] = disk

	// Create the update payload with all disks
	disksPayload := make(map[string]interface{})
	for id, d := range node.Disks {
		disksPayload[id] = map[string]interface{}{
			"path":            d.Path,
			"allowScheduling": d.AllowScheduling,
			"storageReserved": d.StorageReserved,
			"tags":            d.Tags,
		}
	}

	updatePayload := map[string]interface{}{
		"disks": disksPayload,
	}

	// Update the node
	path := fmt.Sprintf("/nodes/%s", nodeName)
	resp, err := c.client.doRequest("PUT", path, updatePayload)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update disk scheduling: %s", string(body))
	}

	return nil
}

// AddNodeTag adds a tag to a node
func (c *nodeClient) AddNodeTag(nodeName, tag string) error {
	node, err := c.Get(nodeName)
	if err != nil {
		return err
	}

	// Check if tag already exists
	for _, t := range node.Tags {
		if t == tag {
			return nil
		}
	}

	tags := append(node.Tags, tag)
	update := &NodeUpdate{Tags: tags}
	_, err = c.Update(nodeName, update)
	return err
}

// RemoveNodeTag removes a tag from a node
func (c *nodeClient) RemoveNodeTag(nodeName, tag string) error {
	node, err := c.Get(nodeName)
	if err != nil {
		return err
	}

	// Filter out the tag
	var tags []string
	for _, t := range node.Tags {
		if t != tag {
			tags = append(tags, t)
		}
	}

	update := &NodeUpdate{Tags: tags}
	_, err = c.Update(nodeName, update)
	return err
}

// Helper function to sanitize disk path for ID generation
func sanitizeDiskPath(path string) string {
	// Remove leading slash and replace all slashes with dashes
	sanitized := path
	if len(sanitized) > 0 && sanitized[0] == '/' {
		sanitized = sanitized[1:]
	}
	// Replace slashes with dashes
	for i := 0; i < len(sanitized); i++ {
		if sanitized[i] == '/' {
			sanitized = sanitized[:i] + "-" + sanitized[i+1:]
		}
	}
	return sanitized
}
