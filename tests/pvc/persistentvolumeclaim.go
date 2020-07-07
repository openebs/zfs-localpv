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

package pvc

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// PVC is a wrapper over persistentvolumeclaim api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PVC struct {
	object *corev1.PersistentVolumeClaim
}

// List is a wrapper over persistentvolumeclaim api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type List struct {
	items []*PVC
}

// Len returns the number of items present
// in the List
func (p *List) Len() int {
	return len(p.items)
}

// ToAPIList converts List to API List
func (p *List) ToAPIList() *corev1.PersistentVolumeClaimList {
	plist := &corev1.PersistentVolumeClaimList{}
	for _, pvc := range p.items {
		plist.Items = append(plist.Items, *pvc.object)
	}
	return plist
}

type pvcBuildOption func(*PVC)

// NewForAPIObject returns a new instance of PVC
func NewForAPIObject(obj *corev1.PersistentVolumeClaim, opts ...pvcBuildOption) *PVC {
	p := &PVC{object: obj}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pvc instance
type Predicate func(*PVC) bool

// IsBound returns true if the pvc is bounded
func (p *PVC) IsBound() bool {
	return p.object.Status.Phase == corev1.ClaimBound
}

// IsBound is a predicate to filter out pvcs
// which is bounded
func IsBound() Predicate {
	return func(p *PVC) bool {
		return p.IsBound()
	}
}

// IsNil returns true if the PVC instance
// is nil
func (p *PVC) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil PVC
// instances
func IsNil() Predicate {
	return func(p *PVC) bool {
		return p.IsNil()
	}
}

// ContainsName is filter function to filter pvc's
// based on the name
func ContainsName(name string) Predicate {
	return func(p *PVC) bool {
		return strings.Contains(p.object.GetName(), name)
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided pvc
// instance
func (l PredicateList) all(p *PVC) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}
