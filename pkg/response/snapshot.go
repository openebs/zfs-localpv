/*
Copyright © 2020 The OpenEBS Authors

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
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

// CreateSnapshotResponseBuilder helps building an
// instance of csi CreateVolumeResponse
type CreateSnapshotResponseBuilder struct {
	response *csi.CreateSnapshotResponse
}

// NewCreateSnapshotResponseBuilder returns a new
// instance of CreateSnapshotResponseBuilder
func NewCreateSnapshotResponseBuilder() *CreateSnapshotResponseBuilder {
	return &CreateSnapshotResponseBuilder{
		response: &csi.CreateSnapshotResponse{
			Snapshot: &csi.Snapshot{},
		},
	}
}

// WithSize sets the size against the
// CreateSnapshotResponse instance
func (b *CreateSnapshotResponseBuilder) WithSize(size int64) *CreateSnapshotResponseBuilder {
	b.response.Snapshot.SizeBytes = size
	return b
}

// WithSnapshotID sets the snapshotID against the
// CreateSnapshotResponse instance
func (b *CreateSnapshotResponseBuilder) WithSnapshotID(snapshotID string) *CreateSnapshotResponseBuilder {
	b.response.Snapshot.SnapshotId = snapshotID
	return b
}

// WithSourceVolumeID sets the sourceVolumeID against the
// CreateSnapshotResponse instance
func (b *CreateSnapshotResponseBuilder) WithSourceVolumeID(volumeID string) *CreateSnapshotResponseBuilder {
	b.response.Snapshot.SourceVolumeId = volumeID
	return b
}

// WithCreationTime sets the creationTime against the
// CreateSnapshotResponse instance
func (b *CreateSnapshotResponseBuilder) WithCreationTime(tsec, tnsec int64) *CreateSnapshotResponseBuilder {
	b.response.Snapshot.CreationTime = &timestamp.Timestamp{
		Seconds: tsec,
		Nanos:   int32(tnsec),
	}
	return b
}

// WithReadyToUse sets the readyToUse feild against the
// CreateSnapshotResponse instance
func (b *CreateSnapshotResponseBuilder) WithReadyToUse(readyToUse bool) *CreateSnapshotResponseBuilder {
	b.response.Snapshot.ReadyToUse = readyToUse
	return b
}

// Build returns the constructed instance
// of csi CreateSnapshotResponse
func (b *CreateSnapshotResponseBuilder) Build() *csi.CreateSnapshotResponse {
	return b.response
}
