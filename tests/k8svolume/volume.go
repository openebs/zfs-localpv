// Copyright 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8svolume

import (
	corev1 "k8s.io/api/core/v1"
)

// Volume is a wrapper over named volume api object, used
// within Pods. It provides build, validations and other common
// logic to be used by various feature specific callers.
type Volume struct {
	object *corev1.Volume
}

type volumeBuildOption func(*Volume)

// NewForAPIObject returns a new instance of Volume
func NewForAPIObject(obj *corev1.Volume, opts ...volumeBuildOption) *Volume {
	v := &Volume{object: obj}
	for _, o := range opts {
		o(v)
	}
	return v
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided volume instance
type Predicate func(*Volume) bool

// IsNil returns true if the Volume instance
// is nil
func (v *Volume) IsNil() bool {
	return v.object == nil
}

// IsNil is predicate to filter out nil Volume
// instances
func IsNil() Predicate {
	return func(v *Volume) bool {
		return v.IsNil()
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided pvc
// instance
func (l PredicateList) all(v *Volume) bool {
	for _, pred := range l {
		if !pred(v) {
			return false
		}
	}
	return true
}
