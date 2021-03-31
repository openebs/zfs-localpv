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
	"context"

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

//NodeStruct returns a struct used to instantiate a kubernetes Node
type NodeStruct struct{}

// Node returnd a pointer to the node struct
func Node() *NodeStruct {
	return &NodeStruct{}
}

// Get returns a node instance from kubernetes cluster
func (n *NodeStruct) Get(name string, options metav1.GetOptions) (*corev1.Node, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node: %s", name)
	}
	return cs.CoreV1().Nodes().Get(context.TODO(), name, options)
}

// List returns a slice of Nodes registered in a Kubernetes cluster
func (n *NodeStruct) List(options metav1.ListOptions) (*corev1.NodeList, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get nodes")
	}
	return cs.CoreV1().Nodes().List(context.TODO(), options)
}

// NumberOfNodes returns the number of nodes registered in a Kubernetes cluster
func NumberOfNodes() (int, error) {
	n := Node()
	nodes, err := n.List(metav1.ListOptions{})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get the number of nodes")
	}
	return len(nodes.Items), nil
}

// GetNode returns a node instance from kubernetes cluster
func GetNode(name string) (*corev1.Node, error) {
	n := Node()
	node, err := n.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node")
	}
	return node, nil
}

// ListNodes returns list of node instance from kubernetes cluster
func ListNodes(options metav1.ListOptions) (*corev1.NodeList, error) {
	n := Node()
	nodelist, err := n.List(options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list node")
	}
	return nodelist, nil
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
