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

package deploy

import (
	"encoding/json"
	"strings"

	client "github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function
// that abstracts fetching of kubernetes clientset
type getClientsetFn func() (*kubernetes.Clientset, error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeconfig path
type getClientsetForPathFn func(path string) (*kubernetes.Clientset, error)

// getFn is a typed function that abstracts fetching a
// deployment instance from kubernetes cluster
type getFn func(
	cli *kubernetes.Clientset,
	name string,
	namespace string,
	opts *metav1.GetOptions,
) (*appsv1.Deployment, error)

// listFn is a typed function that abstracts listing
// deployment instances from kubernetes cluster
type listFn func(
	cli *kubernetes.Clientset,
	namespace string,
	opts *metav1.ListOptions,
) (*appsv1.DeploymentList, error)

// createFn is a typed function that abstracts
// creating a deployment instance in kubernetes cluster
type createFn func(
	cli *kubernetes.Clientset,
	namespace string,
	deploy *appsv1.Deployment,
) (*appsv1.Deployment, error)

// deleteFn is a typed function that abstracts
// deleting a deployment from kubernetes cluster
type deleteFn func(
	cli *kubernetes.Clientset,
	namespace string,
	name string,
	opts *metav1.DeleteOptions,
) error

// patchFn is a typed function that abstracts
// patching a deployment from kubernetes cluster
type patchFn func(
	cli *kubernetes.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*appsv1.Deployment, error)

// rolloutStatusFn is a typed function that abstracts
// fetching rollout status of a deployment instance from
// kubernetes cluster
type rolloutStatusFn func(d *appsv1.Deployment) (*RolloutOutput, error)

// rolloutStatusfFn is a typed function that abstracts
// fetching rollout status of a deployment instance from
// kubernetes cluster
type rolloutStatusfFn func(d *appsv1.Deployment) ([]byte, error)

// defaultGetClientset is the default implementation to
// get kubernetes clientset instance
func defaultGetClientset() (*kubernetes.Clientset, error) {
	return client.Instance().Clientset()
}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(path string) (*kubernetes.Clientset, error) {
	return client.New(client.WithKubeConfigPath(path)).Clientset()
}

// defaultGet is the default implementation to get a
// deployment instance from kubernetes cluster
func defaultGet(
	cli *kubernetes.Clientset,
	name string,
	namespace string,
	opts *metav1.GetOptions,
) (*appsv1.Deployment, error) {

	return cli.AppsV1().Deployments(namespace).Get(name, *opts)
}

// defaultList is the default implementation to list
// deployment instances from kubernetes cluster
func defaultList(
	cli *kubernetes.Clientset,
	namespace string,
	opts *metav1.ListOptions,
) (*appsv1.DeploymentList, error) {

	return cli.AppsV1().Deployments(namespace).List(*opts)
}

// defaultCreate is the default implementation to create
// a deployment instance in kubernetes cluster
func defaultCreate(
	cli *kubernetes.Clientset,
	namespace string,
	deploy *appsv1.Deployment,
) (*appsv1.Deployment, error) {

	return cli.AppsV1().Deployments(namespace).Create(deploy)
}

// defaultDel is the default implementation to delete a
// deployment instance in kubernetes cluster
func defaultDel(
	cli *kubernetes.Clientset,
	namespace string,
	name string,
	opts *metav1.DeleteOptions,
) error {

	return cli.AppsV1().Deployments(namespace).Delete(name, opts)
}

func defaultPatch(
	cli *kubernetes.Clientset,
	name, namespace string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*appsv1.Deployment, error) {
	return cli.AppsV1().Deployments(namespace).Patch(name, pt, data, subresources...)
}

// defaultRolloutStatus is the default implementation to
// fetch rollout status of a deployment instance from kubernetes
// cluster
func defaultRolloutStatus(d *appsv1.Deployment) (*RolloutOutput, error) {
	b := NewForAPIObject(d)
	return b.RolloutStatus()
}

// defaultRolloutStatusf is the default implementation to fetch
// rollout status of a deployment instance from kubernetes cluster
func defaultRolloutStatusf(d *appsv1.Deployment) ([]byte, error) {
	b := NewForAPIObject(d)
	return b.RolloutStatusRaw()
}

// Kubeclient enables kubernetes API operations on deployment instance
type Kubeclient struct {
	// clientset refers to kubernetes clientset
	//
	// It enables CRUD operations of a deployment instance
	// against a kubernetes cluster
	clientset *kubernetes.Clientset

	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	create              createFn
	del                 deleteFn
	patch               patchFn
	rolloutStatus       rolloutStatusFn
	rolloutStatusf      rolloutStatusfFn
}

// KubeclientBuildOption defines the abstraction to build a
// kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets kubeclient instance's fields with defaults
// if these fields are not set
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
	if k.create == nil {
		k.create = defaultCreate
	}
	if k.del == nil {
		k.del = defaultDel
	}
	if k.patch == nil {
		k.patch = defaultPatch
	}
	if k.rolloutStatus == nil {
		k.rolloutStatus = defaultRolloutStatus
	}
	if k.rolloutStatusf == nil {
		k.rolloutStatusf = defaultRolloutStatusf
	}
}

// WithClientset sets the kubernetes client against the kubeclient instance
func WithClientset(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// WithNamespace set namespace in kubeclient object
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// NewKubeClient returns a new instance of kubeclient meant for deployment.
// caller can configure it with different kubeclientBuildOption
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}

	k.withDefaults()
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}

	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}

	k.clientset = c
	return k.clientset, nil
}

// Get returns deployment object for given name
func (k *Kubeclient) Get(name string) (*appsv1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.get(cli, name, k.namespace, &metav1.GetOptions{})
}

// List returns deployment object for given name
func (k *Kubeclient) List(opts *metav1.ListOptions) (*appsv1.DeploymentList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Patch patches deployment object for given name
func (k *Kubeclient) Patch(
	name string,
	pt types.PatchType,
	data []byte,
	subresources ...string,
) (*appsv1.Deployment, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	return k.patch(cli, name, k.namespace, pt, data, subresources...)
}

// GetRaw returns deployment object for given name
func (k *Kubeclient) GetRaw(name string) ([]byte, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return json.Marshal(d)
}

// Delete deletes a deployment instance from the
// kubernetes cluster
func (k *Kubeclient) Delete(name string, opts *metav1.DeleteOptions) error {

	if strings.TrimSpace(name) == "" {
		return errors.New("failed to delete deployment: missing deployment name")
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete deployment {%s}", name)
	}

	return k.del(cli, k.namespace, name, opts)
}

// Create creates a deployment in specified namespace in kubernetes cluster
func (k *Kubeclient) Create(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {

	if deployment == nil {
		return nil, errors.New("failed to create deployment: nil deployment object")
	}

	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create deployment {%s} in namespace {%s}",
			deployment.Name,
			deployment.Namespace,
		)
	}

	return k.create(cli, k.namespace, deployment)
}

// RolloutStatusf returns deployment's rollout status for given name
// in raw bytes
func (k *Kubeclient) RolloutStatusf(name string) (op []byte, err error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return k.rolloutStatusf(d)
}

// RolloutStatus returns deployment's rollout status for given name
func (k *Kubeclient) RolloutStatus(name string) (*RolloutOutput, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}

	d, err := k.get(cli, name, k.namespace, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return k.rolloutStatus(d)
}
