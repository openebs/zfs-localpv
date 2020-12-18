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

package k8svolume

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
)

// Builder is the builder object for Volume
type Builder struct {
	volume *Volume
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{volume: &Volume{object: &corev1.Volume{}}}
}

// WithName sets the Name field of Volume with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Volume object: missing Volume name"),
		)
		return b
	}
	b.volume.object.Name = name
	return b
}

// WithHostDirectory sets the VolumeSource field of Volume with provided hostpath
// as type directory.
func (b *Builder) WithHostDirectory(path string) *Builder {
	if len(path) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing volume path"),
		)
		return b
	}
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
		},
	}

	b.volume.object.VolumeSource = volumeSource
	return b
}

// WithHostPathAndType sets the VolumeSource field of Volume with provided
// hostpath as directory path and type as directory type
func (b *Builder) WithHostPathAndType(
	dirpath string,
	dirtype *corev1.HostPathType,
) *Builder {
	if dirtype == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume type"),
		)
		return b
	}
	if len(dirpath) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing volume path"),
		)
		return b
	}
	newdirtype := *dirtype
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: dirpath,
			Type: &newdirtype,
		},
	}

	b.volume.object.VolumeSource = volumeSource
	return b
}

// WithPVCSource sets the Volume field of Volume with provided pvc
func (b *Builder) WithPVCSource(pvcName string) *Builder {
	if len(pvcName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing pvc name"),
		)
		return b
	}
	volumeSource := corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: pvcName,
		},
	}
	b.volume.object.VolumeSource = volumeSource
	return b
}

// WithEmptyDir sets the EmptyDir field of the Volume with provided dir
func (b *Builder) WithEmptyDir(dir *corev1.EmptyDirVolumeSource) *Builder {
	if dir == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil dir"),
		)
		return b
	}

	newdir := *dir
	b.volume.object.EmptyDir = &newdir
	return b
}

// Build returns the Volume API instance
func (b *Builder) Build() (*corev1.Volume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.volume.object, nil
}
