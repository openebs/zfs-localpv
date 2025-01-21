package pv

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PV is a wrapper over persistentvolume api
type pvBuildOption func(*PV)

// PV is a wrapper over persistentvolume api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PV struct {
	object *corev1.PersistentVolume
}

// Builder is the builder object for PV
type Builder struct {
	pv   *PV
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pv: &PV{object: &corev1.PersistentVolume{}}}
}

// WithName sets the Name field of PV with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing PV name"))
		return b
	}
	b.pv.object.Name = name
	return b
}

// WithStorageClass sets the StorageClass field of PV with provided arguments
func (b *Builder) WithStorageClass(scName string) *Builder {
	if len(scName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing storageclass name"))
		return b
	}
	b.pv.object.Spec.StorageClassName = scName
	return b
}

// WithCapacity sets the Capacity field in PV with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(b.errs, errors.Wrapf(err, "failed to build PV object: failed to parse capacity {%s}", capacity))
		return b
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	b.pv.object.Spec.Capacity = resourceList
	return b
}

// WithAccessModes sets the AccessMode field in PV with provided arguments
func (b *Builder) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *Builder {
	if len(accessMode) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing accessmodes"))
		return b
	}
	b.pv.object.Spec.AccessModes = accessMode
	return b
}

// Build returns the PV API instance
func (b *Builder) Build() (*corev1.PersistentVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pv.object, nil
}
