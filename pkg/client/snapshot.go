package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// snapshotClient implements SnapshotInterface using HTTP API
// It interacts with /volumes/{volume}/snapshots endpoints.
type snapshotClient struct {
	client *Client
}

// List returns all snapshots for a volume
func (c *snapshotClient) List(volumeName string) ([]Snapshot, error) {
	path := fmt.Sprintf("/volumes/%s/snapshots", volumeName)
	resp, err := c.client.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []Snapshot `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Data, nil
}

// Create creates a snapshot for the given volume
func (c *snapshotClient) Create(volumeName string, input *SnapshotCreateInput) (*Snapshot, error) {
	path := fmt.Sprintf("/volumes/%s/snapshots", volumeName)
	resp, err := c.client.doRequest("POST", path, input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var snapshot Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &snapshot, nil
}

// Delete removes a snapshot from a volume
func (c *snapshotClient) Delete(volumeName, snapshotName string) error {
	path := fmt.Sprintf("/volumes/%s/snapshots/%s", volumeName, snapshotName)
	resp, err := c.client.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
