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

package v1alpha1

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeGetter abstracts fetching of Node details from kubernetes cluster
type NodeGetter interface {
	Get(name string, options metav1.GetOptions) (*corev1.Node, error)
}

// NodeLister abstracts fetching of Nodes from kubernetes cluster
type NodeLister interface {
	List(options metav1.ListOptions) (*corev1.NodeList, error)
}
type node struct{}

func Node() *node {
	return &node{}
}

// Get returns a node instance from kubernetes cluster
func (n *node) Get(name string, options metav1.GetOptions) (*corev1.Node, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node: %s", name)
	} else {
		return cs.CoreV1().Nodes().Get(name, options)
	}
}

// List returns a slice of Nodes registered in a Kubernetes cluster
func (n *node) List(options metav1.ListOptions) (*corev1.NodeList, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get nodes")
	} else {
		return cs.CoreV1().Nodes().List(options)
	}
}

// NumberOfNodes returns the number of nodes registered in a Kubernetes cluster
func NumberOfNodes() (int, error) {
	n := Node()
	nodes, err := n.List(metav1.ListOptions{})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get the number of nodes")
	} else {
		return len(nodes.Items), nil
	}
}

// GetOSAndKernelVersion gets us the OS,Kernel version
func GetOSAndKernelVersion() (string, error) {
	nodes := Node()
	// get a single node
	firstNode, err := nodes.List(metav1.ListOptions{Limit: 1})
	if err != nil {
		return "unknown, unknown", errors.Wrapf(err, "failed to get the os kernel/arch")
	}
	nodedetails := firstNode.Items[0].Status.NodeInfo
	return nodedetails.OSImage + ", " + nodedetails.KernelVersion, nil
}
