package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// settingsClient implements SettingsInterface using the HTTP API
// This file provides real implementations for listing, retrieving and
// updating Longhorn settings.

// List retrieves all settings from the Longhorn API.
func (s *settingsClient) List() (map[string]Setting, error) {
	resp, err := s.client.doRequest("GET", "/settings", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []Setting `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	out := make(map[string]Setting)
	for _, s := range result.Data {
		out[s.Name] = s
	}
	return out, nil
}

// Get retrieves a single setting by name.
func (s *settingsClient) Get(name string) (*Setting, error) {
	resp, err := s.client.doRequest("GET", fmt.Sprintf("/settings/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("setting %s not found", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var setting Setting
	if err := json.NewDecoder(resp.Body).Decode(&setting); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &setting, nil
}

// Update updates the value of a Longhorn setting.
func (s *settingsClient) Update(name string, value string) (*Setting, error) {
	payload := map[string]string{
		"value": value,
	}

	resp, err := s.client.doRequest("PUT", fmt.Sprintf("/settings/%s", name), payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var setting Setting
	if err := json.NewDecoder(resp.Body).Decode(&setting); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &setting, nil
}
