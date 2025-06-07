// pkg/client/kube.go
package client

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// KubeConfig represents Kubernetes configuration
type KubeConfig struct {
	ConfigPath string
	Context    string
	Namespace  string
}

// NewKubeClient creates a new Kubernetes client from kubeconfig
func NewKubeClient(config *KubeConfig) (*kubernetes.Clientset, *rest.Config, error) {
	// Build config from kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if config.ConfigPath != "" {
		loadingRules.ExplicitPath = config.ConfigPath
	} else if home := homedir.HomeDir(); home != "" {
		loadingRules.ExplicitPath = filepath.Join(home, ".kube", "config")
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if config.Context != "" {
		configOverrides.CurrentContext = config.Context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, restConfig, nil
}

// NewClientFromKubeconfig creates a Longhorn client using kubeconfig
func NewClientFromKubeconfig(kubeConfig *KubeConfig) (*Client, error) {
	_, restConfig, err := NewKubeClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	namespace := kubeConfig.Namespace
	if namespace == "" {
		namespace = "longhorn-system"
	}

	// Use CRD client for Kubernetes environments
	debugLog("Using CRD client for Longhorn resources")
	return NewLonghornCRDClient(restConfig, namespace)
}

// GetCurrentNamespace gets the current namespace from kubeconfig context
func GetCurrentNamespace(configPath, contextName string) (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if configPath != "" {
		loadingRules.ExplicitPath = configPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return "", err
	}

	return namespace, nil
}
