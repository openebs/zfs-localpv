// Copyright © 2019 The OpenEBS Authors
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

package volbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
)

// ZFSVolume is a wrapper over
// ZFSVolume API instance
type ZFSVolume struct {
	// ZFSVolume object
	Object *apis.ZFSVolume
}

// From returns a new instance of
// zfs volume
func From(vol *apis.ZFSVolume) *ZFSVolume {
	return &ZFSVolume{
		Object: vol,
	}
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*ZFSVolume) bool

// PredicateList holds a list of predicate
type predicateList []Predicate

// ZFSVolumeList holds the list
// of zfs volume instances
type ZFSVolumeList struct {
	// List conatils list of volumes
	List apis.ZFSVolumeList
}

// Len returns the number of items present
// in the ZFSVolumeList
func (volList *ZFSVolumeList) Len() int {
	return len(volList.List.Items)
}

// all returns true if all the predicates
// succeed against the provided ZFSVolume
// instance
func (l predicateList) all(vol *ZFSVolume) bool {
	for _, pred := range l {
		if !pred(vol) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided ZFSVolume instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(vol *ZFSVolume) bool {
		for key, value := range keyValuePair {
			if !vol.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided ZFSVolume instance
func (vol *ZFSVolume) HasLabel(key, value string) bool {
	val, ok := vol.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided ZFSVolume instance
func HasLabel(key, value string) Predicate {
	return func(vol *ZFSVolume) bool {
		return vol.HasLabel(key, value)
	}
}

// IsNil returns true if the zfs volume instance
// is nil
func (vol *ZFSVolume) IsNil() bool {
	return vol.Object == nil
}

// IsNil is predicate to filter out nil zfs volume
// instances
func IsNil() Predicate {
	return func(vol *ZFSVolume) bool {
		return vol.IsNil()
	}
}

// GetAPIObject returns zfs volume's API instance
func (vol *ZFSVolume) GetAPIObject() *apis.ZFSVolume {
	return vol.Object
}
