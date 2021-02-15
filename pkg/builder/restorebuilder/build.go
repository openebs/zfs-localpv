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

package restorebuilder

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
)

// Builder is the builder object for ZFSRestore
type Builder struct {
	rstr *ZFSRestore
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		rstr: &ZFSRestore{
			Object: &apis.ZFSRestore{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(rstr *apis.ZFSRestore) *Builder {
	if rstr == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build rstr object: nil rstr"),
		)
		return b
	}
	return &Builder{
		rstr: &ZFSRestore{
			Object: rstr,
		},
	}
}

// WithNamespace sets the namespace of  ZFSRestore
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi rstr object: missing namespace",
			),
		)
		return b
	}
	b.rstr.Object.Namespace = namespace
	return b
}

// WithName sets the name of ZFSRestore
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi rstr object: missing name",
			),
		)
		return b
	}
	b.rstr.Object.Name = name
	return b
}

// WithVolume sets the name of ZFSRestore
func (b *Builder) WithVolume(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi rstr object: missing volume name",
			),
		)
		return b
	}
	b.rstr.Object.Spec.VolumeName = name
	return b
}

// WithVolSpec copies volume spec to ZFSRestore Object
func (b *Builder) WithVolSpec(vspec apis.VolumeInfo) *Builder {
	b.rstr.Object.VolSpec = vspec
	return b
}

// WithNode sets the node id for ZFSRestore
func (b *Builder) WithNode(node string) *Builder {
	if node == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi rstr object: missing node name",
			),
		)
		return b
	}
	b.rstr.Object.Spec.OwnerNodeID = node
	return b
}

// WithStatus sets the status for ZFSRestore
func (b *Builder) WithStatus(status apis.ZFSRestoreStatus) *Builder {
	b.rstr.Object.Status = status
	return b
}

// WithRemote sets the node id for ZFSRestore
func (b *Builder) WithRemote(server string) *Builder {
	if server == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi rstr object: missing node name",
			),
		)
		return b
	}
	b.rstr.Object.Spec.RestoreSrc = server
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.rstr.Object.Labels == nil {
		b.rstr.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.rstr.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer merge existing finalizers if any
// with the ones that are provided here
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.rstr.Object.Finalizers = append(b.rstr.Object.Finalizers, finalizer...)
	return b
}

// Build returns ZFSRestore API object
func (b *Builder) Build() (*apis.ZFSRestore, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.rstr.Object, nil
}
