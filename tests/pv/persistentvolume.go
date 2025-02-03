package pv

import (
	corev1 "k8s.io/api/core/v1"
)

// IsAvailable returns true if the pv is bounded
func (p *PV) IsAvailable() bool {
	return p.object.Status.Phase == corev1.VolumeAvailable
}

// NewForAPIObject returns a new instance of PV
func NewForAPIObject(obj *corev1.PersistentVolume, opts ...pvBuildOption) *PV {
	p := &PV{object: obj}
	for _, o := range opts {
		o(p)
	}
	return p
}
