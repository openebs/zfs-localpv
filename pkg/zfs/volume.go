// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zfs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/bkpbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/restorebuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	// OpenEBSNamespaceKey is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespaceKey string = "OPENEBS_NAMESPACE"
	// GoogleAnalyticsKey This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
	// ZFSFinalizer for the ZfsVolume CR
	ZFSFinalizer string = "zfs.openebs.io/finalizer"
	// ZFSVolKey for the ZfsSnapshot CR to store Persistence Volume name
	ZFSVolKey string = "openebs.io/persistent-volume"
	// ZFSSrcVolKey key for the source Volume name
	ZFSSrcVolKey string = "openebs.io/source-volume"
	// PoolNameKey is key for ZFS pool name
	PoolNameKey string = "openebs.io/poolname"
	// ZFSNodeKey will be used to insert Label in ZfsVolume CR
	ZFSNodeKey string = "kubernetes.io/nodename"
	// ZFSTopologyKey is supported topology key for the zfs driver
	ZFSTopologyKey string = "openebs.io/nodename"
	// ZFSStatusPending shows object has not handled yet
	ZFSStatusPending string = "Pending"
	// ZFSStatusFailed shows object operation has failed
	ZFSStatusFailed string = "Failed"
	// ZFSStatusReady shows object has been processed
	ZFSStatusReady string = "Ready"
)

var (
	// OpenEBSNamespace is openebs system namespace
	OpenEBSNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string

	// ZFSAffinityKey is the key for setting the node affinity on the PV
	ZFSAffinityKey string

	// GoogleAnalyticsEnabled should send google analytics or not
	GoogleAnalyticsEnabled string
)

func init() {
	var err error

	OpenEBSNamespace = os.Getenv(OpenEBSNamespaceKey)
	ZFSAffinityKey = os.Getenv("NODE_AFFINITY_KEY")

	if os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		if OpenEBSNamespace == "" {
			klog.Fatalf("OPENEBS_NAMESPACE environment variable not set for daemonset")
		}
		nodename := os.Getenv("OPENEBS_NODE_NAME")
		if nodename == "" {
			klog.Fatalf("OPENEBS_NODE_NAME environment variable not set")
		}
		if len(ZFSAffinityKey) > 0 {
			// if affinity key is provided, the node should be labelled with that key
			if NodeID, err = GetNodeID(nodename); err != nil {
				klog.Fatalf("GetNodeID failed for node=%s key=%s, err: %s", nodename, ZFSAffinityKey, err.Error())
			}
		} else {
			// if key is not provided use the Driver's topology key and value
			ZFSAffinityKey = ZFSTopologyKey
			NodeID = nodename
		}
		klog.Infof("zfs: node(%s) affinity key=%s nodeid=%s", nodename, ZFSAffinityKey, NodeID)
	} else if os.Getenv("OPENEBS_CONTROLLER_DRIVER") != "" {
		if OpenEBSNamespace == "" {
			klog.Fatalf("OPENEBS_NAMESPACE environment variable not set for controller")
		}

		if ZFSAffinityKey == "" {
			ZFSAffinityKey = ZFSTopologyKey
		}
		klog.Infof("zfs: controller will use affinity key=%s", ZFSAffinityKey)
	}

	GoogleAnalyticsEnabled = os.Getenv(GoogleAnalyticsKey)
}

func GetNodeID(nodename string) (string, error) {
	node, err := k8sapi.GetNode(nodename)
	if err != nil {
		return "", fmt.Errorf("failed to get the node %s", nodename)
	}

	nodeid, ok := node.Labels[ZFSAffinityKey]
	if !ok {
		return "", fmt.Errorf("node %s is not labelled with the key %s", nodename, ZFSAffinityKey)
	}
	return nodeid, nil
}

