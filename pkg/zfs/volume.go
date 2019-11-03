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

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
	"github.com/openebs/zfs-localpv/pkg/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// OpenEBSNamespace is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespaceKey string = "OPENEBS_NAMESPACE"
	// ZFSFinalizer for the ZfsVolume CR
	ZFSFinalizer string = "zfs.openebs.io/finalizer"
	// ZFSNodeKey will be used to insert Label in ZfsVolume CR
	ZFSNodeKey string = "kubernetes.io/nodename"
	// ZFSTopologyKey is supported topology key for the zfs driver
	ZFSTopologyKey string = "kubernetes.io/hostname"
)

var (
	// OpenEBSNamespace is openebs system namespace
	OpenEBSNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string
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
}

// ProvisionVolume creates a ZFSVolume(zv) CR,
// watcher for zvc is present in CSI agent
func ProvisionVolume(
	size int64,
	vol *apis.ZFSVolume,
) error {

	_, err := builder.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(vol)
	if err == nil {
		logrus.Infof("provisioned volume %s", vol.Name)
	}

	return err
}

// GetVolume the corresponding ZFSVolume CR
func GetVolume(volumeID string) (*apis.ZFSVolume, error) {
	return builder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
}

// DeleteVolume deletes the corresponding ZFSVol CR
func DeleteVolume(volumeID string) (err error) {
	err = builder.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(volumeID)
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

	return builder.NewKubeclient().
		WithNamespace(OpenEBSNamespace).List(listOptions)

}

// GetZFSVolume fetches the current Published csi Volume
func GetZFSVolume(volumeID string) (*apis.ZFSVolume, error) {
	getOptions := metav1.GetOptions{}
	vol, err := builder.NewKubeclient().
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

	newVol, err := builder.BuildFrom(vol).
		WithFinalizer(finalizers).
		WithLabels(labels).Build()

	if err != nil {
		return err
	}

	_, err = builder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(newVol)
	return err
}

// RemoveZvolFinalizer adds finalizer to ZFSVolume CR
func RemoveZvolFinalizer(vol *apis.ZFSVolume) error {
	vol.Finalizers = nil

	_, err := builder.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(vol)
	return err
}
