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

// Package promotebuilder builds ZFSPromote objects.
package promotebuilder

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
)

// Builder is the builder object for ZFSPromote
type Builder struct {
	promote *apis.ZFSPromote
	errs    []error
}

// NewBuilder returns a builder for a new ZFSPromote with status Init
func NewBuilder() *Builder {
	return &Builder{
		promote: &apis.ZFSPromote{Status: apis.PromoteZFSStatusInit},
	}
}

// BuildFrom returns a builder for a copy of the given ZFSPromote
func BuildFrom(promote *apis.ZFSPromote) *Builder {
	if promote == nil {
		b := NewBuilder()
		b.errs = append(b.errs, errors.New("failed to build promote object: nil promote"))
		return b
	}
	return &Builder{promote: promote.DeepCopy()}
}

// WithName sets the name of the ZFSPromote. A volume is promoted at most
// once, so callers name it after the volume.
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(b.errs, errors.New("failed to build promote object: missing name"))
		return b
	}
	b.promote.Name = name
	return b
}

// WithNamespace sets the namespace of the ZFSPromote
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(b.errs, errors.New("failed to build promote object: missing namespace"))
		return b
	}
	b.promote.Namespace = namespace
	return b
}

// WithVolume sets the name of the ZFSVolume to promote
func (b *Builder) WithVolume(name string) *Builder {
	if name == "" {
		b.errs = append(b.errs, errors.New("failed to build promote object: missing volume name"))
		return b
	}
	b.promote.Spec.VolumeName = name
	return b
}

// WithNode sets the node id of the volume, whose node agent does the promote
func (b *Builder) WithNode(node string) *Builder {
	if node == "" {
		b.errs = append(b.errs, errors.New("failed to build promote object: missing node id"))
		return b
	}
	b.promote.Spec.OwnerNodeID = node
	return b
}

// WithStatus sets the status of the ZFSPromote
func (b *Builder) WithStatus(status apis.ZFSPromoteStatus) *Builder {
	b.promote.Status = status
	return b
}

// Build returns the ZFSPromote API object
func (b *Builder) Build() (*apis.ZFSPromote, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.promote, nil
}