func checkVolCreation(ctx context.Context, volname string) (bool, error) {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return true, fmt.Errorf("zfs: context deadline reached")
		case <-timeout:
			return true, fmt.Errorf("zfs: vol creation timeout reached")
		default:
			vol, err := GetZFSVolume(volname)
			if err != nil {
				return false, fmt.Errorf("zfs: wait failed, not able to get the volume %s %s", volname, err.Error())
			}

			switch vol.Status.State {
			case ZFSStatusReady:
				return false, nil
			case ZFSStatusFailed:
				return false, fmt.Errorf("zfs: volume creation failed")
			}

			klog.Infof("zfs: waiting for volume %s/%s to be created on nodeid %s",
				vol.Spec.PoolName, volname, vol.Spec.OwnerNodeID)

			time.Sleep(time.Second)
		}
	}
}

// ProvisionVolume creates a ZFSVolume(zv) CR,
// watcher for zvc is present in CSI agent
func ProvisionVolume(
	ctx context.Context,
	vol *apis.ZFSVolume,
) (bool, error) {
	timeout := false
	zv, err := GetZFSVolume(vol.Name)

	if err == nil {
		// update the spec and status
		zv.Spec = vol.Spec
		zv.Status = vol.Status
		_, err = volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(zv)
	} else {
		_, err = volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(vol)
	}

	if err == nil {
		timeout, err = checkVolCreation(ctx, vol.Name)
	}

	if err != nil {
		klog.Infof("zfs: volume %s/%s provisioning failed on node %s err: %s",
			vol.Spec.PoolName, vol.Name, vol.Spec.OwnerNodeID, err.Error())
	}

	return timeout, err
}

// ResizeVolume resizes the zfs volume
func ResizeVolume(vol *apis.ZFSVolume, newSize int64) error {

	vol.Spec.Capacity = strconv.FormatInt(int64(newSize), 10)

	_, err := volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(vol)
	return err
}

// ProvisionSnapshot creates a ZFSSnapshot CR,
// watcher for zvc is present in CSI agent
func ProvisionSnapshot(
	snap *apis.ZFSSnapshot,
) error {

	_, err := snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(snap)
	if err == nil {
		klog.Infof("provisioned snapshot %s", snap.Name)
	}

	return err
}

// DeleteSnapshot deletes the corresponding ZFSSnapshot CR
func DeleteSnapshot(snapname string) (err error) {
	err = snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(snapname)
	if err == nil {
		klog.Infof("deprovisioned snapshot %s", snapname)
	}

	return
}

// GetVolume the corresponding ZFSVolume CR
func GetVolume(volumeID string) (*apis.ZFSVolume, error) {
	return volbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
}

// DeleteVolume deletes the corresponding ZFSVol CR
func DeleteVolume(volumeID string) (err error) {
	err = volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(volumeID)
	if err == nil {
		klog.Infof("zfs: deleted the volume %s", volumeID)
	} else {
		klog.Infof("zfs: volume %s deletion failed %s", volumeID, err.Error())
	}

	return
}

// GetVolList fetches the current Published Volume list
func GetVolList(volumeID string) (*apis.ZFSVolumeList, error) {
	listOptions := v1.ListOptions{
		LabelSelector: ZFSNodeKey + "=" + NodeID,
	}

	return volbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).List(listOptions)

}

// GetZFSVolume fetches the given ZFSVolume
func GetZFSVolume(volumeID string) (*apis.ZFSVolume, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(volumeID, getOptions)
	return vol, err
}

// GetZFSVolumeState returns ZFSVolume OwnerNode and State for
// the given volume. CreateVolume request may call it again and
// again until volume is "Ready".
func GetZFSVolumeState(volID string) (string, string, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(volID, getOptions)

	if err != nil {
		return "", "", err
	}

	return vol.Spec.OwnerNodeID, vol.Status.State, nil
}

// UpdateZvolInfo updates ZFSVolume CR with node id and finalizer
func UpdateZvolInfo(vol *apis.ZFSVolume, status string) error {
	finalizers := []string{}
	labels := map[string]string{ZFSNodeKey: NodeID}

	switch status {
	case ZFSStatusReady:
		finalizers = append(finalizers, ZFSFinalizer)
	}

	newVol, err := volbuilder.BuildFrom(vol).
		WithFinalizer(finalizers).
		WithVolumeStatus(status).
		WithLabels(labels).Build()

	if err != nil {
		return err
	}

	_, err = volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newVol)
	return err
}

