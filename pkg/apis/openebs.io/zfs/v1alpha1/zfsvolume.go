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
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=zfsvol
// +kubebuilder:printcolumn:name="ZPool",type=string,JSONPath=`.spec.poolName`,description="ZFS Pool where the volume is created"
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.ownerNodeID`,description="Node where the volume is created"
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.capacity`,description="Size of the volume"
// +kubebuilder:printcolumn:name="volblocksize",type=string,JSONPath=`.spec.volblocksize`,description="volblocksize of volume"
// +kubebuilder:printcolumn:name="recordsize",type=string,JSONPath=`.spec.recordsize`,description="recordsize of created zfs dataset"
// +kubebuilder:printcolumn:name="Filesystem",type=string,JSONPath=`.spec.fsType`,description="filesystem created on the volume"
// +kubebuilder:printcolumn:name="CreationTime",type=date,JSONPath=`.status.creationTime`,description="Timestamp when the volume has been created."
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
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
// +resource:path=zfsvolumes

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

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	OwnerNodeID string `json:"ownerNodeID"`

	// poolName specifies the name of the
	// pool where this volume should be created
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	PoolName string `json:"poolName"`

	// SnapName specifies the name of the
	// snapshot where this volume should be cloned
	SnapName string `json:"snapname,omitempty"`

	// Capacity of the volume
	// +kubebuilder:validation:MinLength=1
	Capacity string `json:"capacity"`

	// RecordSize specifies the record size
	// for the zfs dataset
	// +kubebuilder:validation:MinLength=1
	RecordSize string `json:"recordsize,omitempty"`

	// VolBlockSize specifies the block size for the zvol
	// +kubebuilder:validation:MinLength=1
	VolBlockSize string `json:"volblocksize,omitempty"`

	// Controls the compression algorithm used for this dataset. Compression
	// specifies if the it should enabled on the zvol. Setting compression to on
	// indicates that the current default compression algorithm should be used.
	// The current default compression algorithm is either lzjb or, if the lz4_compress
	// feature is enabled, lz4.
	// Changing this property affects only newly-written data.
	// +kubebuilder:validation:Pattern="^(on|off|lzjb|gzip|gzip-[1-9]|zle|lz4)$"
	Compression string `json:"compression,omitempty"`

	// Deduplication is the process for removing redundant data at the block level,
	// reducing the total amount of data stored. If a file system has the dedup property
	// enabled, duplicate data blocks are removed synchronously.
	// The result is that only unique data is stored and common components are shared among files.
	// Deduplication can consume significant processing power (CPU) and memory as well as generate additional disk IO.
	// Before creating a pool with deduplication enabled, ensure that you have planned your hardware
	// requirements appropriately and implemented appropriate recovery practices, such as regular backups.
	// As an alternative to deduplication consider using compression=lz4, as a less resource-intensive alternative.
	// should be enabled on the zvol
	// +kubebuilder:validation:Enum=on;off
	Dedup string `json:"dedup,omitempty"`

	// Enabling the encryption feature allows for the creation of
	// encrypted filesystems and volumes. ZFS will encrypt file and zvol data,
	// file attributes, ACLs, permission bits, directory listings, FUID mappings,
	// and userused / groupused data. ZFS will not encrypt metadata related to the
	// pool structure, including dataset and snapshot names, dataset hierarchy,
	// properties, file size, file holes, and deduplication tables
	// (though the deduplicated data itself is encrypted).
	// +kubebuilder:validation:Pattern="^(on|off|aes-128-[c,g]cm|aes-192-[c,g]cm|aes-256-[c,g]cm)$"
	Encryption string `json:"encryption,omitempty"`

	// KeyLocation is the location of key for the encryption
	KeyLocation string `json:"keylocation,omitempty"`

	// KeyFormat specifies format of the encryption key
	KeyFormat string `json:"keyformat,omitempty"`

	// Thinprovision specifies if we should
	// thin provisioned the volume or not
	// +kubebuilder:validation:Enum=Yes;no
	ThinProvision string `json:"thinProvision,omitempty"`

	// volumeType determines whether the volume is of type "DATASET" or "ZVOL".
	// if fsttype provided in the storageclass as "zfs", then it will create a
	// volume of type "DATASET". If "ext4", "ext3", "ext2" or "xfs" in mentioned as fstype
	// in the storageclass, it will create a volume of type "ZVOL" so that it can be
	// further formatted with the fstype provided in the storageclass.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ZVOL;DATASET
	VolumeType string `json:"volumeType"`

	// FsType specifies filesystem type for the
	// zfs volume/dataset
	FsType string `json:"fsType,omitempty"`
}
