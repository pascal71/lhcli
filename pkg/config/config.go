// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
	Defaults       Defaults  `yaml:"defaults"`
}

// Context represents a Longhorn cluster context
type Context struct {
	Name      string `yaml:"name"`
	Endpoint  string `yaml:"endpoint"`
	Namespace string `yaml:"namespace"`
	Auth      Auth   `yaml:"auth"`
}

// Auth represents authentication configuration
type Auth struct {
	Type    string `yaml:"type"` // "none", "token", "kubeconfig"
	Token   string `yaml:"token,omitempty"`
	Path    string `yaml:"path,omitempty"`    // Path to kubeconfig
	Context string `yaml:"context,omitempty"` // Kubernetes context to use
}

// Defaults represents default settings
type Defaults struct {
	OutputFormat string `yaml:"output-format"`
	Confirmation bool   `yaml:"confirmation"`
	Timeout      string `yaml:"timeout"`
}

// Load loads the configuration from file
func Load(path string) (*Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".lhcli", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file exists, return smart defaults
			return SmartDefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save(path string) error {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".lhcli", "config.yaml")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// GetContext returns the context by name
func (c *Config) GetContext(name string) (*Context, error) {
	if name == "" {
		name = c.CurrentContext
	}

	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			return &ctx, nil
		}
	}

	return nil, fmt.Errorf("context %s not found", name)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Contexts: []Context{},
		Defaults: Defaults{
			OutputFormat: "table",
			Confirmation: true,
			Timeout:      "30s",
		},
	}
}

// SmartDefaultConfig returns a configuration that uses kubeconfig by default
func SmartDefaultConfig() *Config {
	// Check for KUBECONFIG environment variable
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		// Use default location
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Create a default context that uses kubeconfig
	defaultContext := Context{
		Name:      "default",
		Namespace: "longhorn-system",
		Auth: Auth{
			Type:    "kubeconfig",
			Path:    kubeconfigPath,
			Context: "", // Empty means use current context from kubeconfig
		},
	}

	return &Config{
		Contexts:       []Context{defaultContext},
		CurrentContext: "default",
		Defaults: Defaults{
			OutputFormat: "table",
			Confirmation: true,
			Timeout:      "30s",
		},
	}
}

// GetKubeconfigPath returns the path to kubeconfig file
func GetKubeconfigPath() string {
	// First check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Fall back to default location
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}
