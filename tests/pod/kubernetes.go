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
	"bytes"
	"encoding/json"

	"github.com/openebs/lib-csi/pkg/common/errors"
	client "github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (*clientset.Clientset, error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (*clientset.Clientset, error)

// getKubeConfigFn is a typed function that
// abstracts fetching of config
type getKubeConfigFn func() (*rest.Config, error)

// getKubeConfigForPathFn is a typed function that
// abstracts fetching of config from kubeConfigPath
type getKubeConfigForPathFn func(kubeConfigPath string) (*rest.Config, error)

// createFn is a typed function that abstracts
// creation of pod
type createFn func(cli *clientset.Clientset, namespace string, pod *corev1.Pod) (*corev1.Pod, error)

// listFn is a typed function that abstracts
// listing of pods
type listFn func(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)

// deleteFn is a typed function that abstracts
// deleting of pod
type deleteFn func(cli *clientset.Clientset, namespace, name string, opts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pod's collection
type deleteCollectionFn func(cli *clientset.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// getFn is a typed function that abstracts
// to get pod
type getFn func(cli *clientset.Clientset, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error)

// execFn is a typed function that abstracts
// pod exec
type execFn func(cli *clientset.Clientset, config *rest.Config, name, namespace string, opts *corev1.PodExecOptions) (*ExecOutput, error)

// defaultExec is the default implementation of execFn
func defaultExec(
	cli *clientset.Clientset,
	config *rest.Config,
	name string,
	namespace string,
	opts *corev1.PodExecOptions,
) (*ExecOutput, error) {
	var stdout, stderr bytes.Buffer

	req := cli.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(opts, scheme.ParameterCodec)

	// create exec executor which is an interface
	// for transporting shell-style streams
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	// Stream initiates transport of standard shell streams
	// It will transport any non-nil stream to a remote system,
	// and return an error if a problem occurs
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    opts.TTY,
	})
	if err != nil {
		return nil, err
	}

	execOutput := &ExecOutput{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	return execOutput, nil
}

// KubeClient enables kubernetes API operations
// on pod instance
type KubeClient struct {
	// clientset refers to pod clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *clientset.Clientset

	// namespace holds the namespace on which
	// KubeClient has to operate
	namespace string

	// kubeConfig represents kubernetes config
	kubeConfig *rest.Config

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getKubeConfig        getKubeConfigFn
	getKubeConfigForPath getKubeConfigForPathFn
	getClientset         getClientsetFn
	getClientsetForPath  getClientsetForPathFn
	create               createFn
	list                 listFn
	del                  deleteFn
	delCollection        deleteCollectionFn
	get                  getFn
	exec                 execFn
}

// ExecOutput struct contains stdout and stderr
type ExecOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// KubeClientBuildOption defines the abstraction
// to build a KubeClient instance
type KubeClientBuildOption func(*KubeClient)

// withDefaults sets the default options
// of KubeClient instance
func (k *KubeClient) withDefaults() {
	if k.getKubeConfig == nil {
		k.getKubeConfig = func() (config *rest.Config, err error) {
			return client.New().Config()
		}
	}
	if k.getKubeConfigForPath == nil {
		k.getKubeConfigForPath = func(kubeConfigPath string) (
			config *rest.Config, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).
				GetConfigForPathOrDirect()
		}
	}
	if k.getClientset == nil {
		k.getClientset = func() (clients *clientset.Clientset, err error) {
			return client.New().Clientset()
		}
	}
	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (
			clients *clientset.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}
	if k.create == nil {
		k.create = func(cli *clientset.Clientset,
			namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
			return cli.CoreV1().Pods(namespace).Create(pod)
		}
	}
	if k.list == nil {
		k.list = func(cli *clientset.Clientset,
			namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
			return cli.CoreV1().Pods(namespace).List(opts)
		}
	}
	if k.del == nil {
		k.del = func(cli *clientset.Clientset, namespace,
			name string, opts *metav1.DeleteOptions) error {
			return cli.CoreV1().Pods(namespace).Delete(name, opts)
		}
	}
	if k.get == nil {
		k.get = func(cli *clientset.Clientset, namespace,
			name string, opts metav1.GetOptions) (*corev1.Pod, error) {
			return cli.CoreV1().Pods(namespace).Get(name, opts)
		}
	}
	if k.delCollection == nil {
		k.delCollection = func(cli *clientset.Clientset, namespace string,
			listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().Pods(namespace).DeleteCollection(deleteOpts, listOpts)
		}
	}
	if k.exec == nil {
		k.exec = defaultExec
	}
}

