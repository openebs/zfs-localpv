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
	"github.com/Sirupsen/logrus"
	"os"
	"strconv"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1alpha1"
	"github.com/openebs/zfs-localpv/pkg/builder/snapbuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// OpenEBSNamespace is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespaceKey string = "OPENEBS_NAMESPACE"
	// This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
	// ZFSFinalizer for the ZfsVolume CR
	ZFSFinalizer string = "zfs.openebs.io/finalizer"
	// ZFSVolKey for the ZfsSnapshot CR to store Persistence Volume name
	ZFSVolKey string = "openebs.io/persistent-volume"
	// PoolNameKey is key for ZFS pool name
	PoolNameKey string = "openebs.io/poolname"
	// ZFSNodeKey will be used to insert Label in ZfsVolume CR
	ZFSNodeKey string = "kubernetes.io/nodename"
	// ZFSTopologyKey is supported topology key for the zfs driver
	ZFSTopologyKey string = "kubernetes.io/hostname"
	// ZFSStatusPending shows object has not handled yet
	ZFSStatusPending string = "Pending"
	// ZFSStatusReady shows object has been processed
	ZFSStatusReady string = "Ready"
)

var (
	// OpenEBSNamespace is openebs system namespace
	OpenEBSNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string

	// should send google analytics or not
	GoogleAnalyticsEnabled string
)

func init() {

	OpenEBSNamespace = os.Getenv(OpenEBSNamespaceKey)
	if OpenEBSNamespace == "" {
		logrus.Fatalf("OPENEBS_NAMESPACE environment variable not set")
	}
	NodeID = os.Getenv("OPENEBS_NODE_ID")
	if NodeID == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		logrus.Fatalf("NodeID environment variable not set")
	}

	GoogleAnalyticsEnabled = os.Getenv(GoogleAnalyticsKey)
}

// ProvisionVolume creates a ZFSVolume(zv) CR,
// watcher for zvc is present in CSI agent
func ProvisionVolume(
	vol *apis.ZFSVolume,
) error {

	_, err := volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(vol)
	if err == nil {
		logrus.Infof("provisioned volume %s", vol.Name)
	}

	return err
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
		logrus.Infof("provisioned snapshot %s", snap.Name)
	}

	return err
}

// DeleteSnapshot deletes the corresponding ZFSSnapshot CR
func DeleteSnapshot(snapname string) (err error) {
	err = snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(snapname)
	if err == nil {
		logrus.Infof("deprovisioned snapshot %s", snapname)
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
		logrus.Infof("deprovisioned volume %s", volumeID)
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

// UpdateZvolInfo updates ZFSVolume CR with node id and finalizer
func UpdateZvolInfo(vol *apis.ZFSVolume) error {
	finalizers := []string{ZFSFinalizer}
	labels := map[string]string{ZFSNodeKey: NodeID}

	if vol.Finalizers != nil {
		return nil
	}

	newVol, err := volbuilder.BuildFrom(vol).
		WithFinalizer(finalizers).
		WithLabels(labels).Build()

	if err != nil {
		return err
	}

	_, err = volbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newVol)
	return err
}

// RemoveZvolFinalizer adds finalizer to ZFSVolume CR
func RemoveZvolFinalizer(vol *apis.ZFSVolume) error {
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
		logrus.Errorf("Get snapshot failed %s err: %s", snap.Name, err.Error())
		return "", err
	}

	return snap.Status.State, nil
}

// UpdateSnapInfo updates ZFSSnapshot CR with node id and finalizer
func UpdateSnapInfo(snap *apis.ZFSSnapshot) error {
	finalizers := []string{ZFSFinalizer}
	labels := map[string]string{ZFSNodeKey: NodeID}

	if snap.Finalizers != nil {
		return nil
	}

	newSnap, err := snapbuilder.BuildFrom(snap).
		WithFinalizer(finalizers).
		WithLabels(labels).Build()

	// set the status to ready
	newSnap.Status.State = ZFSStatusReady

	if err != nil {
		logrus.Errorf("Update snapshot failed %s err: %s", snap.Name, err.Error())
		return err
	}

	_, err = snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newSnap)
	return err
}

// RemoveSnapFinalizer adds finalizer to ZFSSnapshot CR
func RemoveSnapFinalizer(snap *apis.ZFSSnapshot) error {
	snap.Finalizers = nil

	_, err := snapbuilder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(snap)
	return err
}
