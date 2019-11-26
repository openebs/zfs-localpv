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

package sc

import (
	storagev1 "k8s.io/api/storage/v1"
)

// StorageClass is a wrapper over API based
// storage class instance
type StorageClass struct {
	object *storagev1.StorageClass
}

// StorageClassList holds the list of StorageClass instances
type StorageClassList struct {
	items []*StorageClass
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided StorageClass instance
type Predicate func(*StorageClass) bool

// predicateList holds the list of predicates
type predicateList []Predicate

// ToAPIList converts StorageClassList to API StorageClassList
func (scl *StorageClassList) ToAPIList() *storagev1.StorageClassList {
	sclist := &storagev1.StorageClassList{}
	for _, sc := range scl.items {
		sc := sc // Pin it
		sclist.Items = append(sclist.Items, *sc.object)
	}
	return sclist
}

// all returns true if all the predicateList
// succeed against the provided StorageClass
// instance
func (l predicateList) all(sc *StorageClass) bool {
	for _, pred := range l {
		if !pred(sc) {
			return false
		}
	}
	return true
}

// Len returns the number of items present in the StorageClassList
func (scl *StorageClassList) Len() int {
	return len(scl.items)
}

// NewForAPIObject returns a new instance of StorageClass
func NewForAPIObject(obj *storagev1.StorageClass) *StorageClass {
	sc := &StorageClass{object: obj}
	return sc
}
