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

package driver

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/zfs-localpv/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	errors "github.com/openebs/zfs-localpv/pkg/common/errors"
	csipayload "github.com/openebs/zfs-localpv/pkg/response"
	analytics "github.com/openebs/zfs-localpv/pkg/usage"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// controller is the server implementation
// for CSI Controller
type controller struct {
	driver       *CSIDriver
	capabilities []*csi.ControllerServiceCapability
}

// NewController returns a new instance
// of CSI controller
func NewController(d *CSIDriver) csi.ControllerServer {
	return &controller{
		driver:       d,
		capabilities: newControllerCapabilities(),
	}
}

// SupportedVolumeCapabilityAccessModes contains the list of supported access
// modes for the volume
var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	&csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

// sendEventOrIgnore sends anonymous local-pv provision/delete events
func sendEventOrIgnore(pvName, capacity, stgType, method string) {
	if zfs.GoogleAnalyticsEnabled == "true" {
		analytics.New().Build().ApplicationBuilder().
			SetVolumeType(stgType, method).
			SetDocumentTitle(pvName).
			SetLabel(analytics.EventLabelCapacity).
			SetReplicaCount(analytics.LocalPVReplicaCount, method).
			SetCategory(method).
			SetVolumeCapacity(capacity).Send()
	}
}

func CreateZFSVolume(req *csi.CreateVolumeRequest) (string, error) {
	volName := req.GetName()
	size := req.GetCapacityRange().RequiredBytes
	rs := req.GetParameters()["recordsize"]
	bs := req.GetParameters()["volblocksize"]
	compression := req.GetParameters()["compression"]
	dedup := req.GetParameters()["dedup"]
	encr := req.GetParameters()["encryption"]
	kf := req.GetParameters()["keyformat"]
	kl := req.GetParameters()["keylocation"]
	pool := req.GetParameters()["poolname"]
	tp := req.GetParameters()["thinprovision"]
	schld := req.GetParameters()["scheduler"]
	fstype := req.GetParameters()["fstype"]

	vtype := zfs.GetVolumeType(fstype)

	selected := scheduler(req.AccessibilityRequirements, schld, pool)

	if len(selected) == 0 {
		return "", status.Error(codes.Internal, "scheduler failed")
	}

	logrus.Infof("scheduled the volume %s/%s on node %s", pool, volName, selected)

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithCapacity(strconv.FormatInt(int64(size), 10)).
		WithRecordSize(rs).
		WithVolBlockSize(bs).
		WithPoolName(pool).
		WithDedup(dedup).
		WithEncryption(encr).
		WithKeyFormat(kf).
		WithKeyLocation(kl).
		WithThinProv(tp).
		WithOwnerNode(selected).
		WithVolumeType(vtype).
		WithFsType(fstype).
		WithCompression(compression).Build()

	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	err = zfs.ProvisionVolume(volObj)
	if err != nil {
		return "", status.Errorf(codes.Internal,
			"not able to provision the volume %s", err.Error())
	}

	return selected, nil
}

func CreateZFSClone(req *csi.CreateVolumeRequest, snapshot string) (string, error) {

	volName := req.GetName()
	pool := req.GetParameters()["poolname"]
	size := req.GetCapacityRange().RequiredBytes
	volsize := strconv.FormatInt(int64(size), 10)

	snapshotID := strings.Split(snapshot, "@")
	if len(snapshotID) != 2 {
		return "", status.Errorf(
			codes.Internal,
			"snap name is not valid %s, {%s}",
			snapshot,
			"invalid snapshot name",
		)
	}

	snap, err := zfs.GetZFSSnapshot(snapshotID[1])
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	if snap.Spec.PoolName != pool {
		return "", status.Errorf(codes.Internal,
			"clone to a different pool src pool %s dst pool %s",
			snap.Spec.PoolName, pool)
	}

	if snap.Spec.Capacity != volsize {
		return "", status.Error(codes.Internal, "clone volume size is not matching")
	}

	selected := snap.Spec.OwnerNodeID

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).Build()

	volObj.Spec = snap.Spec
	volObj.Spec.SnapName = snapshot

	err = zfs.ProvisionVolume(volObj)
	if err != nil {
		return "", status.Errorf(codes.Internal,
			"not able to provision the clone volume %s", err.Error())
	}

	return selected, nil
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	var err error
	var selected string

	volName := req.GetName()
	pool := req.GetParameters()["poolname"]
	size := req.GetCapacityRange().RequiredBytes

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	contentSource := req.GetVolumeContentSource()
	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID := contentSource.GetSnapshot().GetSnapshotId()

		selected, err = CreateZFSClone(req, snapshotID)
	} else {
		selected, err = CreateZFSVolume(req)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sendEventOrIgnore(volName, strconv.FormatInt(int64(size), 10), "zfs-localpv", analytics.VolumeProvision)

	topology := map[string]string{zfs.ZFSTopologyKey: selected}
	cntx := map[string]string{zfs.PoolNameKey: pool}

	return csipayload.NewCreateVolumeResponseBuilder().
		WithName(volName).
		WithCapacity(size).
		WithTopology(topology).
		WithContext(cntx).
		WithContentSource(contentSource).
		Build(), nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	logrus.Infof("received request to delete volume {%s}", req.VolumeId)

	var (
		err error
	)

	if err = cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()

	// verify if the volume has already been deleted
	vol, err := zfs.GetVolume(volumeID)
	if vol != nil && vol.DeletionTimestamp != nil {
		goto deleteResponse
	}

	// Delete the corresponding ZV CR
	err = zfs.DeleteVolume(volumeID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			volumeID,
		)
	}

	sendEventOrIgnore(volumeID, vol.Spec.Capacity, "zfs-localpv", analytics.VolumeDeprovision)

