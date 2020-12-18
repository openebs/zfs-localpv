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
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapGetter abstracts fetching of ConfigMap instance from kubernetes
// cluster
type ConfigMapGetter interface {
	Get(options metav1.GetOptions) (*corev1.ConfigMap, error)
}

// Configmap is used to initialise a kubernetes Configmap struct
type Configmap struct {
	namespace string // namespace where this configmap exists
	name      string // name of this configmap
}

// ConfigMap returns a new instance of configmap
func ConfigMap(namespace, name string) *Configmap {
	return &Configmap{namespace: namespace, name: name}
}

// Get returns configmap instance from kubernetes cluster
func (c *Configmap) Get(options metav1.GetOptions) (cm *corev1.ConfigMap, err error) {
	if len(strings.TrimSpace(c.name)) == 0 {
		return nil, errors.Errorf("missing config map name: failed to get config map from namespace %s", c.namespace)
	}
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config map %s %s", c.namespace, c.name)
	}
	return cs.CoreV1().ConfigMaps(c.namespace).Get(c.name, options)
}
