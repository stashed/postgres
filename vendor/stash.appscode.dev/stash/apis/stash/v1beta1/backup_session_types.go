package v1beta1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindBackupSession     = "BackupSession"
	ResourceSingularBackupSession = "backupsession"
	ResourcePluralBackupSession   = "backupsessions"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackupSessionSpec   `json:"spec,omitempty"`
	Status            BackupSessionStatus `json:"status,omitempty"`
}

type BackupSessionSpec struct {
	// BackupConfiguration indicates the target BackupConfiguration crd
	BackupConfiguration core.LocalObjectReference `json:"backupConfiguration,omitempty"`
}

type BackupSessionPhase string

const (
	BackupSessionPending   BackupSessionPhase = "Pending"
	BackupSessionRunning   BackupSessionPhase = "Running"
	BackupSessionSucceeded BackupSessionPhase = "Succeeded"
	BackupSessionFailed    BackupSessionPhase = "Failed"
	BackupSessionSkipped   BackupSessionPhase = "Skipped"
	BackupSessionUnknown   BackupSessionPhase = "Unknown"
)

type HostBackupPhase string

const (
	HostBackupSucceeded HostBackupPhase = "Succeeded"
	HostBackupFailed    HostBackupPhase = "Failed"
)

type BackupSessionStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`
	// Phase indicates the overall phase of the backup process for this BackupSession. Phase will be "Succeeded" only if
	// phase of all hosts are "Succeeded". If any of the host fail to complete backup, Phase will be "Failed".
	// +optional
	Phase BackupSessionPhase `json:"phase,omitempty"`
	// TotalHosts specifies total number of hosts that will be backed up for this BackupSession
	// +Optional
	TotalHosts *int32 `json:"totalHosts,omitempty"`
	// SessionDuration specify total time taken to complete current backup session (sum of backup duration of all hosts)
	// +optional
	SessionDuration string `json:"sessionDuration,omitempty"`
	// Stats shows statistics of individual hosts for this backup session
	// +optional
	Stats []HostBackupStats `json:"stats,omitempty"`
}

type HostBackupStats struct {
	// Hostname indicate name of the host that has been backed up
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// Phase indicates backup phase of this host
	// +optional
	Phase HostBackupPhase `json:"phase,omitempty"`
	// Snapshots specifies the stats of individual snapshots that has been taken for this host in current backup session
	// +optional
	Snapshots []SnapshotStats `json:"snapshots,omitempty"`
	// Duration indicates total time taken to complete backup for this hosts
	// +optional
	Duration string `json:"duration,omitempty"`
	// Error indicates string value of error in case of backup failure
	// +optional
	Error string `json:"error,omitempty"`
}

type SnapshotStats struct {
	// Name indicates the name of the backup snapshot created for this host
	Name string `json:"name,omitempty"`
	// Directory indicates the directory that has been backed up in this snapshot
	Directory string `json:"directory,omitempty"`
	// Size indicates the size of data to backup in target directory
	Size string `json:"size,omitempty"`
	// Uploaded indicates size of data uploaded to backend for this snapshot
	Uploaded string `json:"uploaded,omitempty"`
	// ProcessingTime indicates time taken to process the target data
	ProcessingTime string `json:"processingTime,omitempty"`
	// FileStats shows statistics of files of this snapshot
	FileStats FileStats `json:"fileStats,omitempty"`
}

type FileStats struct {
	// TotalFiles shows total number of files that has been backed up
	TotalFiles *int `json:"totalFiles,omitempty"`
	// NewFiles shows total number of new files that has been created since last backup
	NewFiles *int `json:"newFiles,omitempty"`
	// ModifiedFiles shows total number of files that has been modified since last backup
	ModifiedFiles *int `json:"modifiedFiles,omitempty"`
	// UnmodifiedFiles shows total number of files that has not been changed since last backup
	UnmodifiedFiles *int `json:"unmodifiedFiles,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BackupSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupSession `json:"items,omitempty"`
}
