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
	ctrl "sigs.k8s.io/controller-runtime"

	zfsapi "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/v2/pkg/builder/volbuilder"
	clientset "github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned"
	informers "github.com/openebs/zfs-localpv/v2/pkg/generated/informer/externalversions"
	csipayload "github.com/openebs/zfs-localpv/v2/pkg/response"
	"github.com/openebs/zfs-localpv/v2/pkg/version"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
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

	csi.UnimplementedControllerServer
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
	// TODO: (tech-debt) Setup signal handler more above, several files want to use stopCh and this function is only allowed to be used once (see #647)
	// Affected files: pkg/driver/agent.go pkg/driver/controller.go pkg/driver/grpc.go
	stopCtx := ctrl.SetupSignalHandler()
	stopCh := stopCtx.Done()

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

	if zfs.ZFSBackupGCEnabled {
		bgc := &BackupGarbageCollector{}
		err = bgc.Initialize(openebsClient, stopCh)
		if err != nil {
			return errors.Wrap(err, "failed to initialize backup garbage collector")
		}
	}
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

func waitForReadySnapshot(ctx context.Context, snapname string) error {
	for {
		snap, err := zfs.GetZFSSnapshot(snapname)
		if err != nil {
			return status.Errorf(codes.Internal,
				"zfs: wait failed, not able to get the snapshot %s %s", snapname, err.Error())
		}

		switch snap.Status.State {
		case zfs.ZFSStatusReady:
			return nil
		case zfs.ZFSStatusFailed:
			return status.Errorf(codes.Internal,
				"snapshot %s creation failed on node %s", snapname, snap.Spec.OwnerNodeID)
		}

		select {
		case <-ctx.Done():
			return status.Errorf(codes.DeadlineExceeded,
				"snapshot %s creation: context deadline reached", snapname)
		case <-time.After(time.Second):
		}
	}
}

