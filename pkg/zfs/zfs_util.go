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
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
)

// zfs related constants
const (
	ZFS_DEVPATH = "/dev/zvol/"
	ZFS_FSTYPE  = "zfs"
)

// zfs command related constants
const (
	ZFSVolCmd     = "zfs"
	ZFSCreateArg  = "create"
	ZFSDestroyArg = "destroy"
	ZFSSetArg     = "set"
	ZFSMountArg   = "mount"
	ZFSListArg    = "list"
)

// constants to define volume type
const (
	VOLTYPE_DATASET = "DATASET"
	VOLTYPE_ZVOL    = "ZVOL"
)

func PropertyChanged(oldVol *apis.ZFSVolume, newVol *apis.ZFSVolume) bool {
	return oldVol.Spec.Compression != newVol.Spec.Compression ||
		oldVol.Spec.Dedup != newVol.Spec.Dedup
}

// GetVolumeType returns the volume type
// whether it is a zvol or dataset
func GetVolumeType(fstype string) string {
	/*
	 * if fstype is provided as zfs or it is empty then a zfs dataset will be created
	 * otherwise a zvol will be created
	 */
	switch fstype {
	case ZFS_FSTYPE:
		return VOLTYPE_DATASET
	case "":
		return VOLTYPE_DATASET
	default:
		return VOLTYPE_ZVOL
	}
}

// builldZvolCreateArgs returns zfs create command for zvol along with attributes as a string array
func buildZvolCreateArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSCreateArg)

	if vol.Spec.ThinProvision == "yes" {
		ZFSVolArg = append(ZFSVolArg, "-s")
	}
	if len(vol.Spec.Capacity) != 0 {
		ZFSVolArg = append(ZFSVolArg, "-V", vol.Spec.Capacity)
	}
	if len(vol.Spec.RecordSize) != 0 {
		ZFSVolArg = append(ZFSVolArg, "-b", vol.Spec.RecordSize)
	}
	if len(vol.Spec.Dedup) != 0 {
		dedupProperty := "dedup=" + vol.Spec.Dedup
		ZFSVolArg = append(ZFSVolArg, "-o", dedupProperty)
	}
	if len(vol.Spec.Compression) != 0 {
		compressionProperty := "compression=" + vol.Spec.Compression
		ZFSVolArg = append(ZFSVolArg, "-o", compressionProperty)
	}
	if len(vol.Spec.Encryption) != 0 {
		encryptionProperty := "encryption=" + vol.Spec.Encryption
		ZFSVolArg = append(ZFSVolArg, "-o", encryptionProperty)
	}
	if len(vol.Spec.KeyLocation) != 0 {
		keyLocation := "keylocation=" + vol.Spec.KeyLocation
		ZFSVolArg = append(ZFSVolArg, "-o", keyLocation)
	}
	if len(vol.Spec.KeyFormat) != 0 {
		keyFormat := "keyformat=" + vol.Spec.KeyFormat
		ZFSVolArg = append(ZFSVolArg, "-o", keyFormat)
	}

	ZFSVolArg = append(ZFSVolArg, volume)

	return ZFSVolArg
}

// builldDatasetCreateArgs returns zfs create command for dataset along with attributes as a string array
func buildDatasetCreateArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSCreateArg)

	if len(vol.Spec.Capacity) != 0 {
		quotaProperty := "quota=" + vol.Spec.Capacity
		ZFSVolArg = append(ZFSVolArg, "-o", quotaProperty)
	}
	if len(vol.Spec.RecordSize) != 0 {
		recordsizeProperty := "recordsize=" + vol.Spec.RecordSize
		ZFSVolArg = append(ZFSVolArg, "-o", recordsizeProperty)
	}
	if vol.Spec.ThinProvision == "no" {
		reservationProperty := "reservation=" + vol.Spec.Capacity
		ZFSVolArg = append(ZFSVolArg, "-o", reservationProperty)
	}
	if len(vol.Spec.Dedup) != 0 {
		dedupProperty := "dedup=" + vol.Spec.Dedup
		ZFSVolArg = append(ZFSVolArg, "-o", dedupProperty)
	}
	if len(vol.Spec.Compression) != 0 {
		compressionProperty := "compression=" + vol.Spec.Compression
		ZFSVolArg = append(ZFSVolArg, "-o", compressionProperty)
	}
	if len(vol.Spec.Encryption) != 0 {
		encryptionProperty := "encryption=" + vol.Spec.Encryption
		ZFSVolArg = append(ZFSVolArg, "-o", encryptionProperty)
	}
	if len(vol.Spec.KeyLocation) != 0 {
		keyLocation := "keylocation=" + vol.Spec.KeyLocation
		ZFSVolArg = append(ZFSVolArg, "-o", keyLocation)
	}
	if len(vol.Spec.KeyFormat) != 0 {
		keyFormat := "keyformat=" + vol.Spec.KeyFormat
		ZFSVolArg = append(ZFSVolArg, "-o", keyFormat)
	}

	// set the mount path to none, by default zfs mounts it to the default dataset path
	ZFSVolArg = append(ZFSVolArg, "-o", "mountpoint=none", volume)

	return ZFSVolArg
}

