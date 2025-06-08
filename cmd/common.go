package cmd

import (
	"fmt"
	"time"

	"github.com/pascal71/lhcli/pkg/client"
	"github.com/pascal71/lhcli/pkg/config"
	"k8s.io/client-go/kubernetes"
)

// getClient creates a client based on the current configuration
func getClient() (*client.Client, error) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get current context
	ctx, err := cfg.GetContext(context)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Determine namespace to use
	ns := namespace
	if ns == "" {
		// Use namespace from context if available
		if ctx.Namespace != "" {
			ns = ctx.Namespace
		} else {
			// Default to longhorn-system
			ns = "longhorn-system"
		}
	}

	// Check auth type
	switch ctx.Auth.Type {
	case "kubeconfig":
		// Use kubeconfig for authentication
		kubeConfig := &client.KubeConfig{
			ConfigPath: ctx.Auth.Path,
			Context:    ctx.Auth.Context,
			Namespace:  ns,
		}
		return client.NewClientFromKubeconfig(kubeConfig)

	case "token":
		// Use direct connection with token
		clientConfig := &client.Config{
			Endpoint:  ctx.Endpoint,
			Namespace: ns,
			Token:     ctx.Auth.Token,
			Timeout:   30 * time.Second,
		}
		return client.NewClient(clientConfig)

	case "none", "":
		// Direct connection without auth
		clientConfig := &client.Config{
			Endpoint:  ctx.Endpoint,
			Namespace: ns,
			Timeout:   30 * time.Second,
		}
		return client.NewClient(clientConfig)

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", ctx.Auth.Type)
	}
}

// getKubeClient creates a Kubernetes clientset using kubeconfig from the current context
func getKubeClient() (*kubernetes.Clientset, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetContext(context)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if ctx.Auth.Type != "kubeconfig" {
		return nil, fmt.Errorf("kubeconfig auth required for Kubernetes operations")
	}

	kubeConfig := &client.KubeConfig{
		ConfigPath: ctx.Auth.Path,
		Context:    ctx.Auth.Context,
		Namespace:  namespace,
	}

	clientset, _, err := client.NewKubeClient(kubeConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