// WithClientSet sets the kubernetes client against
// the KubeClient instance
func WithClientSet(c *clientset.Clientset) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeClientBuildOption {
	return func(k *KubeClient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of KubeClient meant for
// zfs volume replica operations
func NewKubeClient(opts ...KubeClientBuildOption) *KubeClient {
	k := &KubeClient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the kubernetes namespace against
// the provided namespace
func (k *KubeClient) WithNamespace(namespace string) *KubeClient {
	k.namespace = namespace
	return k
}

// WithKubeConfig sets the kubernetes config against
// the KubeClient instance
func (k *KubeClient) WithKubeConfig(config *rest.Config) *KubeClient {
	k.kubeConfig = config
	return k
}

func (k *KubeClient) getClientsetForPathOrDirect() (
	*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *KubeClient) getClientsetOrCached() (*clientset.Clientset, error) {
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

func (k *KubeClient) getKubeConfigForPathOrDirect() (*rest.Config, error) {
	if k.kubeConfigPath != "" {
		return k.getKubeConfigForPath(k.kubeConfigPath)
	}
	return k.getKubeConfig()
}

// getKubeConfigOrCached returns either a new instance
// of kubernetes config or its cached copy
func (k *KubeClient) getKubeConfigOrCached() (*rest.Config, error) {
	if k.kubeConfig != nil {
		return k.kubeConfig, nil
	}

	kc, err := k.getKubeConfigForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kube config")
	}
	k.kubeConfig = kc
	return k.kubeConfig, nil
}

// List returns a list of pod
// instances present in kubernetes cluster
func (k *KubeClient) List(opts metav1.ListOptions) (*corev1.PodList, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list pods")
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes a pod instance present in kubernetes cluster
func (k *KubeClient) Delete(name string, opts *metav1.DeleteOptions) error {
	if len(name) == 0 {
		return errors.New("failed to delete pod: missing pod name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to delete pod {%s}: failed to get clientset",
			name,
		)
	}
	return k.del(cli, k.namespace, name, opts)
}

// Create creates a pod in specified namespace in kubernetes cluster
func (k *KubeClient) Create(pod *corev1.Pod) (*corev1.Pod, error) {
	if pod == nil {
		return nil, errors.New("failed to create pod: nil pod object")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to create pod {%s} in namespace {%s}",
			pod.Name,
			pod.Namespace,
		)
	}
	return k.create(cli, k.namespace, pod)
}

// Get gets a pod object present in kubernetes cluster
func (k *KubeClient) Get(name string,
	opts metav1.GetOptions) (*corev1.Pod, error) {
	if len(name) == 0 {
		return nil, errors.New("failed to get pod: missing pod name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get pod {%s}: failed to get clientset",
			name,
		)
	}
	return k.get(cli, k.namespace, name, opts)
}

// GetRaw gets pod object for a given name and namespace present
// in kubernetes cluster and returns result in raw byte.
func (k *KubeClient) GetRaw(name string,
	opts metav1.GetOptions) ([]byte, error) {
	p, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// Exec runs a command remotely in a container of a pod
func (k *KubeClient) Exec(name string,
	opts *corev1.PodExecOptions) (*ExecOutput, error) {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, err
	}
	config, err := k.getKubeConfigOrCached()
	if err != nil {
		return nil, err
	}
	return k.exec(cli, config, name, k.namespace, opts)
}

// ExecRaw runs a command remotely in a container of a pod
// and returns raw output
func (k *KubeClient) ExecRaw(name string,
	opts *corev1.PodExecOptions) ([]byte, error) {
	execOutput, err := k.Exec(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(execOutput)
}

// DeleteCollection deletes a collection of pod objects.
func (k *KubeClient) DeleteCollection(listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error {
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return errors.Wrapf(err, "failed to delete the collection of pods")
	}
	return k.delCollection(cli, k.namespace, listOpts, deleteOpts)
}
