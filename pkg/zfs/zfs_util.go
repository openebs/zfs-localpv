/*
Copyright 2017 The Kubernetes Authors.

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

package zfs

import (
	"os"

	"github.com/Sirupsen/logrus"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
	"k8s.io/kubernetes/pkg/util/mount"
)

const (
	ZFS_DEVPATH = "/dev/zvol/"
)

func PropertyChanged(oldVol *apis.ZFSVolume, newVol *apis.ZFSVolume) bool {
	return oldVol.Spec.Compression != newVol.Spec.Compression ||
		oldVol.Spec.Dedup != newVol.Spec.Dedup ||
		oldVol.Spec.Capacity != newVol.Spec.Capacity
}

// createZvol creates the zvol and returns the corresponding diskPath
// of the volume which gets created on the node
func createZvol(vol *apis.ZFSVolume) (string, error) {
	var out []byte
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		if vol.Spec.ThinProvision == "yes" {
			out, err = mount.NewOsExec().Run(
				"zfs", "create",
				"-s",
				"-V", vol.Spec.Capacity,
				"-b", vol.Spec.BlockSize,
				"-o", "compression="+vol.Spec.Compression,
				"-o", "dedup="+vol.Spec.Dedup,
				zvol,
			)
		} else {
			out, err = mount.NewOsExec().Run(
				"zfs", "create",
				"-V", vol.Spec.Capacity,
				"-b", vol.Spec.BlockSize,
				"-o", "compression="+vol.Spec.Compression,
				"-o", "dedup="+vol.Spec.Dedup,
				zvol,
			)
		}

		if err != nil {
			logrus.Errorf(
				"zfs: could not create zvol %v vol %v error: %s", zvol, vol, string(out),
			)
			return "", err
		}
		logrus.Infof("created zvol %s", zvol)
	} else if err == nil {
		logrus.Infof("using existing zvol %v", zvol)
	} else {
		return "", err
	}

	return devicePath, nil
}

// SetZvolProp sets the zvol property
func SetZvolProp(vol *apis.ZFSVolume) error {
	var out []byte
	var err error
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err = os.Stat(devicePath); err == nil {
		// TODO(pawan) need to find a way to identify
		// which property has changed
		out, err = mount.NewOsExec().Run(
			"zfs", "set",
			"volsize="+vol.Spec.Capacity,
			"compression="+vol.Spec.Compression,
			"dedup="+vol.Spec.Dedup,
			zvol,
		)
		if err != nil {
			logrus.Errorf(
				"zfs: could not set property on zvol %v vol %v error: %s", zvol, vol, string(out),
			)
			return err
		}
		logrus.Infof("property set on zvol %s", zvol)
	}

	return err
}

// DestroyZvol deletes the zvol
func DestroyZvol(vol *apis.ZFSVolume) error {
	var out []byte
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err := os.Stat(devicePath); err == nil {
		out, err = mount.NewOsExec().Run(
			"zfs", "destroy",
			"-R",
			zvol,
		)
		if err != nil {
			logrus.Errorf(
				"zfs: could not destroy zvol %v vol %v error: %s", zvol, vol, string(out),
			)
			return err
		}
		logrus.Infof("destroyed zvol %s", zvol)
	}

	return nil
}
