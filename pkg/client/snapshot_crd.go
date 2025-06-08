package client

// snapshot_crd.go contains placeholder implementations for snapshot operations using Kubernetes CRDs.
// At the moment these return a not implemented error to signal missing support.

import "fmt"

// crdSnapshotClient implements SnapshotInterface via Kubernetes CRDs
// TODO: implement actual CRD logic if needed

type crdSnapshotClient struct {
	crdClient *LonghornCRDClient
}

func (c *crdSnapshotClient) List(volumeName string) ([]Snapshot, error) {
	return nil, fmt.Errorf("snapshot list via CRD not implemented")
}

func (c *crdSnapshotClient) Create(volumeName string, input *SnapshotCreateInput) (*Snapshot, error) {
	return nil, fmt.Errorf("snapshot create via CRD not implemented")
}

func (c *crdSnapshotClient) Delete(volumeName, snapshotName string) error {
	return fmt.Errorf("snapshot delete via CRD not implemented")
}
