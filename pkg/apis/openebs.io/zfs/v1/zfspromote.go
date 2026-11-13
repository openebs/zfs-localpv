/*
Copyright 2026 The OpenEBS Authors.

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
// +resource:path=zfspromote

// ZFSPromote asks the node agent to `zfs promote` a cloned volume, so that
// it takes over its origin's snapshots and the origin becomes its clone.
// The caller creates it with status Init, waits for Done or Failed and
// then deletes it.
type ZFSPromote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ZFSPromoteSpec `json:"spec"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Init;Done;Failed;Pending;InProgress;Invalid
	Status ZFSPromoteStatus `json:"status"`
}

// ZFSPromoteSpec is the spec for a ZFSPromote resource
type ZFSPromoteSpec struct {
	// name of the ZFSVolume to promote
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolumeName string `json:"volumeName"`
	// node id of the ZFSVolume; only the node agent of that node acts on it
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	OwnerNodeID string `json:"ownerNodeID"`
}

// ZFSPromoteStatus is to hold result of action.
type ZFSPromoteStatus string

// Status written onto ZFSPromote object.
const (
	// PromoteZFSStatusDone , the volume has no origin and every moved
	// snapshot record carries its label.
	PromoteZFSStatusDone ZFSPromoteStatus = "Done"

	// PromoteZFSStatusFailed , promote operation failed; the reason is
	// in the events of the ZFSPromote.
	PromoteZFSStatusFailed ZFSPromoteStatus = "Failed"

	// PromoteZFSStatusInit , promote operation is requested.
	PromoteZFSStatusInit ZFSPromoteStatus = "Init"

	// PromoteZFSStatusPending , promote operation is pending.
	PromoteZFSStatusPending ZFSPromoteStatus = "Pending"

	// PromoteZFSStatusInProgress , the node agent is promoting the volume.
	PromoteZFSStatusInProgress ZFSPromoteStatus = "InProgress"

	// PromoteZFSStatusInvalid , promote operation is invalid.
	PromoteZFSStatusInvalid ZFSPromoteStatus = "Invalid"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfspromotes

// ZFSPromoteList is a list of ZFSPromote resources
type ZFSPromoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSPromote `json:"items"`
}
