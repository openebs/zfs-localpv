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
	"os/exec"

	"github.com/Sirupsen/logrus"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
)

const (
	ZFS_DEVPATH   = "/dev/zvol/"
	ZFSVolCmd     = "zfs"
	ZFSCreateArg  = "create"
	ZFSDestroyArg = "destroy"
	ZFSSetArg     = "set"
)

func PropertyChanged(oldVol *apis.ZFSVolume, newVol *apis.ZFSVolume) bool {
	return oldVol.Spec.Compression != newVol.Spec.Compression ||
		oldVol.Spec.Dedup != newVol.Spec.Dedup ||
		oldVol.Spec.Capacity != newVol.Spec.Capacity
}

// builldVolumeCreateArgs returns zvol create command along with attributes as a string array
func buildVolumeCreateArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolCmd []string

	zvol := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolCmd = append(ZFSVolCmd, ZFSCreateArg)

	if vol.Spec.ThinProvision == "yes" {
		ZFSVolCmd = append(ZFSVolCmd, "-s")
	}
	if len(vol.Spec.Capacity) != 0 {
		ZFSVolCmd = append(ZFSVolCmd, "-V", vol.Spec.Capacity)
	}
	if len(vol.Spec.BlockSize) != 0 {
		ZFSVolCmd = append(ZFSVolCmd, "-b", vol.Spec.BlockSize)
	}
	if len(vol.Spec.Dedup) != 0 {
		dedupProperty := "dedup=" + vol.Spec.Dedup
		ZFSVolCmd = append(ZFSVolCmd, "-o", dedupProperty)
	}
	if len(vol.Spec.Compression) != 0 {
		compressionProperty := "compression=" + vol.Spec.Compression
		ZFSVolCmd = append(ZFSVolCmd, "-o", compressionProperty)
	}
	if len(vol.Spec.Encryption) != 0 {
		encryptionProperty := "encryption=" + vol.Spec.Encryption
		ZFSVolCmd = append(ZFSVolCmd, "-o", encryptionProperty)
	}

	ZFSVolCmd = append(ZFSVolCmd, zvol)

	return ZFSVolCmd
}

// builldVolumeSetArgs returns zvol set command along with attributes as a string array
// TODO(pawan) need to find a way to identify which property has changed
func buildVolumeSetArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolCmd []string

	zvol := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolCmd = append(ZFSVolCmd, ZFSSetArg)

	if len(vol.Spec.Capacity) != 0 {
		volsize := "volsize=" + vol.Spec.Capacity
		ZFSVolCmd = append(ZFSVolCmd, volsize)
	}
	if len(vol.Spec.Dedup) != 0 {
		dedupProperty := "dedup=" + vol.Spec.Dedup
		ZFSVolCmd = append(ZFSVolCmd, dedupProperty)
	}
	if len(vol.Spec.Compression) != 0 {
		compressionProperty := "compression=" + vol.Spec.Compression
		ZFSVolCmd = append(ZFSVolCmd, compressionProperty)
	}

	ZFSVolCmd = append(ZFSVolCmd, zvol)

	return ZFSVolCmd
}

// builldVolumeDestroyArgs returns zvol destroy command along with attributes as a string array
func buildVolumeDestroyArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolCmd []string

	zvol := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolCmd = append(ZFSVolCmd, ZFSDestroyArg, "-R", zvol)

	return ZFSVolCmd
}

// createZvol creates the zvol and returns the corresponding diskPath
// of the volume which gets created on the node
func createZvol(vol *apis.ZFSVolume) (string, error) {
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err := os.Stat(devicePath); os.IsNotExist(err) {

		args := buildVolumeCreateArgs(vol)
		cmd := exec.Command(ZFSVolCmd, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logrus.Errorf(
				"zfs: could not create zvol %v cmd %v error: %s", zvol, args, string(out),
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
	var err error
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err = os.Stat(devicePath); err == nil {
		args := buildVolumeSetArgs(vol)
		cmd := exec.Command(ZFSVolCmd, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logrus.Errorf(
				"zfs: could not set property on zvol %v cmd %v error: %s", zvol, args, string(out),
			)
			return err
		}
		logrus.Infof("property set on zvol %s", zvol)
	}

	return err
}

// DestroyZvol deletes the zvol
func DestroyZvol(vol *apis.ZFSVolume) error {
	zvol := vol.Spec.PoolName + "/" + vol.Name
	devicePath := ZFS_DEVPATH + zvol

	if _, err := os.Stat(devicePath); err == nil {
		args := buildVolumeDestroyArgs(vol)
		cmd := exec.Command(ZFSVolCmd, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logrus.Errorf(
				"zfs: could not destroy zvol %v cmd %v error: %s", zvol, args, string(out),
			)
			return err
		}
		logrus.Infof("destroyed zvol %s", zvol)
	}

	return nil
}
