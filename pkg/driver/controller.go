package driver

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	analytics "github.com/openebs/google-analytics-4/usage"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	"github.com/openebs/lib-csi/pkg/common/errors"
	"github.com/openebs/lib-csi/pkg/common/helpers"
	schd "github.com/openebs/lib-csi/pkg/scheduler"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	zfsapi "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	informers "github.com/openebs/zfs-localpv/pkg/generated/informer/externalversions"
	csipayload "github.com/openebs/zfs-localpv/pkg/response"
	"github.com/openebs/zfs-localpv/pkg/version"
	"github.com/openebs/zfs-localpv/pkg/zfs"
)

// size constants
const (
	MB = 1000 * 1000
	GB = 1000 * 1000 * 1000
	Mi = 1024 * 1024
	Gi = 1024 * 1024 * 1024

	// Ping event is sent periodically
	Ping string = "zfs-ping"
	// Heartbeat message.
	Heartbeat string = "zfs-heartbeat"
	// DefaultCASType Event application name constant for volume event
	DefaultCASType string = "zfs-localpv"

	// LocalPVReplicaCount is the constant used by usage to represent
	// replication factor in LocalPV
	LocalPVReplicaCount string = "1"
)

// controller is the server implementation
// for CSI Controller
type controller struct {
	driver       *CSIDriver
	capabilities []*csi.ControllerServiceCapability

	indexedLabel string

	k8sNodeInformer cache.SharedIndexInformer
	zfsNodeInformer cache.SharedIndexInformer

	volumeLock *volumeLock
}

// NewController returns a new instance
// of CSI controller
func NewController(d *CSIDriver) csi.ControllerServer {
	ctrl := &controller{
		driver:       d,
		capabilities: newControllerCapabilities(),
		volumeLock:   newVolumeLock(),
	}
	if err := ctrl.init(); err != nil {
		klog.Fatalf("init controller: %v", err)
	}

	return ctrl
}

func (cs *controller) init() error {
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return errors.Wrapf(err, "failed to build kubeconfig")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to build k8s clientset")
	}

	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to build openebs clientset")
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, 0)
	openebsInformerfactory := informers.NewSharedInformerFactoryWithOptions(openebsClient,
		0, informers.WithNamespace(zfs.OpenEBSNamespace))

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cs.k8sNodeInformer = kubeInformerFactory.Core().V1().Nodes().Informer()
	cs.zfsNodeInformer = openebsInformerfactory.Zfs().V1().ZFSNodes().Informer()

	if err = cs.zfsNodeInformer.AddIndexers(map[string]cache.IndexFunc{
		LabelIndexName(cs.indexedLabel): LabelIndexFunc(cs.indexedLabel),
	}); err != nil {
		return errors.Wrapf(err, "failed to add index on label %v", cs.indexedLabel)
	}

	go cs.k8sNodeInformer.Run(stopCh)
	go cs.zfsNodeInformer.Run(stopCh)

	if zfs.GoogleAnalyticsEnabled == "true" {
		analytics.RegisterVersionGetter(version.GetVersionDetails)
		analytics.New().CommonBuild(DefaultCASType).InstallBuilder(true).Send()
		go analytics.PingCheck(DefaultCASType, Ping, false)
		go analytics.PingCheck(DefaultCASType, Heartbeat, true)
	}

	// wait for all the caches to be populated.
	klog.Info("waiting for k8s & zfs node informer caches to be synced")
	cache.WaitForCacheSync(stopCh,
		cs.k8sNodeInformer.HasSynced,
		cs.zfsNodeInformer.HasSynced)
	klog.Info("synced k8s & zfs node informer caches")
	return nil
}

// SupportedVolumeCapabilityAccessModes contains the list of supported access
// modes for the volume
var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