// builldVolumeSetArgs returns volume set command along with attributes as a string array
// TODO(pawan) need to find a way to identify which property has changed
func buildVolumeSetArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSSetArg)

	if len(vol.Spec.Dedup) != 0 {
		dedupProperty := "dedup=" + vol.Spec.Dedup
		ZFSVolArg = append(ZFSVolArg, dedupProperty)
	}
	if len(vol.Spec.Compression) != 0 {
		compressionProperty := "compression=" + vol.Spec.Compression
		ZFSVolArg = append(ZFSVolArg, compressionProperty)
	}

	ZFSVolArg = append(ZFSVolArg, volume)

	return ZFSVolArg
}

// builldVolumeDestroyArgs returns volume destroy command along with attributes as a string array
func buildVolumeDestroyArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSDestroyArg, "-R", volume)

	return ZFSVolArg
}

func getVolume(volume string) error {
	var ZFSVolArg []string

	ZFSVolArg = append(ZFSVolArg, ZFSListArg, volume)

	cmd := exec.Command(ZFSVolCmd, ZFSVolArg...)
	out, err := cmd.CombinedOutput()
	logrus.Infof("getVolume out %v", out)
	return err
}

// CreateVolume creates the zvol/dataset as per
// info provided in ZFSVolume object
func CreateVolume(vol *apis.ZFSVolume) error {
	volume := vol.Spec.PoolName + "/" + vol.Name

	if err := getVolume(volume); err != nil {
		var args []string
		if vol.Spec.VolumeType == VOLTYPE_DATASET {
			args = buildDatasetCreateArgs(vol)
		} else {
			args = buildZvolCreateArgs(vol)
		}
		cmd := exec.Command(ZFSVolCmd, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logrus.Errorf(
				"zfs: could not create volume %v cmd %v error: %s", volume, args, string(out),
			)
			return err
		}
		logrus.Infof("created volume %s", volume)
	} else if err == nil {
		logrus.Infof("using existing volume %v", volume)
	} else {
		return err
	}

	return nil
}

// SetDatasetMountProp sets mountpoint for the volume
func SetDatasetMountProp(volume string, mountpath string) error {
	var ZFSVolArg []string

	mountProperty := "mountpoint=" + mountpath
	ZFSVolArg = append(ZFSVolArg, ZFSSetArg, mountProperty, volume)

	cmd := exec.Command(ZFSVolCmd, ZFSVolArg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("zfs: could not set mountpoint on dataset %v cmd %v error: %s",
			volume, ZFSVolArg, string(out))
	}
	return err
}

// MountZFSVolume mounts the volume
func MountZFSVolume(volume string) error {
	var ZFSVolArg []string

	ZFSVolArg = append(ZFSVolArg, ZFSMountArg, volume)
	cmd := exec.Command(ZFSVolCmd, ZFSVolArg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("zfs: could not mount the dataset %v cmd %v error: %s",
			volume, ZFSVolArg, string(out))
	}
	return err
}

// MountZFSDataset mounts the dataset to the given mountpoint
func MountZFSDataset(vol *apis.ZFSVolume, mountpath string) error {
	volume := vol.Spec.PoolName + "/" + vol.Name

	err := SetDatasetMountProp(volume, mountpath)

	if err != nil {
		return err
	}

	return MountZFSVolume(volume)
}

// SetZvolProp sets the volume property
func SetZvolProp(vol *apis.ZFSVolume) error {
	var err error
	volume := vol.Spec.PoolName + "/" + vol.Name

	if len(vol.Spec.Compression) == 0 &&
		len(vol.Spec.Dedup) == 0 {
		//nothing to set, just return
		return nil
	}
	args := buildVolumeSetArgs(vol)
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		logrus.Errorf(
			"zfs: could not set property on volume %v cmd %v error: %s", volume, args, string(out),
		)
		return err
	}
	logrus.Infof("property set on volume %s", volume)

	return err
}

// DestroyVolume deletes the zfs volume
func DestroyVolume(vol *apis.ZFSVolume) error {
	volume := vol.Spec.PoolName + "/" + vol.Name

	if err := getVolume(volume); err != nil {
		return nil
	}

	args := buildVolumeDestroyArgs(vol)
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		logrus.Errorf(
			"zfs: could not destroy volume %v cmd %v error: %s", volume, args, string(out),
		)
		return err
	}
	logrus.Infof("destroyed volume %s", volume)

	return nil
}

// GetVolumeDevPath returns devpath for the given volume
func GetVolumeDevPath(vol *apis.ZFSVolume) (string, error) {
	volume := vol.Spec.PoolName + "/" + vol.Name
	if vol.Spec.VolumeType == VOLTYPE_DATASET {
		return volume, nil
	}

	devicePath := ZFS_DEVPATH + volume

	// evaluate the symlink to get the dev path for zvol
	dev, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return "", err
	}

	return dev, nil
}
