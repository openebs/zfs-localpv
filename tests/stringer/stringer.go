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

package stringer

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
)

// Yaml returns the provided object
// as a yaml formatted string
func Yaml(ctx string, obj interface{}) string {
	if obj == nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	str, ok := obj.(string)
	if ok {
		return fmt.Sprintf("\n%s {%s}", ctx, str)
	}

	b, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	return fmt.Sprintf("\n%s {%s}", ctx, string(b))
}

// JSONIndent returns the provided object
// as a json indent string
func JSONIndent(ctx string, obj interface{}) string {
	if obj == nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	str, ok := obj.(string)
	if ok {
		return fmt.Sprintf("\n%s {%s}", ctx, str)
	}

	b, err := json.MarshalIndent(obj, "", ".")
	if err != nil {
		return fmt.Sprintf("\n%s {nil}", ctx)
	}

	return fmt.Sprintf("\n%s %s", ctx, string(b))
}