// sendEventOrIgnore sends anonymous local-pv provision/delete events
func sendEventOrIgnore(pvcName, pvName, capacity, method string) {
	if zfs.GoogleAnalyticsEnabled == "true" {
		analytics.New().CommonBuild(DefaultCASType).ApplicationBuilder().
			SetVolumeName(pvName).
			SetVolumeClaimName(pvcName).
			SetReplicaCount(LocalPVReplicaCount).
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

func waitForVolDestroy(volname string) error {
	for {
		_, err := zfs.GetZFSVolume(volname)
		if err != nil {
			if k8serror.IsNotFound(err) {
				return nil
			}
			return status.Errorf(codes.Internal,
				"zfs: destroy wait failed, not able to get the volume %s %s", volname, err.Error())
		}
		time.Sleep(time.Second)
		klog.Infof("waiting for volume to be destroyed %s", volname)
	}
}

func waitForReadySnapshot(snapname string) error {
	for {
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
}

// CreateZFSVolume create new zfs volume from csi volume request
func CreateZFSVolume(ctx context.Context, req *csi.CreateVolumeRequest) (string, error) {
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
	quotatype := parameters["quotatype"]

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
			if vol.Status.State != zfs.ZFSStatusReady {
				return "", status.Errorf(codes.Aborted,
					"volume %s request already pending", volName)
			}
			return vol.Spec.OwnerNodeID, nil
		}
	}

	nmap, err := getNodeMap(schld, pool)
	if err != nil {
		return "", status.Errorf(codes.Internal, "get node map failed : %s", err.Error())
	}

	var prfList []string

	if node, ok := parameters["node"]; ok {
		// (hack): CSI Sanity test does not pass topology information
		prfList = append(prfList, node)
	} else {
		// run the scheduler
		prfList = schd.Scheduler(req, nmap)
	}

	if len(prfList) == 0 {
		return "", status.Error(codes.Internal, "scheduler failed, node list is empty for creating the PV")
	}

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
		WithVolumeType(vtype).
		WithVolumeStatus(zfs.ZFSStatusPending).
		WithFsType(fstype).
		WithQuotaType(quotatype).
		WithShared(shared).
		WithCompression(compression).Build()

	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	klog.Infof("zfs: trying volume creation %s/%s on node %s", pool, volName, prfList)

	// try volume creation sequentially on all nodes
	for _, node := range prfList {
		var nodeid string
		nodeid, err = zfs.GetNodeID(node)
		if err != nil {
			continue
		}

		vol, _ := volbuilder.BuildFrom(volObj).WithOwnerNodeID(nodeid).WithVolumeStatus(zfs.ZFSStatusPending).Build()

		timeout := false

		timeout, err = zfs.ProvisionVolume(ctx, vol)
		if err == nil {
			return nodeid, nil
		}

		// if timeout reached, return the error and let csi retry the volume creation
		if timeout {
			break
		}
	}

	if err != nil {
		// volume provisioning failed, delete the zfs volume resource
		zfs.DeleteVolume(volName) // ignore error
	}

	return "", status.Errorf(codes.Internal,
		"not able to provision the volume, nodes %v, err : %s", prfList, err.Error())
}

// CreateVolClone creates the clone from a volume
func CreateVolClone(ctx context.Context, req *csi.CreateVolumeRequest, srcVol string) (string, error) {
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
	if err != nil {
		return "", err
	}

	volObj.Spec = vol.Spec
	// use the snapshot name same as new volname
	volObj.Spec.SnapName = vol.Name + "@" + volName

	_, err = zfs.ProvisionVolume(ctx, volObj)
	if err != nil {
		return "", status.Errorf(codes.Internal,
			"clone: not able to provision the volume err : %s", err.Error())
	}

	return selected, nil
}

// CreateSnapClone creates the clone from a snapshot
func CreateSnapClone(ctx context.Context, req *csi.CreateVolumeRequest, snapshot string) (string, error) {
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
	if err != nil {
		return "", err
	}

	volObj.Spec = snap.Spec
	volObj.Spec.SnapName = strings.ToLower(snapshot)

	_, err = zfs.ProvisionVolume(ctx, volObj)
	if err != nil {
		return "", status.Errorf(codes.Internal,
			"not able to provision the clone volume err : %s", err.Error())
	}

	return selected, nil
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	var err error
	var selectedNodeId string

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

	unlock := cs.volumeLock.LockVolume(volName)
	defer unlock()

	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID := contentSource.GetSnapshot().GetSnapshotId()

		selectedNodeId, err = CreateSnapClone(ctx, req, snapshotID)
	} else if contentSource != nil && contentSource.GetVolume() != nil {
		srcVol := contentSource.GetVolume().GetVolumeId()
		selectedNodeId, err = CreateVolClone(ctx, req, srcVol)
	} else {
		selectedNodeId, err = CreateZFSVolume(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	klog.Infof("created the volume %s/%s on node %s", pool, volName, selectedNodeId)

	sendEventOrIgnore(pvcName, volName, strconv.FormatInt(int64(size), 10), analytics.VolumeProvision)

	topology := map[string]string{zfs.ZFSTopologyKey: selectedNodeId}
	cntx := map[string]string{zfs.PoolNameKey: pool, zfs.OpenEBSCasTypeKey: zfs.ZFSCasTypeName}

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

	if err := cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := strings.ToLower(req.GetVolumeId())
	unlock := cs.volumeLock.LockVolume(volumeID)
	defer unlock()

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

	// if volume is not ready, create volume will delete it
	if vol.Status.State != zfs.ZFSStatusReady {
		return nil, status.Error(codes.Internal, "can not delete, volume creation is in progress")
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

	sendEventOrIgnore("", volumeID, vol.Spec.Capacity, analytics.VolumeDeprovision)

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
	unlock := cs.volumeLock.LockVolume(volumeID)
	defer unlock()

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
	unlock := cs.volumeLock.LockVolumeWithSnapshot(volumeID, snapName)
	defer unlock()

	snapTimeStamp := time.Now().Unix()
	var state string
	if snapObj, err := zfs.GetZFSSnapshot(snapName); err == nil {
		state = snapObj.Status.State
		size, err := zfs.GetZFSSnapshotCapacity(snapObj)
		if err != nil {
			return nil, fmt.Errorf("get zfssnapshot capacity failed: %v, capacity: %v", err, snapObj.Spec.Capacity)
		}
		return csipayload.NewCreateSnapshotResponseBuilder().
			WithSourceVolumeID(volumeID).
			WithSnapshotID(volumeID+"@"+snapName).
			WithSize(size).
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

	snapObj, err = zfs.GetZFSSnapshot(snapName)
	if err != nil {
		return nil, fmt.Errorf("get zfssnapshot failed, err: %v", err)
	}
	state = snapObj.Status.State
	size, err := zfs.GetZFSSnapshotCapacity(snapObj)
	if err != nil {
		return nil, fmt.Errorf("get zfssnapshot capacity failed: %v, capacity: %v", err, snapObj.Spec.Capacity)
	}

	return csipayload.NewCreateSnapshotResponseBuilder().
		WithSourceVolumeID(volumeID).
		WithSnapshotID(volumeID+"@"+snapName).
		WithSize(size).
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
	unlock := cs.volumeLock.LockVolumeWithSnapshot(snapshotID[0], snapshotID[1])
	defer unlock()
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
// given node topology segment.
//
// This implements csi.ControllerServer
func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {

	var segments map[string]string
	if topology := req.GetAccessibleTopology(); topology != nil {
		segments = topology.Segments
	}
	nodeNames, err := cs.filterNodesByTopology(segments)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	zfsNodesCache := cs.zfsNodeInformer.GetIndexer()

	params := req.GetParameters()

	poolParam := helpers.GetInsensitiveParameter(&params, "poolname")

	// The "poolname" parameter can either be the name of a ZFS pool
	// (e.g. "zpool"), or a path to a child dataset (e.g. "zpool/k8s/localpv").
	//
	// We parse the "poolname" parameter so the name of the ZFS pool and the
	// path to the dataset is available separately.
	//
	// The dataset path is not used now. It could be used later to query the
	// capacity of the child dataset, which could be smaller than the capacity
	// of the whole pool.
	//
	// This is necessary because capacity calculation currently only works with
	// ZFS pool names. This is why it always returns the capacitry of the whole
	// pool, even if the child dataset given as the "poolname" parameter has a
	// smaller capacity than the whole pool.
	poolname, _ := func() (string, string) {
		poolParamSliced := strings.SplitN(poolParam, "/", 2)
		if len(poolParamSliced) == 2 {
			return poolParamSliced[0], poolParamSliced[1]
		} else {
			return poolParamSliced[0], ""
		}
	}()

	var availableCapacity int64
	for _, nodeName := range nodeNames {
		mappedNodeId, mapErr := zfs.GetNodeID(nodeName)
		if mapErr != nil {
			klog.Warningf("Unable to find mapped node id for %s", nodeName)
			mappedNodeId = nodeName
		}
		v, exists, err := zfsNodesCache.GetByKey(zfs.OpenEBSNamespace + "/" + mappedNodeId)
		if err != nil {
			klog.Warning("unexpected error after querying the zfsNode informer cache")
			continue
		}
		if !exists {
			continue
		}
		zfsNode := v.(*zfsapi.ZFSNode)
		// rather than summing all free capacity, we are calculating maximum
		// zv size that gets fit in given pool.
		// See https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1472-storage-capacity-tracking#available-capacity-vs-maximum-volume-size &
		// https://github.com/container-storage-interface/spec/issues/432 for more details
		for _, zpool := range zfsNode.Pools {
			if zpool.Name != poolname {
				continue
			}
			freeCapacity := zpool.Free.Value()
			if availableCapacity < freeCapacity {
				availableCapacity = freeCapacity
			}
		}
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: availableCapacity,
	}, nil
}

func (cs *controller) filterNodesByTopology(segments map[string]string) ([]string, error) {
	nodesCache := cs.k8sNodeInformer.GetIndexer()
	if len(segments) == 0 {
		return nodesCache.ListKeys(), nil
	}

	filterNodes := func(vs []interface{}) ([]string, error) {
		var names []string
		selector := labels.SelectorFromSet(segments)
		for _, v := range vs {
			meta, err := apimeta.Accessor(v)
			if err != nil {
				return nil, err
			}
			if selector.Matches(labels.Set(meta.GetLabels())) {
				names = append(names, meta.GetName())
			}
		}
		return names, nil
	}

	// first see if we need to filter the informer cache by indexed label,
	// so that we don't need to iterate over all the nodes for performance
	// reasons in large cluster.
	indexName := LabelIndexName(cs.indexedLabel)
	if _, ok := nodesCache.GetIndexers()[indexName]; !ok {
		// run through all the nodes in case indexer doesn't exists.
		return filterNodes(nodesCache.List())
	}

	if segValue, ok := segments[cs.indexedLabel]; ok {
		vs, err := nodesCache.ByIndex(indexName, segValue)
		if err != nil {
			return nil, errors.Wrapf(err, "query indexed store indexName=%v indexKey=%v",
				indexName, segValue)
		}
		return filterNodes(vs)
	}
	return filterNodes(nodesCache.List())
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
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
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

// LabelIndexName add prefix for label index.
func LabelIndexName(label string) string {
	return "l:" + label
}

// LabelIndexFunc defines index values for given label.
func LabelIndexFunc(label string) cache.IndexFunc {
	return func(obj interface{}) ([]string, error) {
		meta, err := apimeta.Accessor(obj)
		if err != nil {
			return nil, fmt.Errorf(
				"k8s api object type (%T) doesn't implements metav1.Object interface: %v", obj, err)
		}
		var vs []string
		if v, ok := meta.GetLabels()[label]; ok {
			vs = append(vs, v)
		}
		return vs, nil
	}
}
