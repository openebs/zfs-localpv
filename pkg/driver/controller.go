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

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	errors "github.com/openebs/lib-csi/pkg/common/errors"
	"github.com/openebs/lib-csi/pkg/common/helpers"
	schd "github.com/openebs/lib-csi/pkg/scheduler"
	"github.com/openebs/zfs-localpv/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	csipayload "github.com/openebs/zfs-localpv/pkg/response"
	analytics "github.com/openebs/zfs-localpv/pkg/usage"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// size constants
const (
	MB = 1000 * 1000
	GB = 1000 * 1000 * 1000
	Mi = 1024 * 1024
	Gi = 1024 * 1024 * 1024
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
func sendEventOrIgnore(pvcName, pvName, capacity, stgType, method string) {
	if zfs.GoogleAnalyticsEnabled == "true" {
		analytics.New().Build().ApplicationBuilder().
			SetVolumeType(stgType, method).
			SetDocumentTitle(pvName).
			SetCampaignName(pvcName).
			SetLabel(analytics.EventLabelCapacity).
			SetReplicaCount(analytics.LocalPVReplicaCount, method).
			SetCategory(method).
			SetVolumeCapacity(capacity).Send()
	}
}

// getRoundedCapacity rounds the capacity on 1024 base
func getRoundedCapacity(size int64) int64 {

	/*
	 * volblocksize and recordsize must be power of 2 from 512B to 1M
	 * so keeping the size in the form of Gi or Mi should be
	 * sufficient to make volsize multiple of volblocksize/recordsize.
	 */
	if size > Gi {
		return ((size + Gi - 1) / Gi) * Gi
	}

	// Keeping minimum allocatable size as 1Mi (1024 * 1024)
	return ((size + Mi - 1) / Mi) * Mi
}

func waitForReadyVolume(volname string) error {
	for true {
		vol, err := zfs.GetZFSVolume(volname)
		if err != nil {
			return status.Errorf(codes.Internal,
				"zfs: wait failed, not able to get the volume %s %s", volname, err.Error())
		}

		switch vol.Status.State {
		case zfs.ZFSStatusReady:
			return nil
		}
		time.Sleep(time.Second)
	}
	return nil
}

func waitForVolDestroy(volname string) error {
	for true {
		_, err := zfs.GetZFSVolume(volname)
		if err != nil {
			if k8serror.IsNotFound(err) {
				return nil
			}
			return status.Errorf(codes.Internal,
				"zfs: destroy wait failed, not able to get the volume %s %s", volname, err.Error())
		}
		time.Sleep(time.Second)
	}
	return nil
}

func waitForReadySnapshot(snapname string) error {
	for true {
		snap, err := zfs.GetZFSSnapshot(snapname)
		if err != nil {
			return status.Errorf(codes.Internal,
				"zfs: wait failed, not able to get the snapshot %s %s", snapname, err.Error())
		}

		switch snap.Status.State {
		case zfs.ZFSStatusReady:
			return nil
		}
		time.Sleep(time.Second)
	}
	return nil
}

// CreateZFSVolume create new zfs volume from csi volume request
func CreateZFSVolume(req *csi.CreateVolumeRequest) (string, error) {
	volName := strings.ToLower(req.GetName())
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)

	// parameter keys may be mistyped from the CRD specification when declaring
	// the storageclass, which kubectl validation will not catch. Because ZFS
	// parameter keys (not values!) are all lowercase, keys may safely be forced
	// to the lower case.
	originalParams := req.GetParameters()
	parameters := helpers.GetCaseInsensitiveMap(&originalParams)

	rs := parameters["recordsize"]
	bs := parameters["volblocksize"]
	compression := parameters["compression"]
	dedup := parameters["dedup"]
	encr := parameters["encryption"]
	kf := parameters["keyformat"]
	kl := parameters["keylocation"]
	pool := parameters["poolname"]
	tp := parameters["thinprovision"]
	schld := parameters["scheduler"]
	fstype := parameters["fstype"]
	shared := parameters["shared"]

	vtype := zfs.GetVolumeType(fstype)

	capacity := strconv.FormatInt(int64(size), 10)

	if vol, err := zfs.GetZFSVolume(volName); err == nil {
		if vol.DeletionTimestamp != nil {
			if _, ok := parameters["wait"]; ok {
				if err := waitForVolDestroy(volName); err != nil {
					return "", err
				}
			}
		} else {
			if vol.Spec.Capacity != capacity {
				return "", status.Errorf(codes.AlreadyExists,
					"volume %s already present", volName)
			}
			return vol.Spec.OwnerNodeID, nil
		}
	}

	nmap, err := getNodeMap(schld, pool)
	if err != nil {
		return "", status.Errorf(codes.Internal, "get node map failed : %s", err.Error())
	}

	// run the scheduler get the preferred nodelist
	var selected string
	nodelist := schd.Scheduler(req, nmap)
	if len(nodelist) != 0 {
		selected = nodelist[0]
	}
	if len(selected) == 0 {
		// (hack): CSI Sanity test does not pass topology information
		selected = parameters["node"]
		if len(selected) == 0 {
			return "", status.Error(codes.Internal, "scheduler failed, not able to select a node to create the PV")
		}
	}

	klog.Infof("scheduled the volume %s/%s on node %s", pool, volName, selected)

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithCapacity(capacity).
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
		WithVolumeStatus(zfs.ZFSStatusPending).
		WithFsType(fstype).
		WithShared(shared).
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

// CreateVolClone creates the clone from a volume
func CreateVolClone(req *csi.CreateVolumeRequest, srcVol string) (string, error) {
	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	pool := helpers.GetInsensitiveParameter(&parameters, "poolname")
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)
	volsize := strconv.FormatInt(int64(size), 10)

	vol, err := zfs.GetZFSVolume(srcVol)
	if err != nil {
		return "", status.Error(codes.NotFound, err.Error())
	}

	if vol.Spec.PoolName != pool {
		return "", status.Errorf(codes.Internal,
			"clone: different pool src pool %s dst pool %s",
			vol.Spec.PoolName, pool)
	}

	if vol.Spec.Capacity != volsize {
		return "", status.Error(codes.Internal, "clone: volume size is not matching")
	}

	selected := vol.Spec.OwnerNodeID

	labels := map[string]string{zfs.ZFSSrcVolKey: vol.Name}

	// create the clone from the source volume

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithVolumeStatus(zfs.ZFSStatusPending).
		WithLabels(labels).Build()

	volObj.Spec = vol.Spec
	// use the snapshot name same as new volname
	volObj.Spec.SnapName = vol.Name + "@" + volName

	err = zfs.ProvisionVolume(volObj)
	if err != nil {
		return "", status.Errorf(codes.Internal,
			"clone: not able to provision the volume %s", err.Error())
	}

	return selected, nil
}

