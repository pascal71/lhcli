// pkg/client/node.go
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// nodeClient implements NodeInterface
type nodeClient struct {
	client *Client
}

// List returns all nodes
func (n *nodeClient) List() ([]Node, error) {
	resp, err := n.client.doRequest("GET", "/nodes", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var nodeList NodeList
	if err := json.NewDecoder(resp.Body).Decode(&nodeList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return nodeList.Items, nil
}

// Get returns a specific node
func (n *nodeClient) Get(name string) (*Node, error) {
	resp, err := n.client.doRequest("GET", fmt.Sprintf("/nodes/%s", name), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", name, err)
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
func (n *nodeClient) Update(name string, update *NodeUpdate) (*Node, error) {
	resp, err := n.client.doRequest("PUT", fmt.Sprintf("/nodes/%s", name), update)
	if err != nil {
		return nil, fmt.Errorf("failed to update node %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// EnableScheduling enables scheduling on a node
func (n *nodeClient) EnableScheduling(name string) error {
	allowScheduling := true
	update := &NodeUpdate{
		AllowScheduling: &allowScheduling,
	}

	_, err := n.Update(name, update)
	return err
}

// DisableScheduling disables scheduling on a node
func (n *nodeClient) DisableScheduling(name string) error {
	allowScheduling := false
	update := &NodeUpdate{
		AllowScheduling: &allowScheduling,
	}

	_, err := n.Update(name, update)
	return err
}

// EvictNode requests eviction of all replicas from a node
func (n *nodeClient) EvictNode(name string) error {
	evictionRequested := true
	update := &NodeUpdate{
		EvictionRequested: &evictionRequested,
	}

	_, err := n.Update(name, update)
	return err
}

// AddDisk adds a new disk to a node
func (n *nodeClient) AddDisk(nodeName string, disk DiskUpdate) error {
	node, err := n.Get(nodeName)
	if err != nil {
		return err
	}

	if node.Disks == nil {
		node.Disks = make(map[string]Disk)
	}

	// Generate disk ID if not provided
	diskID := fmt.Sprintf("disk-%s", disk.Path)
	
	update := &NodeUpdate{
		Disks: map[string]DiskUpdate{
			diskID: disk,
		},
	}

	_, err = n.Update(nodeName, update)
	return err
}

// RemoveDisk removes a disk from a node
func (n *nodeClient) RemoveDisk(nodeName, diskID string) error {
	node, err := n.Get(nodeName)
	if err != nil {
		return err
	}

	if _, exists := node.Disks[diskID]; !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	// Mark disk for removal by setting allowScheduling to false and evictionRequested to true
	evictionRequested := true
	allowScheduling := false
	update := &NodeUpdate{
		Disks: map[string]DiskUpdate{
			diskID: {
				Path:              node.Disks[diskID].Path,
				AllowScheduling:   &allowScheduling,
				EvictionRequested: &evictionRequested,
			},
		},
	}

	_, err = n.Update(nodeName, update)
	return err
}

// UpdateDiskTags updates tags for a specific disk
func (n *nodeClient) UpdateDiskTags(nodeName, diskID string, tags []string) error {
	node, err := n.Get(nodeName)
	if err != nil {
		return err
	}

	disk, exists := node.Disks[diskID]
	if !exists {
		return fmt.Errorf("disk %s not found on node %s", diskID, nodeName)
	}

	update := &NodeUpdate{
		Disks: map[string]DiskUpdate{
			diskID: {
				Path: disk.Path,
				Tags: tags,
			},
		},
	}

	_, err = n.Update(nodeName, update)
	return err
}

// AddNodeTag adds a tag to a node
func (n *nodeClient) AddNodeTag(nodeName, tag string) error {
	node, err := n.Get(nodeName)
	if err != nil {
		return err
	}

	// Check if tag already exists
	for _, t := range node.Tags {
		if t == tag {
			return nil // Tag already exists
		}
	}

	tags := append(node.Tags, tag)
	update := &NodeUpdate{
		Tags: tags,
	}

	_, err = n.Update(nodeName, update)
	return err
}

// RemoveNodeTag removes a tag from a node
func (n *nodeClient) RemoveNodeTag(nodeName, tag string) error {
	node, err := n.Get(nodeName)
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

	update := &NodeUpdate{
		Tags: tags,
	}

	_, err = n.Update(nodeName, update)
	return err
}
