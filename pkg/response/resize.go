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
	"github.com/container-storage-interface/spec/lib/go/csi"
)

// ControllerExpandVolumeResponseBuilder helps building an
// instance of csi ControllerExpandVolumeResponse
type ControllerExpandVolumeResponseBuilder struct {
	response *csi.ControllerExpandVolumeResponse
}

// NewControllerExpandVolumeResponseBuilder returns a new
// instance of ControllerExpandVolumeResponse
func NewControllerExpandVolumeResponseBuilder() *ControllerExpandVolumeResponseBuilder {
	return &ControllerExpandVolumeResponseBuilder{
		response: &csi.ControllerExpandVolumeResponse{},
	}
}

// WithCapacityBytes sets the CapacityBytes against the
// ControllerExpandVolumeResponse instance
func (b *ControllerExpandVolumeResponseBuilder) WithCapacityBytes(
	capacity int64) *ControllerExpandVolumeResponseBuilder {
	b.response.CapacityBytes = capacity
	return b
}

// WithNodeExpansionRequired sets the NodeExpansionRequired against the
// ControllerExpandVolumeResponse instance
func (b *ControllerExpandVolumeResponseBuilder) WithNodeExpansionRequired(
	nodeExpansionRequired bool) *ControllerExpandVolumeResponseBuilder {
	b.response.NodeExpansionRequired = nodeExpansionRequired
	return b
}

// Build returns the constructed instance
// of csi ControllerExpandVolumeResponse
func (b *ControllerExpandVolumeResponseBuilder) Build() *csi.ControllerExpandVolumeResponse {
	return b.response
}
