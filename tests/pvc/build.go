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

package pvc

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for PVC
type Builder struct {
	pvc  *PVC
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pvc: &PVC{object: &corev1.PersistentVolumeClaim{}}}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(pvc *corev1.PersistentVolumeClaim) *Builder {
	if pvc == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build pvc object: nil pvc"),
		)
		return b
	}
	return &Builder{
		pvc: &PVC{
			object: pvc,
		},
	}
}

// WithName sets the Name field of PVC with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing PVC name"))
		return b
	}
	b.pvc.object.Name = name
	return b
}

// WithGenerateName sets the GenerateName field of
// PVC with provided value
func (b *Builder) WithGenerateName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build PVC object: missing PVC generateName"),
		)
		return b
	}

	b.pvc.object.GenerateName = name
	return b
}

// WithNamespace sets the Namespace field of PVC provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		namespace = "default"
	}
	b.pvc.object.Namespace = namespace
	return b
}

// WithAnnotations sets the Annotations field of PVC with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing annotations"))
		return b
	}
	b.pvc.object.Annotations = annotations
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build PVC object: missing labels"),
		)
		return b
	}

	if b.pvc.object.Labels == nil {
		b.pvc.object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.pvc.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build PVC object: missing labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.pvc.object.Labels = newlbls
	return b
}

// WithStorageClass sets the StorageClass field of PVC with provided arguments
func (b *Builder) WithStorageClass(scName string) *Builder {
	if len(scName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing storageclass name"))
		return b
	}
	b.pvc.object.Spec.StorageClassName = &scName
	return b
}

// WithAccessModes sets the AccessMode field in PVC with provided arguments
func (b *Builder) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *Builder {
	if len(accessMode) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing accessmodes"))
		return b
	}
	b.pvc.object.Spec.AccessModes = accessMode
	return b
}

// WithCapacity sets the Capacity field in PVC with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(b.errs, errors.Wrapf(err, "failed to build PVC object: failed to parse capacity {%s}", capacity))
		return b
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	b.pvc.object.Spec.Resources.Requests = resourceList
	return b
}

// Build returns the PVC API instance
func (b *Builder) Build() (*corev1.PersistentVolumeClaim, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pvc.object, nil
}

// WithVolumeMode sets the VolumeMode field in PVC with provided arguments
func (b *Builder) WithVolumeMode(volumemode *corev1.PersistentVolumeMode) *Builder {
	if volumemode == nil {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing volumemode"))
		return b
	}
	b.pvc.object.Spec.VolumeMode = volumemode
	return b
}