// CreateSnapClone creates the clone from a snapshot
func CreateSnapClone(req *csi.CreateVolumeRequest, snapshot string) (string, error) {
	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	pool := helpers.GetInsensitiveParameter(&parameters, "poolname")
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)
	volsize := strconv.FormatInt(int64(size), 10)

	snapshotID := strings.Split(snapshot, "@")
	if len(snapshotID) != 2 {
		return "", status.Errorf(
			codes.NotFound,
			"snap name is not valid %s, {%s}",
			snapshot,
			"invalid snapshot name",
		)
	}

	snap, err := zfs.GetZFSSnapshot(snapshotID[1])
	if err != nil {
		return "", status.Error(codes.NotFound, err.Error())
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
		WithName(volName).
		WithVolumeStatus(zfs.ZFSStatusPending).
		Build()

	volObj.Spec = snap.Spec
	volObj.Spec.SnapName = strings.ToLower(snapshot)

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

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	pool := helpers.GetInsensitiveParameter(&parameters, "poolname")
	size := getRoundedCapacity(req.GetCapacityRange().GetRequiredBytes())
	contentSource := req.GetVolumeContentSource()
	pvcName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/name")

	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID := contentSource.GetSnapshot().GetSnapshotId()

		selected, err = CreateSnapClone(req, snapshotID)
	} else if contentSource != nil && contentSource.GetVolume() != nil {
		srcVol := contentSource.GetVolume().GetVolumeId()
		selected, err = CreateVolClone(req, srcVol)
	} else {
		selected, err = CreateZFSVolume(req)
	}

	if err != nil {
		return nil, err
	}

	if _, ok := parameters["wait"]; ok {
		if err := waitForReadyVolume(volName); err != nil {
			return nil, err
		}
	}

	sendEventOrIgnore(pvcName, volName, strconv.FormatInt(int64(size), 10), "zfs-localpv", analytics.VolumeProvision)

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

	klog.Infof("received request to delete volume {%s}", req.VolumeId)

	var (
		err error
	)

	if err = cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := strings.ToLower(req.GetVolumeId())

	// verify if the volume has already been deleted
	vol, err := zfs.GetVolume(volumeID)
	if vol != nil && vol.DeletionTimestamp != nil {
		goto deleteResponse
	}

	if err != nil {
		if k8serror.IsNotFound(err) {
			goto deleteResponse
		}
		return nil, errors.Wrapf(
			err,
			"failed to get volume for {%s}",
			volumeID,
		)
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

	sendEventOrIgnore("", volumeID, vol.Spec.Capacity, "zfs-localpv", analytics.VolumeDeprovision)

deleteResponse:
	return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
}

func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range SupportedVolumeCapabilityAccessModes {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}

	foundAll := true
	for _, c := range volCaps {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}