// RemoveVolumeFinalizer removes finalizer from ZFSVolume CR
func RemoveVolumeFinalizer(vol *apis.ZFSVolume) error {
	vol.Finalizers = nil

	_, err := volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(vol)
	return err
}

// GetZFSSnapshot fetches the given ZFSSnapshot
func GetZFSSnapshot(snapID string) (*apis.ZFSSnapshot, error) {
	getOptions := metav1.GetOptions{}
	snap, err := snapbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(snapID, getOptions)
	return snap, err
}

// GetZFSSnapshotStatus returns ZFSSnapshot status
func GetZFSSnapshotStatus(snapID string) (string, error) {
	getOptions := metav1.GetOptions{}
	snap, err := snapbuilder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(snapID, getOptions)

	if err != nil {
		klog.Errorf("Get snapshot failed %s err: %s", snap.Name, err.Error())
		return "", err
	}

	return snap.Status.State, nil
}

// UpdateSnapInfo updates ZFSSnapshot CR with node id and finalizer
func UpdateSnapInfo(snap *apis.ZFSSnapshot) error {
	finalizers := []string{ZFSFinalizer}
	labels := map[string]string{ZFSNodeKey: NodeID}

	newSnap, err := snapbuilder.BuildFrom(snap).
		WithFinalizer(finalizers).
		WithLabels(labels).Build()

	// set the status to ready
	newSnap.Status.State = ZFSStatusReady

	if err != nil {
		klog.Errorf("Update snapshot failed %s err: %s", snap.Name, err.Error())
		return err
	}

	_, err = snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newSnap)
	return err
}

// RemoveSnapFinalizer removes finalizer from ZFSSnapshot CR
func RemoveSnapFinalizer(snap *apis.ZFSSnapshot) error {
	snap.Finalizers = nil

	_, err := snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(snap)
	return err
}

// RemoveBkpFinalizer removes finalizer from ZFSBackup CR
func RemoveBkpFinalizer(bkp *apis.ZFSBackup) error {
	bkp.Finalizers = nil

	_, err := bkpbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(bkp)
	return err
}

// UpdateBkpInfo updates the backup info with the status
func UpdateBkpInfo(bkp *apis.ZFSBackup, status apis.ZFSBackupStatus) error {
	finalizers := []string{ZFSFinalizer}
	newBkp, err := bkpbuilder.BuildFrom(bkp).WithFinalizer(finalizers).Build()

	// set the status
	newBkp.Status = status

	if err != nil {
		klog.Errorf("Update backup failed %s err: %s", bkp.Spec.VolumeName, err.Error())
		return err
	}

	_, err = bkpbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newBkp)
	return err
}

// UpdateRestoreInfo updates the rstr info with the status
func UpdateRestoreInfo(rstr *apis.ZFSRestore, status apis.ZFSRestoreStatus) error {
	newRstr, err := restorebuilder.BuildFrom(rstr).Build()

	// set the status
	newRstr.Status = status

	if err != nil {
		klog.Errorf("Update snapshot failed %s err: %s", rstr.Spec.VolumeName, err.Error())
		return err
	}

	_, err = restorebuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newRstr)
	return err
}

// GetUserFinalizers returns all the finalizers present on the ZFSVolume object
// execpt the one owned by ZFS node daemonset. We also need to ignore the foregroundDeletion
// finalizer as this will be present becasue of the foreground cascading deletion
func GetUserFinalizers(finalizers []string) []string {
	var userFin []string
	for _, fin := range finalizers {
		if fin != ZFSFinalizer &&
			fin != "foregroundDeletion" {
			userFin = append(userFin, fin)
		}
	}
	return userFin
}

// IsVolumeReady returns true if volume is Ready
func IsVolumeReady(vol *apis.ZFSVolume) bool {

	if vol.Status.State == ZFSStatusReady {
		return true
	}

	// For older volumes, there was no Status field
	// so checking the node finalizer to make sure volume is Ready
	for _, fin := range vol.Finalizers {
		if fin == ZFSFinalizer {
			return true
		}
	}
	return false
}
