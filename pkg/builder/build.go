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

package builder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
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

// WithNamespace sets the namespace of csi volume
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing namespace",
			),
		)
		return b
	}
	b.volume.Object.Namespace = namespace
	return b
}

// WithName sets the name of csi volume
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing name",
			),
		)
		return b
	}
	b.volume.Object.Name = name
	return b
}

// WithCapacity sets the Capacity of csi volume by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	if capacity == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing capacity",
			),
		)
		return b
	}
	b.volume.Object.Spec.Capacity = capacity
	return b
}

// WithCompression sets compression of CStorVolumeClaim
func (b *Builder) WithCompression(compression string) *Builder {

	comp := "off"
	if compression == "on" {
		comp = "on"
	}
	b.volume.Object.Spec.Compression = comp
	return b
}

// WithDedup sets compression of CStorVolumeClaim
func (b *Builder) WithDedup(dedup string) *Builder {

	dp := "off"
	if dedup == "on" {
		dp = "on"
	}
	b.volume.Object.Spec.Dedup = dp
	return b
}

// WithThinProv sets compression of CStorVolumeClaim
func (b *Builder) WithThinProv(thinprov string) *Builder {

	tp := "no"
	if thinprov == "yes" {
		tp = "yes"
	}
	b.volume.Object.Spec.ThinProvision = tp
	return b
}

// WithBlockSize sets blocksize of CStorVolumeClaim
func (b *Builder) WithBlockSize(blockSize string) *Builder {

	bs := "4k"
	if len(blockSize) > 0 {
		bs = blockSize
	}
	b.volume.Object.Spec.BlockSize = bs
	return b
}

func (b *Builder) WithPoolName(pool string) *Builder {
	if pool == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing pool name",
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
				"failed to build csi volume object: missing node name",
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
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolume object: missing labels"),
		)
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

// Build returns csi volume API object
func (b *Builder) Build() (*apis.ZFSVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
