/*
Copyright 2019 The OpenEBS Authors

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

package volbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1alpha1"
	"github.com/openebs/zfs-localpv/pkg/common/errors"
)

// Builder is the builder object for ZFSVolume
type Builder struct {
	volume *ZFSVolume
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		volume: &ZFSVolume{
			Object: &apis.ZFSVolume{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(volume *apis.ZFSVolume) *Builder {
	if volume == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume"),
		)
		return b
	}
	return &Builder{
		volume: &ZFSVolume{
			Object: volume,
		},
	}
}

// WithNamespace sets the namespace of  ZFSVolume
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs volume object: missing namespace",
			),
		)
		return b
	}
	b.volume.Object.Namespace = namespace
	return b
}

// WithName sets the name of ZFSVolume
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs volume object: missing name",
			),
		)
		return b
	}
	b.volume.Object.Name = name
	return b
}

// WithCapacity sets the Capacity of zfs volume by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	if capacity == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs volume object: missing capacity",
			),
		)
		return b
	}
	b.volume.Object.Spec.Capacity = capacity
	return b
}

// WithEncryption sets the encryption on ZFSVolume
func (b *Builder) WithEncryption(encr string) *Builder {
	b.volume.Object.Spec.Encryption = encr
	return b
}

// WithKeyLocation sets the encryption key location on ZFSVolume
func (b *Builder) WithKeyLocation(kl string) *Builder {
	b.volume.Object.Spec.KeyLocation = kl
	return b
}

// WithKeyFormat sets the encryption key format on ZFSVolume
func (b *Builder) WithKeyFormat(kf string) *Builder {
	b.volume.Object.Spec.KeyFormat = kf
	return b
}

// WithCompression sets compression of ZFSVolume
func (b *Builder) WithCompression(compression string) *Builder {
	b.volume.Object.Spec.Compression = compression
	return b
}

// WithDedup sets dedup property of ZFSVolume
func (b *Builder) WithDedup(dedup string) *Builder {
	b.volume.Object.Spec.Dedup = dedup
	return b
}

// WithThinProv sets if ZFSVolume needs to be thin provisioned
func (b *Builder) WithThinProv(thinprov string) *Builder {
	b.volume.Object.Spec.ThinProvision = thinprov
	return b
}

// WithOwnerNode sets owner node for the ZFSVolume where the volume should be provisioned
func (b *Builder) WithOwnerNode(host string) *Builder {
	b.volume.Object.Spec.OwnerNodeID = host
	return b
}

// WithRecordSize sets the recordsize of ZFSVolume
func (b *Builder) WithRecordSize(rs string) *Builder {
	b.volume.Object.Spec.RecordSize = rs
	return b
}

// WithVolBlockSize sets the volblocksize of ZFSVolume
func (b *Builder) WithVolBlockSize(bs string) *Builder {
	b.volume.Object.Spec.VolBlockSize = bs
	return b
}

// WithVolumeType sets if ZFSVolume needs to be thin provisioned
func (b *Builder) WithVolumeType(vtype string) *Builder {
	b.volume.Object.Spec.VolumeType = vtype
	return b
}

// WithVolumeStatus sets ZFSVolume status
func (b *Builder) WithVolumeStatus(status string) *Builder {
	b.volume.Object.Status.State = status
	return b
}

// WithFsType sets filesystem for the ZFSVolume
func (b *Builder) WithFsType(fstype string) *Builder {
	b.volume.Object.Spec.FsType = fstype
	return b
}

// WithSnapshot sets Snapshot name for creating clone volume
func (b *Builder) WithSnapshot(snap string) *Builder {
	b.volume.Object.Spec.SnapName = snap
	return b
}

func (b *Builder) WithPoolName(pool string) *Builder {
	if pool == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs volume object: missing pool name",
			),
		)
		return b
	}
	b.volume.Object.Spec.PoolName = pool
	return b
}

func (b *Builder) WithNodename(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build zfs volume object: missing node name",
			),
		)
		return b
	}
	b.volume.Object.Spec.OwnerNodeID = name
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.volume.Object.Labels == nil {
		b.volume.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.volume.Object.Labels[key] = value
	}
	return b
}

func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.volume.Object.Finalizers = append(b.volume.Object.Finalizers, finalizer...)
	return b
}

// Build returns ZFSVolume API object
func (b *Builder) Build() (*apis.ZFSVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
