/*
Copyright © 2019 The OpenEBS Authors

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

// CreateVolumeResponseBuilder helps building an
// instance of csi CreateVolumeResponse
type CreateVolumeResponseBuilder struct {
	response *csi.CreateVolumeResponse
}

// NewCreateVolumeResponseBuilder returns a new
// instance of CreateVolumeResponseBuilder
func NewCreateVolumeResponseBuilder() *CreateVolumeResponseBuilder {
	return &CreateVolumeResponseBuilder{
		response: &csi.CreateVolumeResponse{
			Volume: &csi.Volume{},
		},
	}
}

// WithName sets the name against the
// CreateVolumeResponse instance
func (b *CreateVolumeResponseBuilder) WithName(name string) *CreateVolumeResponseBuilder {
	b.response.Volume.VolumeId = name
	return b
}

// WithCapacity sets the capacity against the
// CreateVolumeResponse instance
func (b *CreateVolumeResponseBuilder) WithCapacity(capacity int64) *CreateVolumeResponseBuilder {
	b.response.Volume.CapacityBytes = capacity
	return b
}

// WithContext sets the context against the
// CreateVolumeResponse instance
func (b *CreateVolumeResponseBuilder) WithContext(ctx map[string]string) *CreateVolumeResponseBuilder {
	b.response.Volume.VolumeContext = ctx
	return b
}

// WithContentSource sets the contentSource against the
// CreateVolumeResponse instance
func (b *CreateVolumeResponseBuilder) WithContentSource(cnt *csi.VolumeContentSource) *CreateVolumeResponseBuilder {
	b.response.Volume.ContentSource = cnt
	return b
}

// WithTopology sets the topology for the
// CreateVolumeResponse instance
func (b *CreateVolumeResponseBuilder) WithTopology(topology map[string]string) *CreateVolumeResponseBuilder {
	b.response.Volume.AccessibleTopology = make([]*csi.Topology, 1)
	b.response.Volume.AccessibleTopology[0] = &csi.Topology{Segments: topology}
	return b
}

// Build returns the constructed instance
// of csi CreateVolumeResponse
func (b *CreateVolumeResponseBuilder) Build() *csi.CreateVolumeResponse {
	return b.response
}
