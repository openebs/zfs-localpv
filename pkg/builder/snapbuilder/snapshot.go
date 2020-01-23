// Copyright © 2020 The OpenEBS Authors
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

package snapbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
)

// ZFSSnapshot is a wrapper over
// ZFSSnapshot API instance
type ZFSSnapshot struct {
	// ZFSSnap object
	Object *apis.ZFSSnapshot
}

// From returns a new instance of
// zfssnap volume
func From(snap *apis.ZFSSnapshot) *ZFSSnapshot {
	return &ZFSSnapshot{
		Object: snap,
	}
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*ZFSSnapshot) bool

// PredicateList holds a list of predicate
type predicateList []Predicate

// ZFSSnapshotList holds the list
// of zfs snapshot instances
type ZFSSnapshotList struct {
	// List contains list of snapshots
	List apis.ZFSSnapshotList
}

// Len returns the number of items present
// in the ZFSSnapshotList
func (snapList *ZFSSnapshotList) Len() int {
	return len(snapList.List.Items)
}

// all returns true if all the predicates
// succeed against the provided ZFSSnapshot
// instance
func (l predicateList) all(snap *ZFSSnapshot) bool {
	for _, pred := range l {
		if !pred(snap) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided ZFSSnapshot instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(snap *ZFSSnapshot) bool {
		for key, value := range keyValuePair {
			if !snap.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided ZFSSnapshot instance
func (snap *ZFSSnapshot) HasLabel(key, value string) bool {
	val, ok := snap.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided ZFSSnapshot instance
func HasLabel(key, value string) Predicate {
	return func(snap *ZFSSnapshot) bool {
		return snap.HasLabel(key, value)
	}
}

// IsNil returns true if the zfssnap volume instance
// is nil
func (snap *ZFSSnapshot) IsNil() bool {
	return snap.Object == nil
}

// IsNil is predicate to filter out nil zfssnap volume
// instances
func IsNil() Predicate {
	return func(snap *ZFSSnapshot) bool {
		return snap.IsNil()
	}
}

// GetAPIObject returns zfssnap volume's API instance
func (snap *ZFSSnapshot) GetAPIObject() *apis.ZFSSnapshot {
	return snap.Object
}