deleteResponse:
	return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
}

// TODO Implementation will be taken up later

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities fetches controller capabilities
//
// This implements csi.ControllerServer
func (cs *controller) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.capabilities,
	}

	return resp, nil
}

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {

	updatedSize := req.GetCapacityRange().GetRequiredBytes()

	vol, err := zfs.GetZFSVolume(req.VolumeId)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ControllerExpandVolumeRequest: failed to get ZFSVolume in for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}

	volsize, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ControllerExpandVolumeRequest: failed to parse volsize in for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}
	/*
	 * Controller expand volume must be idempotent. If a volume corresponding
	 * to the specified volume ID is already larger than or equal to the target
	 * capacity of the expansion request, the plugin should reply 0 OK.
	 */
	if volsize >= updatedSize {
		return csipayload.NewControllerExpandVolumeResponseBuilder().
			WithCapacityBytes(volsize).
			Build(), nil
	}

	if err := zfs.ResizeVolume(vol, updatedSize); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle ControllerExpandVolumeRequest for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}
	return csipayload.NewControllerExpandVolumeResponseBuilder().
		WithCapacityBytes(updatedSize).
		WithNodeExpansionRequired(true).
		Build(), nil
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	logrus.Infof("CreateSnapshot volume %s@%s", req.SourceVolumeId, req.Name)

	snapTimeStamp := time.Now().Unix()
	state, err := zfs.GetZFSSnapshotStatus(req.Name)

	if err == nil {
		return csipayload.NewCreateSnapshotResponseBuilder().
			WithSourceVolumeID(req.SourceVolumeId).
			WithSnapshotID(req.SourceVolumeId+"@"+req.Name).
			WithCreationTime(snapTimeStamp, 0).
			WithReadyToUse(state == zfs.ZFSStatusReady).
			Build(), nil
	}

	vol, err := zfs.GetZFSVolume(req.SourceVolumeId)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"CreateSnapshot not able to get volume %s: %s, {%s}",
			req.SourceVolumeId, req.Name,
			err.Error(),
		)
	}

	labels := map[string]string{zfs.ZFSVolKey: vol.Name}

	snapObj, err := snapbuilder.NewBuilder().
		WithName(req.Name).
		WithLabels(labels).Build()

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create snapshotobject for %s: %s, {%s}",
			req.SourceVolumeId, req.Name,
			err.Error(),
		)
	}

	snapObj.Spec = vol.Spec
	snapObj.Status.State = zfs.ZFSStatusPending

	if err := zfs.ProvisionSnapshot(snapObj); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s: %s, {%s}",
			req.SourceVolumeId, req.Name,
			err.Error(),
		)
	}

	state, _ = zfs.GetZFSSnapshotStatus(req.Name)

	return csipayload.NewCreateSnapshotResponseBuilder().
		WithSourceVolumeID(req.SourceVolumeId).
		WithSnapshotID(req.SourceVolumeId+"@"+req.Name).
		WithCreationTime(snapTimeStamp, 0).
		WithReadyToUse(state == zfs.ZFSStatusReady).
		Build(), nil
}

// DeleteSnapshot deletes given snapshot
//
// This implements csi.ControllerServer
func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	logrus.Infof("DeleteSnapshot request for %s", req.SnapshotId)

	// snapshodID is formed as <volname>@<snapname>
	// parsing them here
	snapshotID := strings.Split(req.SnapshotId, "@")
	if len(snapshotID) != 2 {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle DeleteSnapshot for %s, {%s}",
			req.SnapshotId,
			"failed to get the snapshot name, Manual intervention required",
		)
	}
	if err := zfs.DeleteSnapshot(snapshotID[1]); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle DeleteSnapshot for %s, {%s}",
			req.SnapshotId,
			err.Error(),
		)
	}
	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots lists all snapshots for the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume removes a previously
// attached volume from the given node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume attaches given volume
// at the specified node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes lists all the volumes
//
// This implements csi.ControllerServer
func (cs *controller) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// validateCapabilities validates if provided capabilities
// are supported by this driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {

	for _, cap := range caps {
		if !IsSupportedVolumeCapabilityAccessMode(cap.AccessMode.Mode) {
			return false
		}
	}
	return true
}

func (cs *controller) validateDeleteVolumeReq(req *csi.DeleteVolumeRequest) error {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle delete volume request: missing volume id",
		)
	}

	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s} : validation failed",
			volumeID,
		)
	}
	return nil
}

// IsSupportedVolumeCapabilityAccessMode valides the requested access mode
func IsSupportedVolumeCapabilityAccessMode(
	accessMode csi.VolumeCapability_AccessMode_Mode,
) bool {

	for _, access := range SupportedVolumeCapabilityAccessModes {
		if accessMode == access.Mode {
			return true
		}
	}
	return false
}

// newControllerCapabilities returns a list
// of this controller's capabilities
func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(
		cap csi.ControllerServiceCapability_RPC_Type,
	) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

// validateRequest validates if the requested service is
// supported by the driver
func (cs *controller) validateRequest(
	c csi.ControllerServiceCapability_RPC_Type,
) error {

	for _, cap := range cs.capabilities {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}

	return status.Error(
		codes.InvalidArgument,
		fmt.Sprintf("failed to validate request: {%s} is not supported", c),
	)
}

func (cs *controller) validateVolumeCreateReq(req *csi.CreateVolumeRequest) error {
	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle create volume request for {%s}",
			req.GetName(),
		)
	}

	if req.GetName() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume name",
		)
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume capabilities",
		)
	}
	return nil
}
