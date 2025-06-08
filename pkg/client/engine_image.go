package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// engineImageClient implements EngineImageInterface using the HTTP API

// List retrieves all engine images.
func (e *engineImageClient) List() ([]EngineImage, error) {
	resp, err := e.client.doRequest("GET", "/engineimages", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []EngineImage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Data, nil
}

// Get retrieves a specific engine image.
func (e *engineImageClient) Get(name string) (*EngineImage, error) {
	resp, err := e.client.doRequest("GET", fmt.Sprintf("/engineimages/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("engine image %s not found", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ei EngineImage
	if err := json.NewDecoder(resp.Body).Decode(&ei); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &ei, nil
}

// Delete removes an engine image.
func (e *engineImageClient) Delete(name string) error {
	resp, err := e.client.doRequest("DELETE", fmt.Sprintf("/engineimages/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
