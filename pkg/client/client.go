// pkg/client/client.go
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config represents the client configuration
type Config struct {
	Endpoint  string
	Namespace string
	Token     string
	Timeout   time.Duration
	Insecure  bool
}

// Client is the Longhorn API client
type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
	crdClient  *LonghornCRDClient // For CRD-based operations
}

// NewClient creates a new Longhorn client
func NewClient(config *Config) (*Client, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Ensure endpoint doesn't have trailing slash
	config.Endpoint = strings.TrimSuffix(config.Endpoint, "/")

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// For development/testing with self-signed certificates
	if config.Insecure {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("%s/v1", config.Endpoint),
	}, nil
}

// doRequest performs an HTTP request
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication if configured
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	// Add namespace header if configured
	if c.config.Namespace != "" {
		req.Header.Set("X-Namespace", c.config.Namespace)
	}

	// Debug logging
	debugLog("Request: %s %s", method, url)
	debugLog("Headers: %v", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Debug response
	debugResponse(resp)

	return resp, nil
}

// Nodes returns the node interface
func (c *Client) Nodes() NodeInterface {
	// If we have a CRD client, use it
	if c.crdClient != nil {
		return &crdNodeClient{crdClient: c.crdClient}
	}
	// Otherwise use the HTTP client
	return &nodeClient{client: c}
}

// Volumes returns the volume interface
func (c *Client) Volumes() VolumeInterface {
	// If we have a CRD client, use it
	if c.crdClient != nil {
		return &crdVolumeClient{crdClient: c.crdClient}
	}
	// Otherwise use the HTTP client
	return &volumeClient{client: c}
}

// Settings returns the settings interface
func (c *Client) Settings() SettingsInterface {
	return &settingsClient{client: c}
}

// EngineImages returns the engine image interface
func (c *Client) EngineImages() EngineImageInterface {
	return &engineImageClient{client: c}
}

// Backups returns the backup interface
func (c *Client) Backups() BackupInterface {
	return &backupClient{client: c}
}

// Events returns the event interface
func (c *Client) Events() EventInterface {
	return &eventClient{client: c}
}

// NodeInterface defines node operations
type NodeInterface interface {
	List() ([]Node, error)
	Get(name string) (*Node, error)
	Update(name string, update *NodeUpdate) (*Node, error)
	EnableScheduling(name string) error
	DisableScheduling(name string) error
	EvictNode(name string) error
	AddDisk(nodeName string, disk DiskUpdate) error
	RemoveDisk(nodeName, diskID string) error
	UpdateDiskTags(nodeName, diskID string, tags []string) error
	AddNodeTag(nodeName, tag string) error
	RemoveNodeTag(nodeName, tag string) error
	EnableDiskScheduling(nodeName, diskID string) error
	DisableDiskScheduling(nodeName, diskID string) error
}

// VolumeInterface defines volume operations
type VolumeInterface interface {
	List() ([]Volume, error)
	Get(name string) (*Volume, error)
	Create(volume *VolumeCreateInput) (*Volume, error)
	Delete(name string) error
	Update(name string, volume *VolumeUpdateInput) (*Volume, error)
	Attach(name string, input *VolumeAttachInput) (*Volume, error)
	Detach(name string) error
}

// SettingsInterface defines settings operations
type SettingsInterface interface {
	List() (map[string]Setting, error)
	Get(name string) (*Setting, error)
	Update(name string, value string) (*Setting, error)
}

// BackupInterface defines backup operations
type BackupInterface interface {
	List(volumeName string) ([]Backup, error)
	Get(backupName string) (*Backup, error)
	Create(volumeName string, input *BackupCreateInput) (*Backup, error)
	Delete(backupName string) error
	GetTarget() (*BackupTarget, error)
	SetTarget(target *BackupTarget) error
}

// EngineImageInterface defines engine image operations
type EngineImageInterface interface {
	List() ([]EngineImage, error)
	Get(name string) (*EngineImage, error)
	Delete(name string) error
}

// EventInterface defines event operations
type EventInterface interface {
	List(opts EventListOptions) ([]Event, error)
	Watch(ctx context.Context, opts EventListOptions, callback func(Event)) error
}
