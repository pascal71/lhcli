// pkg/client/types.go
package client

import "time"

// Node represents a Longhorn node
type Node struct {
	Name                     string            `json:"name"`
	Namespace                string            `json:"namespace,omitempty"`
	Address                  string            `json:"address"`
	AllowScheduling          bool              `json:"allowScheduling"`
	EvictionRequested        bool              `json:"evictionRequested"`
	Conditions               map[string]Status `json:"conditions"`
	Disks                    map[string]Disk   `json:"disks"`
	Region                   string            `json:"region"`
	Zone                     string            `json:"zone"`
	Tags                     []string          `json:"tags"`
	EngineManagerCPURequest  int               `json:"engineManagerCPURequest"`
	ReplicaManagerCPURequest int               `json:"replicaManagerCPURequest"`
}

// NodeList represents a list of nodes
type NodeList struct {
	Items []Node `json:"data"`
}

// Status represents a condition status
type Status struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastProbeTime      string `json:"lastProbeTime"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
}

// Disk represents a disk on a node
type Disk struct {
	Path              string            `json:"path"`
	AllowScheduling   bool              `json:"allowScheduling"`
	EvictionRequested bool              `json:"evictionRequested"`
	StorageMaximum    int64             `json:"storageMaximum"`
	StorageAvailable  int64             `json:"storageAvailable"`
	StorageReserved   int64             `json:"storageReserved"`
	StorageScheduled  int64             `json:"storageScheduled"`
	DiskUUID          string            `json:"diskUUID"`
	DiskType          string            `json:"diskType"`
	Conditions        map[string]Status `json:"conditions"`
	ScheduledReplica  map[string]int64  `json:"scheduledReplica"`
	Tags              []string          `json:"tags"`
}

// NodeUpdate represents node update parameters
type NodeUpdate struct {
	AllowScheduling   *bool                 `json:"allowScheduling,omitempty"`
	EvictionRequested *bool                 `json:"evictionRequested,omitempty"`
	Tags              []string              `json:"tags,omitempty"`
	Disks             map[string]DiskUpdate `json:"disks,omitempty"`
}

// DiskUpdate represents disk update parameters
type DiskUpdate struct {
	Path              string   `json:"path"`
	AllowScheduling   *bool    `json:"allowScheduling,omitempty"`
	EvictionRequested *bool    `json:"evictionRequested,omitempty"`
	StorageReserved   int64    `json:"storageReserved,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

// Volume represents a Longhorn volume
// In pkg/client/types.go, update the Volume struct to include ActualSize:

// Volume represents a Longhorn volume
type Volume struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace,omitempty"`
	Size             string            `json:"size"`
	ActualSize       int64             `json:"actualSize,omitempty"` // Add this field
	NumberOfReplicas int               `json:"numberOfReplicas"`
	State            string            `json:"state"`
	Robustness       string            `json:"robustness"`
	Frontend         string            `json:"frontend"`
	DataLocality     string            `json:"dataLocality"`
	AccessMode       string            `json:"accessMode"`
	Migratable       bool              `json:"migratable"`
	Encrypted        bool              `json:"encrypted"`
	Image            string            `json:"image"`
	LastBackup       string            `json:"lastBackup"`
	LastBackupAt     string            `json:"lastBackupAt"`
	Created          string            `json:"created"`
	Conditions       map[string]Status `json:"conditions"`
	Replicas         []Replica         `json:"replicas"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// VolumeCreateInput represents volume creation parameters
type VolumeCreateInput struct {
	Name             string            `json:"name"`
	Size             string            `json:"size"`
	NumberOfReplicas int               `json:"numberOfReplicas"`
	Frontend         string            `json:"frontend"`
	DataLocality     string            `json:"dataLocality"`
	AccessMode       string            `json:"accessMode"`
	Migratable       bool              `json:"migratable"`
	Encrypted        bool              `json:"encrypted"`
	NodeSelector     []string          `json:"nodeSelector"`
	DiskSelector     []string          `json:"diskSelector"`
	RecurringJobs    []RecurringJob    `json:"recurringJobs"`
	Labels           map[string]string `json:"labels"`
}

// VolumeUpdateInput represents volume update parameters
type VolumeUpdateInput struct {
	NumberOfReplicas *int              `json:"numberOfReplicas,omitempty"`
	DataLocality     string            `json:"dataLocality,omitempty"`
	AccessMode       string            `json:"accessMode,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// VolumeAttachInput represents volume attach parameters
