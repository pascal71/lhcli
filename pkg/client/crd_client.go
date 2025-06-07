// pkg/client/crd_client.go
package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// LonghornCRDClient uses Kubernetes API to interact with Longhorn CRDs
type LonghornCRDClient struct {
	dynamicClient dynamic.Interface
	namespace     string
}

// Longhorn CRD Group Version Resources
var (
	nodeGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "nodes",
	}

	volumeGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "volumes",
	}

	engineGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "engines",
	}

	replicaGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "replicas",
	}

	settingGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "settings",
	}

	backupGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "backups",
	}

	backupTargetGVR = schema.GroupVersionResource{
		Group:    "longhorn.io",
		Version:  "v1beta2",
		Resource: "backuptargets",
	}
)

// NewLonghornCRDClient creates a new client that uses Kubernetes CRDs
func NewLonghornCRDClient(restConfig *rest.Config, namespace string) (*Client, error) {
	if namespace == "" {
		namespace = "longhorn-system"
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	crdClient := &LonghornCRDClient{
		dynamicClient: dynamicClient,
		namespace:     namespace,
	}

	// Return a Client that wraps the CRD client
	return &Client{
		config: &Config{
			Namespace: namespace,
		},
		// We'll use a special httpClient that delegates to CRD operations
		httpClient: nil,
		baseURL:    "crd://longhorn.io", // Special URL to indicate CRD mode
		// Store the CRD client in a way we can access it
		crdClient: crdClient,
	}, nil
}

// nodeClient implementation for CRDs
type crdNodeClient struct {
	crdClient *LonghornCRDClient
}

// List returns all Longhorn nodes
func (c *crdNodeClient) List() ([]Node, error) {
	debugLog("Listing Longhorn nodes via CRD")

	list, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]Node, 0, len(list.Items))
	for _, item := range list.Items {
		node, err := unstructuredToNode(&item)
		if err != nil {
			debugLog("Failed to convert node %s: %v", item.GetName(), err)
			continue
		}
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

// Get returns a specific node
func (c *crdNodeClient) Get(name string) (*Node, error) {
	debugLog("Getting Longhorn node %s via CRD", name)

	unstructuredNode, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", name, err)
	}

	return unstructuredToNode(unstructuredNode)
}

// Update updates a node
func (c *crdNodeClient) Update(name string, update *NodeUpdate) (*Node, error) {
	// Get current node CRD
	currentUnstructured, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get current node: %w", err)
	}

	// Get the spec
	spec, found, err := unstructured.NestedMap(currentUnstructured.Object, "spec")
	if err != nil || !found {
		spec = make(map[string]interface{})
	}

	// Apply updates to spec
	if update.AllowScheduling != nil {
		spec["allowScheduling"] = *update.AllowScheduling
	}
	if update.EvictionRequested != nil {
		spec["evictionRequested"] = *update.EvictionRequested
	}
	if update.Tags != nil {
		// Convert tags to []interface{}
		tags := make([]interface{}, len(update.Tags))
		for i, tag := range update.Tags {
			tags[i] = tag
		}
		spec["tags"] = tags
	}

	// Set the updated spec back
	if err := unstructured.SetNestedMap(currentUnstructured.Object, spec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to update spec: %w", err)
	}

	// Update via CRD
	updated, err := c.crdClient.dynamicClient.Resource(nodeGVR).
		Namespace(c.crdClient.namespace).
		Update(context.TODO(), currentUnstructured, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update node %s: %w", name, err)
	}

	return unstructuredToNode(updated)
}

// Helper functions to convert between unstructured and typed objects

