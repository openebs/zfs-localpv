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
	"k8s.io/client-go/kubernetes"
)

// ClientsetGetter abstracts fetching of kubernetes clientset
type ClientsetGetter interface {
	Get() (*kubernetes.Clientset, error)
}

// ClientsetStruct is used to export a kuberneter Clientset
type ClientsetStruct struct{}

// Clientset returns a pointer to clientset struct
func Clientset() *ClientsetStruct {
	return &ClientsetStruct{}
}

// Get returns a new instance of kubernetes clientset
func (c *ClientsetStruct) Get() (*kubernetes.Clientset, error) {
	config, err := Config().Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kubernetes clientset")
	}
	return kubernetes.NewForConfig(config)
}
