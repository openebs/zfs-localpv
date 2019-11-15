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
	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	ctrl "github.com/openebs/zfs-localpv/cmd/controller"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
	"github.com/openebs/zfs-localpv/pkg/builder"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

// node is the server implementation
// for CSI NodeServer
type node struct {
	driver *CSIDriver
}

// NewNode returns a new instance
// of CSI NodeServer
func NewNode(d *CSIDriver) csi.NodeServer {
	var ControllerMutex = sync.RWMutex{}
	// start the zfsvolume watcher
	go func() {
		err := ctrl.Start(&ControllerMutex)
		if err != nil {
			logrus.Errorf("Failed to start cstorvolume claim controller: %s", err.Error())
		}
	}()

	return &node{
		driver: d,
	}
}

func GetVolAndMountInfo(
	req *csi.NodePublishVolumeRequest,
) (*apis.ZFSVolume, *apis.MountInfo, error) {
	var mountinfo apis.MountInfo

	mountinfo.FSType = req.GetVolumeCapability().GetMount().GetFsType()
	mountinfo.MountPath = req.GetTargetPath()
	mountinfo.ReadOnly = req.GetReadonly()
	mountinfo.MountOptions = append(mountinfo.MountOptions, req.GetVolumeCapability().GetMount().GetMountFlags()...)

	getOptions := metav1.GetOptions{}
	vol, err := builder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		Get(req.GetVolumeId(), getOptions)

	if err != nil {
		return nil, nil, err
	}

	return vol, &mountinfo, nil
}

// NodePublishVolume publishes (mounts) the volume
// at the corresponding node at a given path
//
// This implements csi.NodeServer
func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	var (
		err error
	)

	if err = ns.validateNodePublishReq(req); err != nil {
		return nil, err
	}

	vol, mountInfo, err := GetVolAndMountInfo(req)
	if err != nil {
		goto PublishVolumeResponse
	}
	// Create the zfs volume and attempt mount operation on the requested path
	if err = zfs.MountVolume(vol, mountInfo); err != nil {
		goto PublishVolumeResponse
	}

PublishVolumeResponse:
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unpublishes (unmounts) the volume
// from the corresponding node from the given path
//
// This implements csi.NodeServer
func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	var (
		err           error
		vol           *apis.ZFSVolume
		currentMounts []string
	)

	if err = ns.validateNodeUnpublishReq(req); err != nil {
		return nil, err
	}

	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	if vol, err = zfs.GetZFSVolume(volumeID); err != nil {
		return nil, err
	}

	devpath, _ := zfs.GetVolumeDevPath(vol)
	currentMounts, err = zfs.GetMounts(devpath)
	if err != nil {
		return nil, err
	} else if len(currentMounts) == 0 {
		goto NodeUnpublishResponse
	} else if len(currentMounts) == 1 {
		if currentMounts[0] != targetPath {
			return nil, status.Error(codes.Internal, "device not mounted at right path")
		}
	} else {
		logrus.Errorf(
			"can not unmount, more than one mounts for volume:%s path %s mounts: %v",
			volumeID, targetPath, currentMounts,
		)
		return nil, status.Error(codes.Internal, "device not mounted at rightpath")
	}

	if err = zfs.UmountVolume(vol, req.GetTargetPath()); err != nil {
		goto NodeUnpublishResponse
	}

NodeUnpublishResponse:
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	logrus.Infof("hostpath: volume %s path: %s has been unmounted.",
		volumeID, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo returns node details
//
// This implements csi.NodeServer
func (ns *node) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {

	topology := map[string]string{zfs.ZFSTopologyKey: ns.driver.config.NodeID}
	return &csi.NodeGetInfoResponse{
		NodeId: ns.driver.config.NodeID,
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// NodeGetCapabilities returns capabilities supported
// by this node service
//
// This implements csi.NodeServer
func (ns *node) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

// TODO
// This needs to be implemented
//
// NodeStageVolume mounts the volume on the staging
// path
//
// This implements csi.NodeServer
func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unmounts the volume from
// the staging path
//
// This implements csi.NodeServer
func (ns *node) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// TODO
// Verify if this needs to be implemented
//
// NodeExpandVolume resizes the filesystem if required
//
// If ControllerExpandVolumeResponse returns true in
// node_expansion_required then FileSystemResizePending
// condition will be added to PVC and NodeExpandVolume
// operation will be queued on kubelet
//
// This implements csi.NodeServer
func (ns *node) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {

	return nil, nil
}

// NodeGetVolumeStats returns statistics for the
// given volume
//
// This implements csi.NodeServer
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	in *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *node) validateNodePublishReq(
	req *csi.NodePublishVolumeRequest,
) error {
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}

	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	return nil
}

func (ns *node) validateNodeUnpublishReq(
	req *csi.NodeUnpublishVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}