// CreateZFSVolume create new zfs volume from csi volume request
func CreateZFSVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*zfsapi.ZFSVolume, error) {
	volName := strings.ToLower(req.GetName())
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)

	// parameter keys may be mistyped from the CRD specification when declaring
	// the storageclass, which kubectl validation will not catch. Because ZFS
	// parameter keys (not values!) are all lowercase, keys may safely be forced
	// to the lower case.
	originalParams := req.GetParameters()
	parameters := helpers.GetCaseInsensitiveMap(&originalParams)

	pvcName := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/pvc/name")
	pvcNamespace := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/pvc/namespace")
	pvName := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/pv/name")

	rs := parameters["recordsize"]
	bs := parameters["volblocksize"]
	compression := parameters["compression"]
	dedup := parameters["dedup"]
	atime := parameters["atime"]
	logbias := parameters["logbias"]
	encr := parameters["encryption"]
	kf := parameters["keyformat"]
	kl := parameters["keylocation"]
	pool := parameters["poolname"]
	poolpattern := parameters["poolpattern"]
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
					return nil, err
				}
			}
		} else {
			if vol.Spec.Capacity != capacity {
				return nil, status.Errorf(codes.AlreadyExists,
					"volume %s already present", volName)
			}
			if vol.Status.State != zfs.ZFSStatusReady {
				return nil, status.Errorf(codes.Aborted,
					"volume %s request already pending", volName)
			}
			return vol, nil
		}
	}

	// the pool parameters are matched as a single regular expression, an exact
	// name being an anchored, quoted one
	pattern, err := compilePoolPattern(pool, poolpattern)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nmap, err := getNodeMap(schld, pattern)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get node map failed : %s", err.Error())
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
		return nil, status.Error(codes.Internal, "scheduler failed, node list is empty for creating the PV")
	}

	reserves := reservesSpace(vtype, tp)

	// resolvePool only picks among the pools which can hold the reservation
	var minFree int64
	if reserves {
		minFree = size
	}

	// only a poolpattern or a reservation needs the pools the agents reported,
	// so a thin volume on a fixed poolname never waits on that report
	var suitable map[string]bool
	if poolpattern != "" || reserves {
		var matched bool
		var serr error

		suitable, matched, serr = getSuitableNodes(pattern, size)
		if serr != nil {
			return nil, status.Errorf(codes.Internal, "get suitable nodes failed : %s", serr.Error())
		}

		// a poolpattern matching no pool anywhere cannot be resolved yet; the
		// pools may still be reported, so this stays retryable, unlike the
		// clone guards. A fixed poolname is used as it is and is not held to it
		if poolpattern != "" && !matched {
			return nil, status.Errorf(codes.FailedPrecondition,
				"no pool matching %s is present on any node, volume %s",
				poolDesc(pool, poolpattern), volName)
		}
	}

	// the scheduler orders the nodes but never drops one, so the nodes which
	// cannot hold the reservation are filtered out of its output here. The fit
	// is best effort: the capacity is the pool root's, ignoring a dataset quota
	// under a poolname, and the ZFSNode capacity is a periodic snapshot, so a
	// create which slips through still fails on the node and CSI retries.
	if reserves {
		if prfList = filterKeep(prfList, suitable); len(prfList) == 0 {
			// the pools are there but full: external-provisioner retries with a
			// backoff and reschedules once a node has room
			return nil, status.Errorf(codes.ResourceExhausted,
				"no pool matching %s has %d bytes free, volume %s",
				poolDesc(pool, poolpattern), size, volName)
		}
	}

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithPVCName(pvcName).
		WithPVCNamespace(pvcNamespace).
		WithPVName(pvName).
		WithCapacity(capacity).
		WithRecordSize(rs).
		WithVolBlockSize(bs).
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
		WithATime(atime).
		WithLogBias(logbias).
		WithCompression(compression).
		Build()

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	klog.Infof("zfs: trying volume creation %s in %s on nodes %v", volName, poolDesc(pool, poolpattern), prfList)

	// try volume creation sequentially on all nodes
	for _, node := range prfList {
		var nodeid string
		nodeid, err = zfs.GetNodeID(node)
		if err != nil {
			klog.Warningf("zfs: volume %s skipping node %s, no node id : %s", volName, node, err.Error())
			continue
		}

		// a poolname is used as it is, since it can carry a dataset path the
		// ZFSNode CR does not advertise. A poolpattern resolves against the
		// pools of the node picked, so the agent finds the pool on it.
		volPool := pool
		if poolpattern != "" {
			if volPool, err = resolvePool(schld, nodeid, pattern, minFree); err != nil {
				klog.Warningf("zfs: volume %s skipping node %s, pool lookup failed : %s", volName, node, err.Error())
				continue
			}
			if volPool == "" {
				// only reachable when the volume was not fit filtered
				err = fmt.Errorf("no pool matching poolpattern %q on node %s", poolpattern, node)
				klog.Warningf("zfs: volume %s skipping node %s : %s", volName, node, err.Error())
				continue
			}
		}

		var vol *zfsapi.ZFSVolume
		vol, err = volbuilder.BuildFrom(volObj).
			WithOwnerNodeID(nodeid).
			WithPoolName(volPool).
			WithVolumeStatus(zfs.ZFSStatusPending).Build()
		if err != nil {
			continue
		}

		klog.Infof("zfs: creating volume %s/%s on node %s", volPool, volName, node)

		timeout := false

		timeout, err = zfs.ProvisionVolume(ctx, vol)
		if err == nil {
			return vol, nil
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

	return nil, status.Errorf(codes.Internal,
		"not able to provision the volume, nodes %v, err : %s", prfList, err.Error())
}

// CreateVolClone creates the clone from a volume
func CreateVolClone(ctx context.Context, req *csi.CreateVolumeRequest, srcVol string) (*zfsapi.ZFSVolume, error) {
	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	pool := helpers.GetInsensitiveParameter(&parameters, "poolname")
	poolpattern := helpers.GetInsensitiveParameter(&parameters, "poolpattern")
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)
	volsize := strconv.FormatInt(int64(size), 10)

	pvcName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/name")
	pvcNamespace := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/namespace")
	pvName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pv/name")

	pattern, err := compilePoolPattern(pool, poolpattern)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	vol, err := zfs.GetZFSVolume(srcVol)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// ZFS cannot clone across pools, so the pool is inherited with the spec
	// below and only checked here
	if !sourcePoolAllowed(vol.Spec.PoolName, pool, pattern) {
		return nil, status.Errorf(codes.InvalidArgument,
			"clone: source pool %s is not covered by %s",
			vol.Spec.PoolName, poolDesc(pool, poolpattern))
	}

	if vol.Spec.Capacity != volsize {
		return nil, status.Error(codes.Internal, "clone: volume size is not matching")
	}

	labels := map[string]string{zfs.ZFSSrcVolKey: vol.Name}

	// create the clone from the source volume

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithPVCName(pvcName).
		WithPVCNamespace(pvcNamespace).
		WithPVName(pvName).
		WithVolumeStatus(zfs.ZFSStatusPending).
		WithLabels(labels).Build()
	if err != nil {
		return nil, err
	}

	// make sure not to override the reserved userprops set by the builder
	userProps := volObj.Spec.UserProperties
	volObj.Spec = vol.Spec
	volObj.Spec.UserProperties = userProps
	// use the snapshot name same as new volname
	volObj.Spec.SnapName = vol.Name + "@" + volName

	_, err = zfs.ProvisionVolume(ctx, volObj)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"clone: not able to provision the volume err : %s", err.Error())
	}

	return volObj, nil
}

