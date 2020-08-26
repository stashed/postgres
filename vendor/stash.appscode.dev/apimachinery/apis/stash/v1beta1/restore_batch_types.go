/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindRestoreBatch     = "RestoreBatch"
	ResourceSingularRestoreBatch = "restorebatch"
	ResourcePluralRestoreBatch   = "restorebatches"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=restorebatches,singular=restorebatch,categories={stash,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Repository",type="string",JSONPath=".spec.repository.name"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RestoreBatch struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              RestoreBatchSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            RestoreBatchStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type RestoreBatchSpec struct {
	// Driver indicates the name of the agent to use to restore the target.
	// Supported values are "Restic", "VolumeSnapshotter".
	// Default value is "Restic".
	// +optional
	// +kubebuilder:default=Restic
	Driver Snapshotter `json:"driver,omitempty" protobuf:"bytes,1,opt,name=driver,casttype=Snapshotter"`
	// Repository refer to the Repository crd that holds backend information
	// +optional
	Repository core.LocalObjectReference `json:"repository,omitempty" protobuf:"bytes,2,opt,name=repository"`
	// Members is a list of restore targets and their configuration that are part of this batch
	// +optional
	Members []RestoreTargetSpec `json:"members,omitempty" protobuf:"bytes,3,rep,name=members"`
	// ExecutionOrder indicate whether to restore the members in the sequential order as they appear in the members list.
	// The default value is "Parallel" which means the members will be restored in parallel.
	// +kubebuilder:default=Parallel
	// +optional
	ExecutionOrder ExecutionOrder `json:"executionOrder,omitempty" protobuf:"bytes,4,opt,name=executionOrder"`
	// Hooks specifies the actions that Stash should take before or after restore.
	// Cannot be updated.
	// +optional
	Hooks *RestoreHooks `json:"hooks,omitempty" protobuf:"bytes,5,opt,name=hooks"`
}

type RestoreBatchStatus struct {
	// Phase indicates the overall phase of the restore process for this RestoreBatch. Phase will be "Succeeded" only if
	// phase of all members are "Succeeded". If the restore process fail for any of the members, Phase will be "Failed".
	// +optional
	Phase RestorePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=RestorePhase"`
	// SessionDuration specify total time taken to complete restore of all the members.
	// +optional
	SessionDuration string `json:"sessionDuration,omitempty" protobuf:"bytes,2,opt,name=sessionDuration"`
	// Conditions shows the condition of different steps for the RestoreBatch.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// Members shows the restore status for the members of the RestoreBatch.
	// +optional
	Members []RestoreMemberStatus `json:"members,omitempty" protobuf:"bytes,4,rep,name=members"`
}

// +kubebuilder:validation:Enum=Pending;Succeeded;Running;Failed
type RestoreTargetPhase string

const (
	TargetRestorePending      RestoreTargetPhase = "Pending"
	TargetRestoreRunning      RestoreTargetPhase = "Running"
	TargetRestoreSucceeded    RestoreTargetPhase = "Succeeded"
	TargetRestoreFailed       RestoreTargetPhase = "Failed"
	TargetRestorePhaseUnknown RestoreTargetPhase = "Unknown"
)

type RestoreMemberStatus struct {
	// Ref is the reference to the respective target whose status is shown here.
	Ref TargetRef `json:"ref" protobuf:"bytes,1,opt,name=ref"`
	// Conditions shows the condition of different steps to restore this member.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,2,rep,name=conditions"`
	// TotalHosts specifies total number of hosts that will be restored for this member.
	// +optional
	TotalHosts *int32 `json:"totalHosts,omitempty" protobuf:"varint,3,opt,name=totalHosts"`
	// Phase indicates restore phase of this member
	// +optional
	Phase RestoreTargetPhase `json:"phase,omitempty" protobuf:"bytes,4,opt,name=phase"`
	// Stats shows restore statistics of individual hosts for this member
	// +optional
	Stats []HostRestoreStats `json:"stats,omitempty" protobuf:"bytes,5,rep,name=stats"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RestoreBatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []RestoreBatch `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
