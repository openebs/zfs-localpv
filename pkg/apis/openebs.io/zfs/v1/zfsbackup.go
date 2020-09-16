/*
Copyright 2020 The OpenEBS Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsbackup

// ZFSBackup describes a zfs backup resource created as a custom resource
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,shortName=zb
// +kubebuilder:printcolumn:name="PrevSnap",type=string,JSONPath=`.spec.prevSnapName`,description="Previous snapshot for backup"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status`,description="Backup status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Age of the volume"
type ZFSBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ZFSBackupSpec `json:"spec"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Init;Done;Failed;Pending;InProgress;Invalid
	Status ZFSBackupStatus `json:"status"`
}

// ZFSBackupSpec is the spec for a ZFSBackup resource
type ZFSBackupSpec struct {
	// VolumeName is a name of the volume for which this backup is destined
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolumeName string `json:"volumeName"`

	// OwnerNodeID is a name of the nodes where the source volume is
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	OwnerNodeID string `json:"ownerNodeID"`

	// SnapName is the snapshot name for backup
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	SnapName string `json:"snapName,omitempty"`

	// PrevSnapName is the last completed-backup's snapshot name
	PrevSnapName string `json:"prevSnapName,omitempty"`

	// BackupDest is the remote address for backup transfer
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern="^([0-9]+.[0-9]+.[0-9]+.[0-9]+:[0-9]+)$"
	BackupDest string `json:"backupDest"`
}

// ZFSBackupStatus is to hold status of backup
type ZFSBackupStatus string

// Status written onto ZFSBackup objects.
const (
	// BKPZFSStatusDone , backup is completed.
	BKPZFSStatusDone ZFSBackupStatus = "Done"

	// BKPZFSStatusFailed , backup is failed.
	BKPZFSStatusFailed ZFSBackupStatus = "Failed"

	// BKPZFSStatusInit , backup is initialized.
	BKPZFSStatusInit ZFSBackupStatus = "Init"

	// BKPZFSStatusPending , backup is pending.
	BKPZFSStatusPending ZFSBackupStatus = "Pending"

	// BKPZFSStatusInProgress , backup is in progress.
	BKPZFSStatusInProgress ZFSBackupStatus = "InProgress"

	// BKPZFSStatusInvalid , backup operation is invalid.
	BKPZFSStatusInvalid ZFSBackupStatus = "Invalid"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsbackups

// ZFSBackupList is a list of ZFSBackup resources
type ZFSBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSBackup `json:"items"`
}