// CreateSnapClone creates the clone from a snapshot
func CreateSnapClone(ctx context.Context, req *csi.CreateVolumeRequest, snapshot string) (*zfsapi.ZFSVolume, error) {
	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	pool := helpers.GetInsensitiveParameter(&parameters, "poolname")
	poolpattern := helpers.GetInsensitiveParameter(&parameters, "poolpattern")
	size := getRoundedCapacity(req.GetCapacityRange().RequiredBytes)
	volsize := strconv.FormatInt(int64(size), 10)

	pvcName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/name")
	pvcNamespace := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/namespace")
	pvName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pv/name")

	pattern, err := compilePoolPattern(pool, poolpattern)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	snapshotID := strings.Split(snapshot, "@")
	if len(snapshotID) != 2 {
		return nil, status.Errorf(
			codes.NotFound,
			"snap name is not valid %s, {%s}",
			snapshot,
			"invalid snapshot name",
		)
	}

	snap, err := zfs.GetZFSSnapshot(snapshotID[1])
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// the restore lands in the snapshot's pool, inherited with the spec below
	// and only checked here
	if !sourcePoolAllowed(snap.Spec.PoolName, pool, pattern) {
		return nil, status.Errorf(codes.InvalidArgument,
			"clone: snapshot pool %s is not covered by %s",
			snap.Spec.PoolName, poolDesc(pool, poolpattern))
	}

	if snap.Spec.Capacity != volsize {
		return nil, status.Error(codes.Internal, "clone volume size is not matching")
	}

	volObj, err := volbuilder.NewBuilder().
		WithName(volName).
		WithPVCName(pvcName).
		WithPVCNamespace(pvcNamespace).
		WithPVName(pvName).
		WithVolumeStatus(zfs.ZFSStatusPending).
		Build()
	if err != nil {
		return nil, err
	}

	// make sure not to override the reserved userprops set by the builder
	userProps := volObj.Spec.UserProperties
	volObj.Spec = snap.Spec
	volObj.Spec.UserProperties = userProps
	volObj.Spec.SnapName = strings.ToLower(snapshot)

	_, err = zfs.ProvisionVolume(ctx, volObj)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"not able to provision the clone volume err : %s", err.Error())
	}

	return volObj, nil
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	var err error
	var vol *zfsapi.ZFSVolume

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	volName := strings.ToLower(req.GetName())
	parameters := req.GetParameters()
	// lower case keys, cf CreateZFSVolume()
	size := getRoundedCapacity(req.GetCapacityRange().GetRequiredBytes())
	contentSource := req.GetVolumeContentSource()
	pvcName := helpers.GetInsensitiveParameter(&parameters, "csi.storage.k8s.io/pvc/name")

	unlock := cs.volumeLock.LockVolume(volName)
	defer unlock()

	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID := contentSource.GetSnapshot().GetSnapshotId()

		vol, err = CreateSnapClone(ctx, req, snapshotID)
	} else if contentSource != nil && contentSource.GetVolume() != nil {
		srcVol := contentSource.GetVolume().GetVolumeId()
		vol, err = CreateVolClone(ctx, req, srcVol)
	} else {
		vol, err = CreateZFSVolume(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	klog.Infof("created the volume %s/%s on node %s", vol.Spec.PoolName, volName, vol.Spec.OwnerNodeID)

	sendEventOrIgnore(pvcName, volName, strconv.FormatInt(int64(size), 10), analytics.VolumeProvision)

	topology := map[string]string{zfs.ZFSTopologyKey: vol.Spec.OwnerNodeID}
	cntx := map[string]string{zfs.PoolNameKey: vol.Spec.PoolName, zfs.OpenEBSCasTypeKey: zfs.ZFSCasTypeName}

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
	vol, err := zfs.GetZFSVolume(volumeID)
	if vol != nil && vol.DeletionTimestamp != nil {
		return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
	}

	if err != nil {
		if k8serror.IsNotFound(err) {
			return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
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

	// Fetch the list of snapshot for the given volume
	snapList, err := zfs.GetSnapshotForVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"failed to handle delete volume request for {%s}, "+
				"validation failed checking for snapshots. Error: %s",
			req.VolumeId,
			err.Error(),
		)
	}

	// Delete the corresponding ZV CR only if there are no snapshots present for the volume
	if len(snapList.Items) == 0 {
		err = zfs.DeleteVolume(volumeID)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to handle delete volume request for {%s}",
				volumeID,
			)
		}
	} else {
		// add annotation to the volume to indicate that it is eligible for deletion
		// once all the snapshots are deleted and the reclaim policy is not Retain
		// this volume will be deleted
		err = zfs.MarkForDeletion(volumeID)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to annotate volume on deletion request for {%s}",
				volumeID,
			)
		}

	}

	sendEventOrIgnore("", volumeID, vol.Spec.Capacity, analytics.VolumeDeprovision)

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

	originalParams := req.GetParameters()

	vsName := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/volumesnapshot/name")
	vsNamespace := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/volumesnapshot/namespace")
	vscName := helpers.GetInsensitiveParameter(&originalParams, "csi.storage.k8s.io/volumesnapshotcontent/name")

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
		WithVSName(vsName).
		WithVSNamespace(vsNamespace).
		WithVSCName(vscName).
		WithLabels(labels).Build()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create snapshotobject for %s: %s, {%s}",
			volumeID, snapName,
			err.Error(),
		)
	}
	// make sure not to override the reserved userprops set by the builder
	userProps := snapObj.Spec.UserProperties
	snapObj.Spec = vol.Spec
	snapObj.Spec.UserProperties = userProps
	snapObj.Status.State = zfs.ZFSStatusPending
	if err := zfs.ProvisionSnapshot(snapObj); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s: %s, {%s}",
			volumeID, snapName,
			err.Error(),
		)
	}
	parameters := helpers.GetCaseInsensitiveMap(&originalParams)
	if _, ok := parameters["wait"]; ok {
		if err := waitForReadySnapshot(ctx, snapName); err != nil {
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
	volumeID := snapshotID[0]
	unlock := cs.volumeLock.LockVolumeWithSnapshot(snapshotID[0], snapshotID[1])
	defer unlock()

	// verify if the snapshot has already been deleted
	_, err := zfs.GetZFSSnapshot(snapshotID[1])

	if err != nil {
		if k8serror.IsNotFound(err) {
			return csipayload.NewDeleteSnapshotResponseBuilder().Build(), nil
		}
		return nil, errors.Wrapf(
			err,
			"failed to get snapshot for {%s}",
			snapshotID[1],
		)
	}

	// Fetch the list of snapshot for the given volume
	snapList, err := zfs.GetSnapshotForVolume(volumeID)
	if err != nil {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"failed to handle delete snapshot request for {%s}, "+
				"validation failed checking for snapshot list for volume. Error: %s",
			volumeID,
			err.Error(),
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

	eligibleForDeletion, err := zfs.IsVolumeEligibleForDeletion(volumeID)
	if err != nil {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"failed to handle delete snapshot request for {%s}, "+
				"validation failed checking for eligible for deletion. Error: %s",
			volumeID,
			err.Error(),
		)
	}

	klog.Infof(" The snap list size %v and eligibleForDeletion %v", len(snapList.Items), eligibleForDeletion)
	// Delete the corresponding ZV CR only if this is the last snapshot
	// for the volume and the corresponding pvc is deleted
	if len(snapList.Items) == 1 && eligibleForDeletion {
		err = zfs.DeleteVolume(volumeID)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to handle delete volume request for {%s}",
				volumeID,
			)
		}
		klog.Infof("volume %s deleted after the deletion of last snapshot %s ", volumeID, snapshotID[1])
	}

	return csipayload.NewDeleteSnapshotResponseBuilder().Build(), nil
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

	poolname := helpers.GetInsensitiveParameter(&params, "poolname")
	poolpattern := helpers.GetInsensitiveParameter(&params, "poolpattern")

	// pools are matched by their root, so a "poolname" naming a child dataset
	// reports the whole pool's capacity rather than the dataset's quota
	pattern, err := compilePoolPattern(poolname, poolpattern)
	if err != nil {
		// a storageclass declaring no pool, or both, provisions nothing. It is
		// logged quietly, as this is polled, and CreateVolume is where the
		// misconfiguration is reported as InvalidArgument
		klog.V(4).Infof("GetCapacity: %v", err)
		return &csi.GetCapacityResponse{AvailableCapacity: 0}, nil
	}

	var availableCapacity int64
	for _, nodeName := range nodeNames {
		mappedNodeID, mapErr := zfs.GetNodeID(nodeName)
		if mapErr != nil {
			klog.Warningf("Unable to find mapped node id for %s", nodeName)
			mappedNodeID = nodeName
		}
		v, exists, err := zfsNodesCache.GetByKey(zfs.OpenEBSNamespace + "/" + mappedNodeID)
		if err != nil {
			klog.Warning("unexpected error after querying the zfsNode informer cache")
			continue
		}
		if !exists {
			continue
		}
		zfsNode := v.(*zfsapi.ZFSNode)
		// rather than summing all free capacity, we are calculating maximum
		// zv size that gets fit in given pool, so the node's roomiest matching
		// pool decides: a volume lands in a single pool.
		// See https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1472-storage-capacity-tracking#available-capacity-vs-maximum-volume-size &
		// https://github.com/container-storage-interface/spec/issues/432 for more details
		if _, freeCapacity := maxFreePool(zfsNode.Pools, pattern); availableCapacity < freeCapacity {
			availableCapacity = freeCapacity
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

func (cs *controller) ControllerGetVolume(
	ctx context.Context,
	req *csi.ControllerGetVolumeRequest,
) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume is not implemented")
}

func (cs *controller) ControllerModifyVolume(
	ctx context.Context,
	req *csi.ControllerModifyVolumeRequest,
) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerModifyVolume is not implemented")
}
