package types

import "time"

// Common types used across the CLI

// VolumeSpec represents volume creation parameters
type VolumeSpec struct {
    Name         string
    Size         string
    ReplicaCount int
    Frontend     string
    AccessMode   string
    Labels       map[string]string
}

// SnapshotSpec represents snapshot creation parameters
type SnapshotSpec struct {
    VolumeName string
    Name       string
    Labels     map[string]string
}

// BackupSpec represents backup creation parameters
type BackupSpec struct {
    VolumeName   string
    SnapshotName string
    Labels       map[string]string
}

// Event represents a Longhorn event
type Event struct {
    Type      string
    Reason    string
    Message   string
    Object    string
    Timestamp time.Time
}

// DiskSpec represents disk configuration
type DiskSpec struct {
    Path              string
    AllowScheduling   bool
    StorageReserved   string
    Tags              []string
}

// EngineImage represents a Longhorn engine image
type EngineImage struct {
    Name       string
    State      string
    RefCount   int
    Created    time.Time
    NodeDeploymentMap map[string]bool
}
