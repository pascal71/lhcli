// pkg/client/replica_crd.go
package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// replicaClient implementation for CRDs
type crdReplicaClient struct {
	crdClient *LonghornCRDClient
}

// List returns all Longhorn replicas
func (c *crdReplicaClient) List() ([]Replica, error) {
	debugLog("Listing Longhorn replicas via CRD")

	list, err := c.crdClient.dynamicClient.Resource(replicaGVR).
		Namespace(c.crdClient.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list replicas: %w", err)
	}

	replicas := make([]Replica, 0, len(list.Items))
	for _, item := range list.Items {
		replica, volumeName, err := unstructuredToReplica(&item)
		if err != nil {
			debugLog("Failed to convert replica %s: %v", item.GetName(), err)
			continue
		}
		// Set the volume name from the conversion
		if replica != nil {
			replica.VolumeName = volumeName
			replicas = append(replicas, *replica)
		}
	}

	return replicas, nil
}

// Get returns a specific replica
func (c *crdReplicaClient) Get(name string) (*Replica, error) {
	debugLog("Getting Longhorn replica %s via CRD", name)

	unstructuredReplica, err := c.crdClient.dynamicClient.Resource(replicaGVR).
		Namespace(c.crdClient.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get replica %s: %w", name, err)
	}

	replica, volumeName, err := unstructuredToReplica(unstructuredReplica)
	if err != nil {
		return nil, err
	}
	if replica != nil {
		replica.VolumeName = volumeName
	}

	return replica, nil
}

// Delete deletes a replica
func (c *crdReplicaClient) Delete(name string) error {
	debugLog("Deleting Longhorn replica %s via CRD", name)

	err := c.crdClient.dynamicClient.Resource(replicaGVR).
		Namespace(c.crdClient.namespace).
		Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete replica %s: %w", name, err)
	}

	return nil
}
