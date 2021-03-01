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
// +resource:path=zfsrestore

// ZFSRestore describes a cstor restore resource created as a custom resource
type ZFSRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"` // set name to restore name + volume name + something like csp tag
	Spec              ZFSRestoreSpec              `json:"spec"`
	VolSpec           VolumeInfo                  `json:"volSpec,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Init;Done;Failed;Pending;InProgress;Invalid
	Status ZFSRestoreStatus `json:"status"`
}

// ZFSRestoreSpec is the spec for a ZFSRestore resource
type ZFSRestoreSpec struct {
	// volume name to where restore has to be performed
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolumeName string `json:"volumeName"`
	// owner node name where restore volume is present
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	OwnerNodeID string `json:"ownerNodeID"`

	// it can be ip:port in case of restore from remote or volumeName in case of local restore
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern="^([0-9]+.[0-9]+.[0-9]+.[0-9]+:[0-9]+)$"
	RestoreSrc string `json:"restoreSrc"`
}

// ZFSRestoreStatus is to hold result of action.
type ZFSRestoreStatus string

// Status written onto CStrorRestore object.
const (
	// RSTZFSStatusDone , restore operation is completed.
	RSTZFSStatusDone ZFSRestoreStatus = "Done"

	// RSTZFSStatusFailed , restore operation is failed.
	RSTZFSStatusFailed ZFSRestoreStatus = "Failed"

	// RSTZFSStatusInit , restore operation is initialized.
	RSTZFSStatusInit ZFSRestoreStatus = "Init"

	// RSTZFSStatusPending , restore operation is pending.
	RSTZFSStatusPending ZFSRestoreStatus = "Pending"

	// RSTZFSStatusInProgress , restore operation is in progress.
	RSTZFSStatusInProgress ZFSRestoreStatus = "InProgress"

	// RSTZFSStatusInvalid , restore operation is invalid.
	RSTZFSStatusInvalid ZFSRestoreStatus = "Invalid"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsrestores

// ZFSRestoreList is a list of ZFSRestore resources
type ZFSRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSRestore `json:"items"`
}
