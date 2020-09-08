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

package bkpbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
)

// ZFSBackup is a wrapper over
// ZFSBackup API instance
type ZFSBackup struct {
	// ZFSSnap object
	Object *apis.ZFSBackup
}

// From returns a new instance of
// zfsbkp bkpume
func From(bkp *apis.ZFSBackup) *ZFSBackup {
	return &ZFSBackup{
		Object: bkp,
	}
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*ZFSBackup) bool

// PredicateList holds a list of predicate
type predicateList []Predicate

// ZFSBackupList holds the list
// of zfs backup instances
type ZFSBackupList struct {
	// List contains list of backups
	List apis.ZFSBackupList
}

// Len returns the number of items present
// in the ZFSBackupList
func (bkpList *ZFSBackupList) Len() int {
	return len(bkpList.List.Items)
}

// all returns true if all the predicates
// succeed against the provided ZFSBackup
// instance
func (l predicateList) all(bkp *ZFSBackup) bool {
	for _, pred := range l {
		if !pred(bkp) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided ZFSBackup instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(bkp *ZFSBackup) bool {
		for key, value := range keyValuePair {
			if !bkp.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided ZFSBackup instance
func (bkp *ZFSBackup) HasLabel(key, value string) bool {
	val, ok := bkp.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided ZFSBackup instance
func HasLabel(key, value string) Predicate {
	return func(bkp *ZFSBackup) bool {
		return bkp.HasLabel(key, value)
	}
}

// IsNil returns true if the zfsbkp bkpume instance
// is nil
func (bkp *ZFSBackup) IsNil() bool {
	return bkp.Object == nil
}

// IsNil is predicate to filter out nil zfsbkp bkpume
// instances
func IsNil() Predicate {
	return func(bkp *ZFSBackup) bool {
		return bkp.IsNil()
	}
}

// GetAPIObject returns zfsbkp bkpume's API instance
func (bkp *ZFSBackup) GetAPIObject() *apis.ZFSBackup {
	return bkp.Object
}