// TODO Implementation will be taken up later

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeID := strings.ToLower(req.GetVolumeId())
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volCaps := req.GetVolumeCapabilities()
	if len(volCaps) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not provided")
	}

	if _, err := zfs.GetZFSVolume(volumeID); err != nil {
		return nil, status.Errorf(codes.NotFound, "Get volume failed err %s", err.Error())
	}

	var confirmed *csi.ValidateVolumeCapabilitiesResponse_Confirmed
	if isValidVolumeCapabilities(volCaps) {
		confirmed = &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: volCaps}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: confirmed,
	}, nil
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
	volumeID := strings.ToLower(req.GetVolumeId())
	if volumeID == "" {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"ControllerExpandVolume: no volumeID provided",
		)
	}

	/* round off the new size */
	updatedSize := getRoundedCapacity(req.GetCapacityRange().GetRequiredBytes())

	vol, err := zfs.GetZFSVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ControllerExpandVolumeRequest: failed to get ZFSVolume in for %s, {%s}",
			volumeID,
			err.Error(),
		)
	}

	volsize, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"ControllerExpandVolumeRequest: failed to parse volsize in for %s, {%s}",
			volumeID,
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
			volumeID,
			err.Error(),
		)
	}
	return csipayload.NewControllerExpandVolumeResponseBuilder().
		WithCapacityBytes(updatedSize).
		WithNodeExpansionRequired(true).
		Build(), nil
}

func verifySnapshotRequest(req *csi.CreateSnapshotRequest) error {
	snapName := strings.ToLower(req.GetName())
	volumeID := strings.ToLower(req.GetSourceVolumeId())

	if snapName == "" || volumeID == "" {
		return status.Errorf(
			codes.InvalidArgument,
			"CreateSnapshot error invalid request %s: %s",
			volumeID, snapName,
		)
	}

	snap, err := zfs.GetZFSSnapshot(snapName)

	if err != nil {
		if k8serror.IsNotFound(err) {
			return nil
		}
		return status.Errorf(
			codes.NotFound,
			"CreateSnapshot error snap %s %s get failed : %s",
			snapName, volumeID, err.Error(),
		)
	}
	if snap.Labels[zfs.ZFSVolKey] != volumeID {
		return status.Errorf(
			codes.AlreadyExists,
			"CreateSnapshot error snapshot %s already exist for different source vol %s: %s",
			snapName, snap.Labels[zfs.ZFSVolKey], volumeID,
		)
	}
	return nil
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {
	snapName := strings.ToLower(req.GetName())
	volumeID := strings.ToLower(req.GetSourceVolumeId())

	klog.Infof("CreateSnapshot volume %s@%s", volumeID, snapName)

	err := verifySnapshotRequest(req)
	if err != nil {
		return nil, err
	}

	snapTimeStamp := time.Now().Unix()
	state, err := zfs.GetZFSSnapshotStatus(snapName)

	if err == nil {
		return csipayload.NewCreateSnapshotResponseBuilder().
			WithSourceVolumeID(volumeID).
			WithSnapshotID(volumeID+"@"+snapName).
			WithCreationTime(snapTimeStamp, 0).
			WithReadyToUse(state == zfs.ZFSStatusReady).
			Build(), nil
	}

	vol, err := zfs.GetZFSVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"CreateSnapshot not able to get volume %s: %s, {%s}",
			volumeID, snapName,
			err.Error(),
		)
	}

	labels := map[string]string{zfs.ZFSVolKey: vol.Name}

	snapObj, err := snapbuilder.NewBuilder().
		WithName(snapName).
		WithLabels(labels).Build()

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create snapshotobject for %s: %s, {%s}",
			volumeID, snapName,
			err.Error(),
		)
	}

	snapObj.Spec = vol.Spec
	snapObj.Status.State = zfs.ZFSStatusPending

	if err := zfs.ProvisionSnapshot(snapObj); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s: %s, {%s}",
			volumeID, snapName,
			err.Error(),
		)
	}

	originalParams := req.GetParameters()
	parameters := helpers.GetCaseInsensitiveMap(&originalParams)

	if _, ok := parameters["wait"]; ok {
		if err := waitForReadySnapshot(snapName); err != nil {
			return nil, err
		}
	}

	state, _ = zfs.GetZFSSnapshotStatus(snapName)

	return csipayload.NewCreateSnapshotResponseBuilder().
		WithSourceVolumeID(volumeID).
		WithSnapshotID(volumeID+"@"+snapName).
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

	if req.SnapshotId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "DeleteSnapshot: empty snapshotID")
	}

	klog.Infof("DeleteSnapshot request for %s", req.SnapshotId)

	// snapshodID is formed as <volname>@<snapname>
	// parsing them here
	snapshotID := strings.Split(req.SnapshotId, "@")
	if len(snapshotID) != 2 {
		// should succeed when an invalid snapshot id is used
		return &csi.DeleteSnapshotResponse{}, nil
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
