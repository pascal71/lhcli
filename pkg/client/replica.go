// pkg/client/replica.go
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// replicaClient implements ReplicaInterface using HTTP API
type replicaClient struct {
	client *Client
}

// List returns all replicas
func (c *replicaClient) List() ([]Replica, error) {
	debugLog("Listing replicas via HTTP")

	resp, err := c.client.doRequest("GET", "/replicas", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []Replica `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// Get returns a specific replica
func (c *replicaClient) Get(name string) (*Replica, error) {
	debugLog("Getting replica: %s", name)

	resp, err := c.client.doRequest("GET", fmt.Sprintf("/replicas/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("replica %s not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var replica Replica
	if err := json.NewDecoder(resp.Body).Decode(&replica); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &replica, nil
}

// Delete deletes a replica
func (c *replicaClient) Delete(name string) error {
	debugLog("Deleting replica: %s", name)

	resp, err := c.client.doRequest("DELETE", fmt.Sprintf("/replicas/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
