/*
Copyright 2021 The OpenEBS Authors

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsnode

// ZFSNode records information about all zfs pools available
// in a node. In general, the openebs node-agent creates the ZFSNode
// object & periodically synchronizing the zfs pools available in the node.
// ZFSNode has an owner reference pointing to the corresponding node object.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=zfsnode
type ZFSNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Pools []Pool `json:"pools"`
}

// Pool specifies attributes of a given zfs pool that exists on the node.
type Pool struct {
	// Name of the zfs pool.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// UUID denotes a unique identity of a zfs pool.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	UUID string `json:"uuid"`

	// Free specifies the available capacity of zfs pool.
	// +kubebuilder:validation:Required
	Free resource.Quantity `json:"free"`
}

// ZFSNodeList is a collection of ZFSNode resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsnodes
type ZFSNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSNode `json:"items"`
}
