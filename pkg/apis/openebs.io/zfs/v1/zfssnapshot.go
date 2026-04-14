/*
Copyright 2020 The OpenEBS Authors

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
// +resource:path=zfssnapshot

// ZFSSnapshot represents a ZFS Snapshot of the zfsvolume
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=zfssnap
type ZFSSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeInfo `json:"spec"`
	Status SnapStatus `json:"status"`
}

// ZFSSnapshotList is a list of ZFSSnapshot resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfssnapshots
type ZFSSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSSnapshot `json:"items"`
}

// SnapStatus string that reflects if the snapshot was created successfully
type SnapStatus struct {
	// State reflects whether the snapshot was created successfully.
	// The state "Failed" is accompanied by a human-readable Message.
	State string `json:"state,omitempty"`

	// Message is a human-readable description of why the snapshot is in the
	// current state. It is populated when State transitions to Failed and
	// contains the underlying ZFS error so that operators can diagnose the
	// failure without inspecting node-agent logs directly.
	// +kubebuilder:validation:MaxLength=1024
	Message string `json:"message,omitempty"`
}
