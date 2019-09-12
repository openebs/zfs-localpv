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

package builder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
)

// ZFSVolume is a wrapper over
// ZFSVolume API instance
type ZFSVolume struct {
	Object *apis.ZFSVolume
}

// From returns a new instance of
// csi volume
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
// of csi volume instances
type ZFSVolumeList struct {
	List apis.ZFSVolumeList
}

// Len returns the number of items present
// in the ZFSVolumeList
func (p *ZFSVolumeList) Len() int {
	return len(p.List.Items)
}

// all returns true if all the predicates
// succeed against the provided ZFSVolume
// instance
func (l predicateList) all(p *ZFSVolume) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided ZFSVolume instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(p *ZFSVolume) bool {
		for key, value := range keyValuePair {
			if !p.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided ZFSVolume instance
func (p *ZFSVolume) HasLabel(key, value string) bool {
	val, ok := p.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided ZFSVolume instance
func HasLabel(key, value string) Predicate {
	return func(p *ZFSVolume) bool {
		return p.HasLabel(key, value)
	}
}

// IsNil returns true if the csi volume instance
// is nil
func (p *ZFSVolume) IsNil() bool {
	return p.Object == nil
}

// IsNil is predicate to filter out nil csi volume
// instances
func IsNil() Predicate {
	return func(p *ZFSVolume) bool {
		return p.IsNil()
	}
}

// GetAPIObject returns csi volume's API instance
func (p *ZFSVolume) GetAPIObject() *apis.ZFSVolume {
	return p.Object
}
