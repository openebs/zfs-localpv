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
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	k8sapi "github.com/openebs/zfs-localpv/pkg/client/k8s/v1alpha1"
	"github.com/openebs/zfs-localpv/pkg/mgmt/backup"
	"github.com/openebs/zfs-localpv/pkg/mgmt/restore"
	"github.com/openebs/zfs-localpv/pkg/mgmt/snapshot"
	"github.com/openebs/zfs-localpv/pkg/mgmt/volume"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
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

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// start the zfsvolume watcher
	go func() {
		err := volume.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start ZFS volume management controller: %s", err.Error())
		}
	}()

	// start the snapshot watcher
	go func() {
		err := snapshot.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start ZFS volume snapshot management controller: %s", err.Error())
		}
	}()

	// start the backup controller
	go func() {
		err := backup.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start ZFS volume snapshot management controller: %s", err.Error())
		}
	}()

	// start the restore controller
	go func() {
		err := restore.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start ZFS volume snapshot management controller: %s", err.Error())
		}
	}()

	return &node{
		driver: d,
	}
}

// GetVolAndMountInfo get volume and mount info from node csi volume request
func GetVolAndMountInfo(
	req *csi.NodePublishVolumeRequest,
) (*apis.ZFSVolume, *apis.MountInfo, error) {
	var mountinfo apis.MountInfo

	mountinfo.FSType = req.GetVolumeCapability().GetMount().GetFsType()
	mountinfo.MountPath = req.GetTargetPath()
	mountinfo.MountOptions = append(mountinfo.MountOptions, req.GetVolumeCapability().GetMount().GetMountFlags()...)

	if req.GetReadonly() {
		mountinfo.MountOptions = append(mountinfo.MountOptions, "ro")
	}

	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
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
		return nil, status.Error(codes.Internal, err.Error())
	}
	// If the access type is block, do nothing for stage
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		// attempt block mount operation on the requested path
		err = zfs.MountBlock(vol, mountInfo)
	case *csi.VolumeCapability_Mount:
		// attempt filesystem mount operation on the requested path
		err = zfs.MountFilesystem(vol, mountInfo)
	}

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
		err error
		vol *apis.ZFSVolume
	)

	if err = ns.validateNodeUnpublishReq(req); err != nil {
		return nil, err
	}

	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	if vol, err = zfs.GetZFSVolume(volumeID); err != nil {
		return nil, status.Errorf(codes.Internal,
			"not able to get the ZFSVolume %s err : %s",
			volumeID, err.Error())
	}

	err = zfs.UmountVolume(vol, targetPath)

	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"unable to umount the volume %s err : %s",
			volumeID, err.Error())
	}
	klog.Infof("hostpath: volume %s path: %s has been unmounted.",
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

	node, err := k8sapi.GetNode(ns.driver.config.NodeID)
	if err != nil {
		klog.Errorf("failed to get the node %s", ns.driver.config.NodeID)
		return nil, err
	}
	/*
	 * The driver will support all the keys and values defined in the node's label.
	 * if nodes are labeled with the below keys and values
	 * map[beta.kubernetes.io/arch:amd64 beta.kubernetes.io/os:linux kubernetes.io/arch:amd64 kubernetes.io/hostname:pawan-node-1 kubernetes.io/os:linux node-role.kubernetes.io/worker:true openebs.io/zone:zone1 openebs.io/zpool:ssd]
	 * The driver will support below key and values
	 * {
	 *	beta.kubernetes.io/arch:amd64
	 *	beta.kubernetes.io/os:linux
	 *	kubernetes.io/arch:amd64
	 *	kubernetes.io/hostname:pawan-node-1
	 *	kubernetes.io/os:linux
	 *	node-role.kubernetes.io/worker:true
	 *	openebs.io/zone:zone1
	 *	openebs.io/zpool:ssd
	 * }
	 */

	// support all the keys that node has
	topology := node.Labels

	// add driver's topology key
	topology[zfs.ZFSTopologyKey] = ns.driver.config.NodeID

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
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
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

	volumeID := req.GetVolumeId()
	vol, err := zfs.GetZFSVolume(volumeID)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle NodeExpandVolume Request for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}
	if err = zfs.ResizeZFSVolume(vol, req.GetVolumePath()); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle NodeExpandVolume Request for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
	}, nil
}

// NodeGetVolumeStats returns statistics for the
// given volume
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	volID := req.GetVolumeId()
	path := req.GetVolumePath()

	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume id is not provided")
	}
	if len(path) == 0 {
		return nil, status.Error(codes.InvalidArgument, "path is not provided")
	}

	if zfs.IsMountPath(path) == false {
		return nil, status.Error(codes.InvalidArgument, "path is not a mount path")
	}

	var sfs unix.Statfs_t
	if err := unix.Statfs(path, &sfs); err != nil {
		return nil, status.Errorf(codes.Internal, "statfs on %s failed: %v", path, err)
	}

	var usage []*csi.VolumeUsage
	usage = append(usage, &csi.VolumeUsage{
		Unit:      csi.VolumeUsage_BYTES,
		Total:     int64(sfs.Blocks) * sfs.Bsize,
		Used:      int64(sfs.Blocks-sfs.Bfree) * sfs.Bsize,
		Available: int64(sfs.Bavail) * sfs.Bsize,
	})
	usage = append(usage, &csi.VolumeUsage{
		Unit:      csi.VolumeUsage_INODES,
		Total:     int64(sfs.Files),
		Used:      int64(sfs.Files - sfs.Ffree),
		Available: int64(sfs.Ffree),
	})

	return &csi.NodeGetVolumeStatsResponse{Usage: usage}, nil
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