type VolumeAttachInput struct {
	HostID          string `json:"hostId"`
	DisableFrontend bool   `json:"disableFrontend"`
	AttachedBy      string `json:"attachedBy"`
}

// Controller represents a volume controller
type Controller struct {
	Name                string `json:"name"`
	NodeID              string `json:"nodeID"`
	Endpoint            string `json:"endpoint"`
	CurrentImage        string `json:"currentImage"`
	InstanceManagerName string `json:"instanceManagerName"`
}

// Replica represents a volume replica
type Replica struct {
	Name            string            `json:"name"`
	NodeID          string            `json:"nodeID"`
	DiskID          string            `json:"diskID"`
	DiskPath        string            `json:"diskPath,omitempty"` // Add this field
	DataPath        string            `json:"dataPath"`
	Mode            string            `json:"mode"`
	FailedAt        string            `json:"failedAt"`
	Running         bool              `json:"running"`
	SpecSize        string            `json:"specSize"`
	ActualSize      string            `json:"actualSize"`
	IP              string            `json:"ip"`
	Port            int               `json:"port"`
	InstanceManager string            `json:"instanceManager"`
	Image           string            `json:"image"`
	StorageIP       string            `json:"storageIP"`
	StoragePort     int               `json:"storagePort"`
	DataEngine      string            `json:"dataEngine"`
	Conditions      map[string]Status `json:"conditions"`
}

// Setting represents a Longhorn setting
type Setting struct {
	Name       string            `json:"name"`
	Value      string            `json:"value"`
	Default    string            `json:"default"`
	Definition SettingDefinition `json:"definition"`
}

// SettingDefinition represents setting metadata
type SettingDefinition struct {
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	ReadOnly    bool     `json:"readOnly"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

// Backup represents a volume backup
type Backup struct {
	Name            string            `json:"name"`
	State           string            `json:"state"`
	Progress        int               `json:"progress"`
	Error           string            `json:"error"`
	URL             string            `json:"url"`
	SnapshotName    string            `json:"snapshotName"`
	SnapshotCreated string            `json:"snapshotCreated"`
	Created         string            `json:"created"`
	Size            string            `json:"size"`
	Labels          map[string]string `json:"labels"`
	VolumeName      string            `json:"volumeName"`
	VolumeSize      string            `json:"volumeSize"`
	VolumeCreated   string            `json:"volumeCreated"`
}

// BackupCreateInput represents backup creation parameters
type BackupCreateInput struct {
	SnapshotName string            `json:"snapshotName"`
	Labels       map[string]string `json:"labels"`
}

// BackupTarget represents the backup target configuration
type BackupTarget struct {
	BackupTargetURL  string `json:"backupTargetURL"`
	CredentialSecret string `json:"credentialSecret"`
	Available        bool   `json:"available"`
	Message          string `json:"message"`
}

// EngineImage represents a Longhorn engine image
type EngineImage struct {
	Name              string            `json:"name"`
	Image             string            `json:"image"`
	Default           bool              `json:"default"`
	State             string            `json:"state"`
	RefCount          int               `json:"refCount"`
	Created           string            `json:"created"`
	NodeDeploymentMap map[string]bool   `json:"nodeDeploymentMap"`
	Conditions        map[string]Status `json:"conditions"`
}

// RecurringJob represents a recurring job configuration
type RecurringJob struct {
	Name        string            `json:"name"`
	Task        string            `json:"task"`
	Cron        string            `json:"cron"`
	Retain      int               `json:"retain"`
	Concurrency int               `json:"concurrency"`
	Labels      map[string]string `json:"labels"`
}

// Event represents a Longhorn event
type Event struct {
	Type           string    `json:"type"`
	Object         string    `json:"object"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
	Count          int       `json:"count"`
}

// EventListOptions represents event listing options
type EventListOptions struct {
	ResourceType string
	ResourceName string
	EventType    string
}

// ErrorResponse from Longhorn API
type ErrorResponse struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
