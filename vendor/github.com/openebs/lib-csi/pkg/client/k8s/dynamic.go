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
	k8sdynamic "k8s.io/client-go/dynamic"
)

// DynamicProvider abstracts providing kubernetes dynamic client interface
type DynamicProvider interface {
	Provide() (k8sdynamic.Interface, error)
}

//DynamicStruct is used to initialise a kuberenets dynamic interface
type DynamicStruct struct{}

// Dynamic returns a new instance of dynamic
func Dynamic() *DynamicStruct {
	return &DynamicStruct{}
}

// Provide provides a kubernetes dynamic client capable of invoking operations
// against kubernetes resources
func (d *DynamicStruct) Provide() (k8sdynamic.Interface, error) {
	config, err := Config().Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to provide dynamic client")
	}
	return k8sdynamic.NewForConfig(config)
}
