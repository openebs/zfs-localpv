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

package snapbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/common/errors"
)

// Builder is the builder object for ZFSSnapshot
type Builder struct {
	snap *ZFSSnapshot
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		snap: &ZFSSnapshot{
			Object: &apis.ZFSSnapshot{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(snap *apis.ZFSSnapshot) *Builder {
	if snap == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build snap object: nil snap"),
		)
		return b
	}
	return &Builder{
		snap: &ZFSSnapshot{
			Object: snap,
		},
	}
}

// WithNamespace sets the namespace of  ZFSSnapshot
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi snap object: missing namespace",
			),
		)
		return b
	}
	b.snap.Object.Namespace = namespace
	return b
}

// WithName sets the name of ZFSSnapshot
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi snap object: missing name",
			),
		)
		return b
	}
	b.snap.Object.Name = name
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.snap.Object.Labels == nil {
		b.snap.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.snap.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer merge existing finalizers if any
// with the ones that are provided here
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.snap.Object.Finalizers = append(b.snap.Object.Finalizers, finalizer...)
	return b
}

// Build returns ZFSSnapshot API object
func (b *Builder) Build() (*apis.ZFSSnapshot, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.snap.Object, nil
}
