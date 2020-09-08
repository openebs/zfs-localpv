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

package bkpbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/common/errors"
)

// Builder is the builder object for ZFSBackup
type Builder struct {
	bkp  *ZFSBackup
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		bkp: &ZFSBackup{
			Object: &apis.ZFSBackup{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(bkp *apis.ZFSBackup) *Builder {
	if bkp == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build bkp object: nil bkp"),
		)
		return b
	}
	return &Builder{
		bkp: &ZFSBackup{
			Object: bkp,
		},
	}
}

// WithNamespace sets the namespace of  ZFSBackup
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi bkp object: missing namespace",
			),
		)
		return b
	}
	b.bkp.Object.Namespace = namespace
	return b
}

// WithName sets the name of ZFSBackup
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi bkp object: missing name",
			),
		)
		return b
	}
	b.bkp.Object.Name = name
	return b
}

// WithPrevSnap sets the previous snapshot for ZFSBackup
func (b *Builder) WithPrevSnap(snap string) *Builder {
	b.bkp.Object.Spec.PrevSnapName = snap
	return b
}

// WithSnap sets the snapshot for ZFSBackup
func (b *Builder) WithSnap(snap string) *Builder {
	b.bkp.Object.Spec.SnapName = snap
	return b
}

// WithVolume sets the volume name of ZFSBackup
func (b *Builder) WithVolume(volume string) *Builder {
	if volume == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi bkp object: missing volume name",
			),
		)
		return b
	}
	b.bkp.Object.Spec.VolumeName = volume
	return b
}

// WithNode sets the owenr node for the ZFSBackup
func (b *Builder) WithNode(node string) *Builder {
	if node == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build bkp object: missing node id",
			),
		)
		return b
	}
	b.bkp.Object.Spec.OwnerNodeID = node
	return b
}

// WithStatus sets the status of the Backup progress
func (b *Builder) WithStatus(status apis.ZFSBackupStatus) *Builder {
	if status == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build bkp object: missing snap name",
			),
		)
		return b
	}
	b.bkp.Object.Status = status
	return b
}

// WithRemote sets the remote address for the ZFSBackup
func (b *Builder) WithRemote(server string) *Builder {
	if server == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build bkp object: missing remote",
			),
		)
		return b
	}
	b.bkp.Object.Spec.BackupDest = server
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.bkp.Object.Labels == nil {
		b.bkp.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.bkp.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer merge existing finalizers if any
// with the ones that are provided here
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.bkp.Object.Finalizers = append(b.bkp.Object.Finalizers, finalizer...)
	return b
}

// Build returns ZFSBackup API object
func (b *Builder) Build() (*apis.ZFSBackup, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.bkp.Object, nil
}