func unstructuredToNode(u *unstructured.Unstructured) (*Node, error) {
	// Get the spec and status
	spec, _, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil {
		return nil, fmt.Errorf("failed to get spec: %w", err)
	}

	status, _, err := unstructured.NestedMap(u.Object, "status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	node := &Node{
		Name: u.GetName(),
	}

	// Extract fields from spec
	if v, ok := spec["allowScheduling"].(bool); ok {
		node.AllowScheduling = v
	}
	if v, ok := spec["evictionRequested"].(bool); ok {
		node.EvictionRequested = v
	}
	if v, ok := spec["tags"].([]interface{}); ok {
		node.Tags = make([]string, len(v))
		for i, tag := range v {
			if s, ok := tag.(string); ok {
				node.Tags[i] = s
			}
		}
	}

	// Extract fields from status
	if v, ok := status["address"].(string); ok {
		node.Address = v
	}
	if v, ok := status["region"].(string); ok {
		node.Region = v
	}
	if v, ok := status["zone"].(string); ok {
		node.Zone = v
	}

	// Extract conditions (they're an array, not a map)
	if conditions, ok := status["conditions"].([]interface{}); ok {
		node.Conditions = make(map[string]Status)
		for _, condData := range conditions {
			if condMap, ok := condData.(map[string]interface{}); ok {
				condition := Status{}
				condType := ""

				if v, ok := condMap["type"].(string); ok {
					condType = v
					condition.Type = v
				}
				if v, ok := condMap["status"].(string); ok {
					condition.Status = v
				}
				if v, ok := condMap["message"].(string); ok {
					condition.Message = v
				}
				if v, ok := condMap["reason"].(string); ok {
					condition.Reason = v
				}
				if v, ok := condMap["lastProbeTime"].(string); ok {
					condition.LastProbeTime = v
				}
				if v, ok := condMap["lastTransitionTime"].(string); ok {
					condition.LastTransitionTime = v
				}

				if condType != "" {
					node.Conditions[condType] = condition
				}
			}
		}
	}

	// Extract disks
	if disks, ok := spec["disks"].(map[string]interface{}); ok {
		node.Disks = make(map[string]Disk)
		for diskID, diskData := range disks {
			if diskMap, ok := diskData.(map[string]interface{}); ok {
				disk := Disk{}
				if v, ok := diskMap["path"].(string); ok {
					disk.Path = v
				}
				if v, ok := diskMap["allowScheduling"].(bool); ok {
					disk.AllowScheduling = v
				}
				if v, ok := diskMap["evictionRequested"].(bool); ok {
					disk.EvictionRequested = v
				}
				if v, ok := diskMap["storageReserved"].(int64); ok {
					disk.StorageReserved = v
				} else if v, ok := diskMap["storageReserved"].(float64); ok {
					disk.StorageReserved = int64(v)
				}
				if v, ok := diskMap["diskType"].(string); ok {
					disk.DiskType = v
				}
				if tags, ok := diskMap["tags"].([]interface{}); ok {
					disk.Tags = make([]string, 0, len(tags))
					for _, tag := range tags {
						if s, ok := tag.(string); ok {
							disk.Tags = append(disk.Tags, s)
						}
					}
				}

				// Get disk status from status.diskStatus
				if diskStatuses, ok := status["diskStatus"].(map[string]interface{}); ok {
					if diskStatus, ok := diskStatuses[diskID].(map[string]interface{}); ok {
						if v, ok := diskStatus["storageMaximum"].(int64); ok {
							disk.StorageMaximum = v
						} else if v, ok := diskStatus["storageMaximum"].(float64); ok {
							disk.StorageMaximum = int64(v)
						}
						if v, ok := diskStatus["storageAvailable"].(int64); ok {
							disk.StorageAvailable = v
						} else if v, ok := diskStatus["storageAvailable"].(float64); ok {
							disk.StorageAvailable = int64(v)
						}
						if v, ok := diskStatus["storageScheduled"].(int64); ok {
							disk.StorageScheduled = v
						} else if v, ok := diskStatus["storageScheduled"].(float64); ok {
							disk.StorageScheduled = int64(v)
						}
						if v, ok := diskStatus["diskUUID"].(string); ok {
							disk.DiskUUID = v
						}

						// Get scheduled replicas
						if replicas, ok := diskStatus["scheduledReplica"].(map[string]interface{}); ok {
							disk.ScheduledReplica = make(map[string]int64)
							for repName, repSize := range replicas {
								if size, ok := repSize.(int64); ok {
									disk.ScheduledReplica[repName] = size
								} else if size, ok := repSize.(float64); ok {
									disk.ScheduledReplica[repName] = int64(size)
								}
							}
						}
					}
				}
				node.Disks[diskID] = disk
			}
		}
	}

	return node, nil
}

func nodeToUnstructured(node *Node) (*unstructured.Unstructured, error) {
	// This is a simplified version - in production you'd need to handle all fields
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "longhorn.io",
		Version: "v1beta2",
		Kind:    "Node",
	})
	u.SetName(node.Name)
	u.SetNamespace(node.Namespace)

	// Convert tags to []interface{} for unstructured
	tags := make([]interface{}, len(node.Tags))
	for i, tag := range node.Tags {
		tags[i] = tag
	}

	// Set spec fields
	spec := map[string]interface{}{
		"allowScheduling":   node.AllowScheduling,
		"evictionRequested": node.EvictionRequested,
		"tags":              tags,
	}

	if err := unstructured.SetNestedMap(u.Object, spec, "spec"); err != nil {
		return nil, err
	}

	return u, nil
}

// Implement other methods...
func (c *crdNodeClient) EnableScheduling(name string) error {
	_, err := c.Update(name, &NodeUpdate{AllowScheduling: &[]bool{true}[0]})
	return err
}

func (c *crdNodeClient) DisableScheduling(name string) error {
	_, err := c.Update(name, &NodeUpdate{AllowScheduling: &[]bool{false}[0]})
	return err
}

func (c *crdNodeClient) EvictNode(name string) error {
	_, err := c.Update(name, &NodeUpdate{EvictionRequested: &[]bool{true}[0]})
	return err
}

func (c *crdNodeClient) AddDisk(nodeName string, disk DiskUpdate) error {
	// This would need more complex implementation
	return fmt.Errorf("not implemented for CRD client")
}

func (c *crdNodeClient) RemoveDisk(nodeName, diskID string) error {
	return fmt.Errorf("not implemented for CRD client")
}

func (c *crdNodeClient) UpdateDiskTags(nodeName, diskID string, tags []string) error {
	return fmt.Errorf("not implemented for CRD client")
}

func (c *crdNodeClient) AddNodeTag(nodeName, tag string) error {
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
	_, err = c.Update(nodeName, &NodeUpdate{Tags: tags})
	return err
}

func (c *crdNodeClient) RemoveNodeTag(nodeName, tag string) error {
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

	_, err = c.Update(nodeName, &NodeUpdate{Tags: tags})
	return err
}
