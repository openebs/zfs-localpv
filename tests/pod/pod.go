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

package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// Pod holds the api's pod objects
type Pod struct {
	object *corev1.Pod
}

// List holds the list of API pod instances
type List struct {
	items []*Pod
}

// PredicateList holds a list of predicate
type predicateList []Predicate

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*Pod) bool

// ToAPIList converts List to API List
func (pl *List) ToAPIList() *corev1.PodList {
	plist := &corev1.PodList{}
	for _, pod := range pl.items {
		plist.Items = append(plist.Items, *pod.object)
	}
	return plist
}

type podBuildOption func(*Pod)

// NewForAPIObject returns a new instance of Pod
func NewForAPIObject(obj *corev1.Pod, opts ...podBuildOption) *Pod {
	p := &Pod{object: obj}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Len returns the number of items present in the List
func (pl *List) Len() int {
	return len(pl.items)
}

// all returns true if all the predicates
// succeed against the provided pod
// instance
func (l predicateList) all(p *Pod) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// IsRunning returns true if the pod is in running
// state
func (p *Pod) IsRunning() bool {
	return p.object.Status.Phase == "Running"
}

// IsRunning is a predicate to filter out pods
// which in running state
func IsRunning() Predicate {
	return func(p *Pod) bool {
		return p.IsRunning()
	}
}

// IsCompleted returns true if the pod is in completed
// state
func (p *Pod) IsCompleted() bool {
	return p.object.Status.Phase == "Succeeded"
}

// IsCompleted is a predicate to filter out pods
// which in completed state
func IsCompleted() Predicate {
	return func(p *Pod) bool {
		return p.IsCompleted()
	}
}

// HasLabels returns true if provided labels
// map[key]value are present in the provided List
// instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(p *Pod) bool {
		//		objKeyValues := p.object.GetLabels()
		for key, value := range keyValuePair {
			if !p.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel return true if provided lable
// key and value are present in the the provided List
// instance
func (p *Pod) HasLabel(key, value string) bool {
	val, ok := p.object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel is predicate to filter out labeled
// pod instances
func HasLabel(key, value string) Predicate {
	return func(p *Pod) bool {
		return p.HasLabel(key, value)
	}
}

// IsNil returns true if the pod instance
// is nil
func (p *Pod) IsNil() bool {
	return p.object == nil
}

// IsNil is predicate to filter out nil pod
// instances
func IsNil() Predicate {
	return func(p *Pod) bool {
		return p.IsNil()
	}
}

// GetAPIObject returns a API's Pod
func (p *Pod) GetAPIObject() *corev1.Pod {
	return p.object
}

// FromList created a List with provided api List
func FromList(pods *corev1.PodList) *List {
	pl := ListBuilderForAPIList(pods).
		List()
	return pl
}

// GetScheduledNodes returns the nodes on which pods are scheduled
func (pl *List) GetScheduledNodes() map[string]int {
	nodeNames := make(map[string]int)
	for _, p := range pl.items {
		p := p // pin it
		nodeNames[p.object.Spec.NodeName]++
	}
	return nodeNames
}

// IsMatchNodeAny checks the List is running on the provided nodes
func (pl *List) IsMatchNodeAny(nodes map[string]int) bool {
	for _, p := range pl.items {
		p := p // pin it
		if nodes[p.object.Spec.NodeName] == 0 {
			return false
		}
	}
	return true
}
