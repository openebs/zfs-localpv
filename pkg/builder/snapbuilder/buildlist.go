/*
Copyright 2020 The OpenEBS Authors

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

package snapbuilder

import (
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
)

// ListBuilder enables building an instance of
// ZFSSnapshotList
type ListBuilder struct {
	list    *apis.ZFSSnapshotList
	filters predicateList
}

// NewListBuilder returns a new instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{
		list: &apis.ZFSSnapshotList{},
	}
}

// ListBuilderFrom returns a new instance of
// ListBuilder from API list instance
func ListBuilderFrom(vols apis.ZFSSnapshotList) *ListBuilder {
	b := &ListBuilder{list: &apis.ZFSSnapshotList{}}
	if len(vols.Items) == 0 {
		return b
	}

	b.list.Items = append(b.list.Items, vols.Items...)
	return b
}

// List returns the list of pod
// instances that was built by this
// builder
func (b *ListBuilder) List() *apis.ZFSSnapshotList {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}

	filtered := &apis.ZFSSnapshotList{}
	for _, vol := range b.list.Items {
		vol := vol // pin it
		if b.filters.all(From(&vol)) {
			filtered.Items = append(filtered.Items, vol)
		}
	}
	return filtered
}

// WithFilter add filters on which the pod
// has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
