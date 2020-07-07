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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// test if configmap implements ConfigMapGetter interface
var _ ConfigMapGetter = &Configmap{}

func TestConfigMapGet(t *testing.T) {
	tests := map[string]struct {
		namespace string
		name      string
		options   metav1.GetOptions
		iserr     bool
	}{
		"101": {"", "", metav1.GetOptions{}, true},
		"102": {"default", "", metav1.GetOptions{}, true},
		"103": {"default", "myconf", metav1.GetOptions{}, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ConfigMap(mock.namespace, mock.name).Get(mock.options)
			if !mock.iserr && err != nil {
				t.Fatalf("Test '%s' failed: expected no error: actual '%s'", name, err)
			}
		})
	}
}
