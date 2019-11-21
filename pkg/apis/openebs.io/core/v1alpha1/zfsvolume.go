/*
Copyright © 2019 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=zfsvolume

// ZFSVolume represents a ZFS based volume
type ZFSVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeInfo `json:"spec"`
}

// MountInfo contains the volume related info
// for all types of volumes in ZFSVolume
type MountInfo struct {
	// FSType of a volume will specify the
	// format type - ext4(default), xfs of PV
	FSType string `json:"fsType"`

	// AccessMode of a volume will hold the
	// access mode of the volume
	AccessModes []string `json:"accessModes"`

	// MountPath of the volume will hold the
	// path on which the volume is mounted
	// on that node
	MountPath string `json:"mountPath"`

	// ReadOnly specifies if the volume needs
	// to be mounted in ReadOnly mode
	ReadOnly bool `json:"readOnly"`

	// MountOptions specifies the options with
	// which mount needs to be attempted
	MountOptions []string `json:"mountOptions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=csivolumes

// ZFSVolumeList is a list of ZFSVolume resources
type ZFSVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ZFSVolume `json:"items"`
}

// VolumeInfo contains the volume related info
// for all types of volumes in ZFSVolume
type VolumeInfo struct {
	// OwnerNodeID is the Node ID which
	// is the owner of this Volume
	OwnerNodeID string `json:"ownerNodeID"`

	// poolName specifies the name of the
	// pool where this volume should be created
	PoolName string `json:"poolName"`

	// Capacity of the volume
	Capacity string `json:"capacity"`

	// RecordSize specifies the record size
	// for the zfs dataset
	RecordSize string `json:"recordsize,omitempty"`

	// VolBlockSize specifies the block size for the zvol
	VolBlockSize string `json:"volblocksize,omitempty"`

	// Compression specifies if the it should
	// enabled on the zvol
	Compression string `json:"compression,omitempty"`

	// Dedup specifies the deduplication
	// should be enabled on the zvol
	Dedup string `json:"dedup,omitempty"`

	// Encryption specifies the encryption
	// should be enabled on the zvol
	Encryption string `json:"encryption,omitempty"`

	// KeyLocation is the location of key
	// for the encryption
	KeyLocation string `json:"keylocation,omitempty"`

	// KeyFormat specifies format of the
	// encryption key
	KeyFormat string `json:"keyformat,omitempty"`

	// Thinprovision specifies if we should
	// thin provisioned the volume or not
	ThinProvision string `json:"thinProvision,omitempty"`

	// VolumeType specifies whether the volume is
	// zvol or a dataset
	VolumeType string `json:"volumeType"`

	// FsType specifies filesystem type for the
	// zfs volume/dataset
	FsType string `json:"fsType,omitempty"`
}
