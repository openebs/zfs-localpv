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
	"k8s.io/apimachinery/pkg/version"
)

// GetServerVersion uses the client-go Discovery client to get the
// kubernetes version struct
func GetServerVersion() (*version.Info, error) {
	cs, err := Clientset().Get()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get apiserver version")
	}
	return cs.Discovery().ServerVersion()
}
