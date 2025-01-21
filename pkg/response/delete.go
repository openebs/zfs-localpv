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

// DeleteVolumeResponseBuilder helps building an
// instance of csi DeleteVolumeResponse
type DeleteVolumeResponseBuilder struct {
	response *csi.DeleteVolumeResponse
}

// NewDeleteVolumeResponseBuilder returns a new
// instance of DeleteVolumeResponseBuilder
func NewDeleteVolumeResponseBuilder() *DeleteVolumeResponseBuilder {
	return &DeleteVolumeResponseBuilder{
		response: &csi.DeleteVolumeResponse{},
	}
}

// Build returns the constructed instance
// of csi DeleteVolumeResponse
func (b *DeleteVolumeResponseBuilder) Build() *csi.DeleteVolumeResponse {
	return b.response
}

// DeleteSnapshotResponseBuilder helps building an
// instance of csi DeleteSnapshotResponse
type DeleteSnapshotResponseBuilder struct {
	response *csi.DeleteSnapshotResponse
}

// NewDeleteSnapshotResponseBuilder returns a new
// instance of DeleteSnapshotResponseBuilder
func NewDeleteSnapshotResponseBuilder() *DeleteSnapshotResponseBuilder {
	return &DeleteSnapshotResponseBuilder{
		response: &csi.DeleteSnapshotResponse{},
	}
}

// Build returns the constructed instance
// of csi DeleteSnapshotResponse
func (b *DeleteSnapshotResponseBuilder) Build() *csi.DeleteSnapshotResponse {
	return b.response
}
