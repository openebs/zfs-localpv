// Copyright © 2021 The OpenEBS Authors
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

package nodebuilder

import (
	"context"
	"encoding/json"

	client "github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (
	clientset *clientset.Clientset,
	err error,
)

// createFn is a typed function that abstracts
// creating zfs node instance
type createFn func(
	cs *clientset.Clientset,
	upgradeResultObj *apis.ZFSNode,
	namespace string,
) (*apis.ZFSNode, error)

// getFn is a typed function that abstracts
// fetching a zfs node instance
type getFn func(
	cli *clientset.Clientset,
	name,
	namespace string,
	opts metav1.GetOptions,
) (*apis.ZFSNode, error)

// listFn is a typed function that abstracts
// listing of zfs node instances
type listFn func(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.ZFSNodeList, error)

// delFn is a typed function that abstracts
// deleting a zfs node instance
type delFn func(
	cli *clientset.Clientset,
	name,
	namespace string,
	opts *metav1.DeleteOptions,
) error

// updateFn is a typed function that abstracts
// updating zfs node instance
type updateFn func(
	cs *clientset.Clientset,
	node *apis.ZFSNode,
	namespace string,
) (*apis.ZFSNode, error)

// Kubeclient enables kubernetes API operations
// on zfs node instance
type Kubeclient struct {
	// clientset refers to zfs node's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	kubeConfigPath string

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	del                 delFn
	create              createFn
	update              updateFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// defaultGetClientset is the default implementation to
// get kubernetes clientset instance
func defaultGetClientset() (clients *clientset.Clientset, err error) {

	config, err := client.GetConfig(client.New())
	if err != nil {
		return nil, err
	}

	return clientset.NewForConfig(config)

}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *clientset.Clientset, err error) {
	config, err := client.GetConfig(
		client.New(client.WithKubeConfigPath(kubeConfigPath)))
	if err != nil {
		return nil, err
	}

	return clientset.NewForConfig(config)
}

// defaultGet is the default implementation to get
// a zfs node instance in kubernetes cluster
func defaultGet(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apis.ZFSNode, error) {
	return cli.ZfsV1().
		ZFSNodes(namespace).
		Get(context.TODO(), name, opts)
}

// defaultList is the default implementation to list
// zfs node instances in kubernetes cluster
func defaultList(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apis.ZFSNodeList, error) {
	return cli.ZfsV1().
		ZFSNodes(namespace).
		List(context.TODO(), opts)
}

// defaultCreate is the default implementation to delete
// a zfs node instance in kubernetes cluster
func defaultDel(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	deletePropagation := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &deletePropagation
	err := cli.ZfsV1().
		ZFSNodes(namespace).
		Delete(context.TODO(), name, *opts)
	return err
}

// defaultCreate is the default implementation to create
// a zfs node instance in kubernetes cluster
func defaultCreate(
	cli *clientset.Clientset,
	node *apis.ZFSNode,
	namespace string,
) (*apis.ZFSNode, error) {
	return cli.ZfsV1().
		ZFSNodes(namespace).
		Create(context.TODO(), node, metav1.CreateOptions{})
}

// defaultUpdate is the default implementation to update
// a zfs node instance in kubernetes cluster
func defaultUpdate(
	cli *clientset.Clientset,
	node *apis.ZFSNode,
	namespace string,
) (*apis.ZFSNode, error) {
	return cli.ZfsV1().
		ZFSNodes(namespace).
		Update(context.TODO(), node, metav1.UpdateOptions{})
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = defaultGetClientsetForPath
	}
	if k.get == nil {
		k.get = defaultGet
	}
	if k.list == nil {
		k.list = defaultList
	}
	if k.del == nil {
		k.del = defaultDel
	}
	if k.create == nil {
		k.create = defaultCreate
	}
	if k.update == nil {
		k.update = defaultUpdate
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithNamespace sets the provided namespace
// against this Kubeclient instance
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubernetes client
// against the provided path
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeclient returns a new instance of
// kubeclient meant for zfs node operations
func NewKubeclient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}

	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (
	*clientset.Clientset,
	error,
) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}

	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil,
			errors.Wrapf(
				err,
				"failed to get clientset",
			)
	}

	k.clientset = c
	return k.clientset, nil
}

// Create creates a zfs node instance
// in kubernetes cluster
func (k *Kubeclient) Create(node *apis.ZFSNode) (*apis.ZFSNode, error) {
	if node == nil {
		return nil,
			errors.New(
				"failed to create zfs node: nil node object",
			)
	}
	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create zfs node {%s} in namespace {%s}",
			node.Name,
			k.namespace,
		)
	}

	return k.create(cs, node, k.namespace)
}

// Get returns zfs node object for given name
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apis.ZFSNode, error) {
	if name == "" {
		return nil,
			errors.New(
				"failed to get zfs node: missing zfs node name",
			)
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get zfs node {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}

	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns zfs node instance
// in bytes
func (k *Kubeclient) GetRaw(
	name string,
	opts metav1.GetOptions,
) ([]byte, error) {
	if name == "" {
		return nil, errors.New(
			"failed to get raw zfs node: missing node name",
		)
	}
	csiv, err := k.Get(name, opts)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get zfs node {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}

	return json.Marshal(csiv)
}

// List returns a list of zfs node
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*apis.ZFSNodeList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to list zfs nodes in namespace {%s}",
			k.namespace,
		)
	}

	return k.list(cli, k.namespace, opts)
}

// Delete deletes the zfs node from
// kubernetes
func (k *Kubeclient) Delete(name string) error {
	if name == "" {
		return errors.New(
			"failed to delete zfsnode: missing node name",
		)
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete zfsnode {%s} in namespace {%s}",
			name,
			k.namespace,
		)
	}

	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}

// Update updates this zfs node instance
// against kubernetes cluster
func (k *Kubeclient) Update(node *apis.ZFSNode) (*apis.ZFSNode, error) {
	if node == nil {
		return nil,
			errors.New(
				"failed to update zfsnode: nil node object",
			)
	}

	cs, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to update zfsnode {%s} in namespace {%s}",
			node.Name,
			node.Namespace,
		)
	}

	return k.update(cs, node, k.namespace)
}
