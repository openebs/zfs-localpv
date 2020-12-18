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
	"github.com/openebs/lib-csi/pkg/common/errors"
	storagev1 "k8s.io/api/storage/v1"
)

// ListBuilder enables building an instance of StorageClassList
type ListBuilder struct {
	list    *StorageClassList
	filters predicateList
	errs    []error
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &StorageClassList{items: []*StorageClass{}}}
}

// ListBuilderForAPIList builds the ListBuilder object based on SC API list
func ListBuilderForAPIList(scl *storagev1.StorageClassList) *ListBuilder {
	b := &ListBuilder{list: &StorageClassList{}}
	if scl == nil {
		b.errs = append(b.errs, errors.New("failed to build storageclass list: missing api list"))
		return b
	}
	for _, sc := range scl.Items {
		sc := sc
		b.list.items = append(b.list.items, &StorageClass{object: &sc})
	}
	return b
}

// ListBuilderForObjects returns a instance of ListBuilder from SC instances
func ListBuilderForObjects(scl *StorageClassList) *ListBuilder {
	b := &ListBuilder{list: &StorageClassList{}}
	if scl == nil {
		b.errs = append(b.errs, errors.New("failed to build storageclass list: missing object list"))
		return b
	}
	b.list = scl
	return b
}

// List returns the list of StorageClass instances that was built by this builder
func (b *ListBuilder) List() (*StorageClassList, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("failed to list storageclass: %+v", b.errs)
	}
	if b.filters == nil || len(b.filters) == 0 {
		return b.list, nil
	}
	filtered := &StorageClassList{}
	for _, sc := range b.list.items {
		if b.filters.all(sc) {
			sc := sc // Pin it
			filtered.items = append(filtered.items, sc)
		}
	}
	return filtered, nil
}

// WithFilter add filters on which the StorageClass has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*storagev1.StorageClassList, error) {
	l, err := b.List()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}

// Len returns the number of items present
// in the List of a builder
func (b *ListBuilder) Len() (int, error) {
	l, err := b.List()
	if err != nil {
		return 0, err
	}
	return l.Len(), nil
}
