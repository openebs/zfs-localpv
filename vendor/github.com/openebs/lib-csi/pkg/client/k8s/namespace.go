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

package k8s

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceGetter abstracts fetching of Namespace from kubernetes cluster
type NamespaceGetter interface {
	Get(name string, options metav1.GetOptions) (*corev1.Namespace, error)
}

// NamespaceLister abstracts fetching of a list of namespaces from kubernetes cluster
type NamespaceLister interface {
	List(options metav1.ListOptions) (*corev1.NamespaceList, error)
}

// NamespaceStruct is used to initialise kubernetes namespace instnaces
type NamespaceStruct struct{}

// Namespace returns a pointer to the namespace struct
func Namespace() *NamespaceStruct {
	return &NamespaceStruct{}
}

// Get returns a namespace instance from kubernetes cluster
func (ns *NamespaceStruct) Get(name string, options metav1.GetOptions) (*corev1.Namespace, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get namespace: %s", name)
	}
	return cs.CoreV1().Namespaces().Get(name, options)
}

// List returns a slice of namespaces defined in a Kubernetes cluster
func (ns *NamespaceStruct) List(options metav1.ListOptions) (*corev1.NamespaceList, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get namespaces")
	}
	return cs.CoreV1().Namespaces().List(options)
}
