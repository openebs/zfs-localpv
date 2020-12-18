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

package pvc

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
)

// ListBuilder enables building an instance of
// List
type ListBuilder struct {
	// template to build a list of pvcs
	template *corev1.PersistentVolumeClaim

	// count determines the number of
	// pvcs to be built using the provided
	// template
	count int

	list    *List
	filters PredicateList
	errs    []error
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &List{}}
}

// ListBuilderFromTemplate returns a new instance of
// ListBuilder based on the provided pvc template
func ListBuilderFromTemplate(pvc *corev1.PersistentVolumeClaim) *ListBuilder {
	b := NewListBuilder()
	if pvc == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build pvc list: nil pvc template"),
		)
		return b
	}

	b.template = pvc
	b.count = 1
	return b
}

// ListBuilderForAPIObjects returns a new instance of
// ListBuilder based on provided api pvc list
func ListBuilderForAPIObjects(pvcs *corev1.PersistentVolumeClaimList) *ListBuilder {
	b := &ListBuilder{list: &List{}}

	if pvcs == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build pvc list: missing api list"),
		)
		return b
	}

	for _, pvc := range pvcs.Items {
		pvc := pvc
		b.list.items = append(b.list.items, &PVC{object: &pvc})
	}

	return b
}

// ListBuilderForObjects returns a new instance of
// ListBuilder based on provided pvc list
func ListBuilderForObjects(pvcs *List) *ListBuilder {
	b := &ListBuilder{}
	if pvcs == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build pvc list: missing object list"),
		)
		return b
	}

	b.list = pvcs
	return b
}

// WithFilter adds filters on which the pvcs
// are filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}

// WithCount sets the count that determines
// the number of pvcs to be built
func (b *ListBuilder) WithCount(count int) *ListBuilder {
	b.count = count
	return b
}

func (b *ListBuilder) buildFromTemplateIfNilList() {
	if len(b.list.items) != 0 || b.template == nil {
		return
	}

	for i := 0; i < b.count; i++ {
		b.list.items = append(b.list.items, &PVC{object: b.template})
	}
}

// List returns the list of pvc instances
// that was built by this builder
func (b *ListBuilder) List() (*List, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("failed to build pvc list: %+v", b.errs)
	}

	b.buildFromTemplateIfNilList()

	if b.filters == nil || len(b.filters) == 0 {
		return b.list, nil
	}

	filteredList := &List{}
	for _, pvc := range b.list.items {
		if b.filters.all(pvc) {
			filteredList.items = append(filteredList.items, pvc)
		}
	}

	return filteredList, nil
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

// APIList builds core API PVC list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeClaimList, error) {
	l, err := b.List()
	if err != nil {
		return nil, err
	}

	return l.ToAPIList(), nil
}
