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

package pvc

import (
	"strings"

	"github.com/openebs/lib-csi/pkg/common/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	client "github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getpvcFn is a typed function that
// abstracts fetching of pvc
type getFn func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error)

// listFn is a typed function that abstracts
// listing of pvcs
type listFn func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error)

// deleteFn is a typed function that abstracts
// deletion of pvcs
type deleteFn func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pvc's collection
type deleteCollectionFn func(cli *kubernetes.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// createFn is a typed function that abstracts
// creation of pvc
type createFn func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)

// updateFn is a typed function that abstracts
// updation of pvc
type updateFn func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)

// Kubeclient enables kubernetes API operations
// on pvc instance
type Kubeclient struct {
	// clientset refers to pvc clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	create              createFn
	update              updateFn
	del                 deleteFn
	delCollection       deleteCollectionFn
}

// KubeclientBuildOption abstracts creating an
// instance of kubeclient
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}

	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}

	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Get(name, opts)
		}
	}

	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).List(opts)
		}
	}

	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Delete(name, deleteOpts)
		}
	}

	if k.delCollection == nil {
		k.delCollection = func(cli *kubernetes.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(deleteOpts, listOpts)
		}
	}

	if k.create == nil {
		k.create = func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
		}
	}

	if k.update == nil {
		k.update = func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Update(pvc)
		}
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// pvc operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = cs
	return k.clientset, nil
}

// Get returns a pvc resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get pvc: missing pvc name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pvc {%s}", name)
	}
	return k.get(cli, name, k.namespace, opts)
}

// List returns a list of pvc
// instances present in kubernetes cluster
func (k *Kubeclient) List(opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pvc listoptions: '%v'", opts)
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pvc instance from the
// kubecrnetes cluster
func (k *Kubeclient) Delete(name string, deleteOpts *metav1.DeleteOptions) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete pvc: missing pvc name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete pvc {%s}", name)
	}
	return k.del(cli, k.namespace, name, deleteOpts)
}

// Create creates a pvc in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	if pvc == nil {
		return nil, errors.New("failed to create pvc: nil pvc object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pvc {%s} in namespace {%s}", pvc.Name, pvc.Namespace)
	}
	return k.create(cli, k.namespace, pvc)
}

// Update updates a pvc in specified namespace in kubernetes cluster
func (k *Kubeclient) Update(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	if pvc == nil {
		return nil, errors.New("failed to update pvc: nil pvc object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update pvc {%s} in namespace {%s}", pvc.Name, pvc.Namespace)
	}
	return k.update(cli, k.namespace, pvc)
}

// CreateCollection creates a list of pvcs
// in specified namespace in kubernetes cluster
func (k *Kubeclient) CreateCollection(
	list *corev1.PersistentVolumeClaimList,
) (*corev1.PersistentVolumeClaimList, error) {
	if list == nil || len(list.Items) == 0 {
		return nil, errors.New("failed to create list of pvcs: nil pvc list provided")
	}

	newlist := &corev1.PersistentVolumeClaimList{}
	for _, item := range list.Items {
		item := item
		obj, err := k.Create(&item)
		if err != nil {
			return nil, err
		}

		newlist.Items = append(newlist.Items, *obj)
	}

	return newlist, nil
}

// DeleteCollection deletes a collection of pvc objects.
func (k *Kubeclient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of pvcs")
	}
	return k.delCollection(cli, k.namespace, listOpts, deleteOpts)
}
