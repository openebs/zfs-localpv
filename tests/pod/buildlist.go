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

package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// ListBuilder enables building an instance of
// List
type ListBuilder struct {
	list    *List
	filters predicateList
}

// NewListBuilder returns a instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &List{items: []*Pod{}}}
}

// ListBuilderForAPIList returns a instance of ListBuilder from API List
func ListBuilderForAPIList(pods *corev1.PodList) *ListBuilder {
	b := &ListBuilder{list: &List{}}
	if pods == nil {
		return b
	}
	for _, p := range pods.Items {
		p := p
		b.list.items = append(b.list.items, &Pod{object: &p})
	}
	return b
}

// ListBuilderForObjectList returns a instance of ListBuilder from API Pods
func ListBuilderForObjectList(pods ...*Pod) *ListBuilder {
	b := &ListBuilder{list: &List{}}
	if pods == nil {
		return b
	}
	for _, p := range pods {
		p := p
		b.list.items = append(b.list.items, p)
	}
	return b
}

// List returns the list of pod
// instances that was built by this
// builder
func (b *ListBuilder) List() *List {
	if b.filters == nil || len(b.filters) == 0 {
		return b.list
	}
	filtered := &List{}
	for _, pod := range b.list.items {
		if b.filters.all(pod) {
			filtered.items = append(filtered.items, pod)
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
