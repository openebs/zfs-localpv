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
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1alpha1"
)

// zfs related constants
const (
	ZFS_DEVPATH = "/dev/zvol/"
	FSTYPE_ZFS  = "zfs"
)

// zfs command related constants
const (
	ZFSVolCmd      = "zfs"
	ZFSCreateArg   = "create"
	ZFSCloneArg    = "clone"
	ZFSDestroyArg  = "destroy"
	ZFSSetArg      = "set"
	ZFSGetArg      = "get"
	ZFSListArg     = "list"
	ZFSSnapshotArg = "snapshot"
)

// constants to define volume type
const (
	VOLTYPE_DATASET = "DATASET"
	VOLTYPE_ZVOL    = "ZVOL"
)

func PropertyChanged(oldVol *apis.ZFSVolume, newVol *apis.ZFSVolume) bool {
	if oldVol.Spec.VolumeType == VOLTYPE_DATASET &&
		newVol.Spec.VolumeType == VOLTYPE_DATASET &&
		oldVol.Spec.RecordSize != newVol.Spec.RecordSize {
		return true
	}

	return oldVol.Spec.Compression != newVol.Spec.Compression ||
		oldVol.Spec.Dedup != newVol.Spec.Dedup
}

// GetVolumeType returns the volume type
// whether it is a zvol or dataset
func GetVolumeType(fstype string) string {
	/*
	 * if fstype is provided as zfs then a zfs dataset will be created
	 * otherwise a zvol will be created
	 */
	switch fstype {
	case FSTYPE_ZFS:
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
	if len(vol.Spec.VolBlockSize) != 0 {
		ZFSVolArg = append(ZFSVolArg, "-b", vol.Spec.VolBlockSize)
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

// builldCloneCreateArgs returns zfs clone commands for zfs volume/dataset along with attributes as a string array
func buildCloneCreateArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name
	snapshot := vol.Spec.PoolName + "/" + vol.Spec.SnapName

	ZFSVolArg = append(ZFSVolArg, ZFSCloneArg)

	if vol.Spec.VolumeType == VOLTYPE_DATASET {
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
		ZFSVolArg = append(ZFSVolArg, "-o", "mountpoint=none")
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
	ZFSVolArg = append(ZFSVolArg, snapshot, volume)
	return ZFSVolArg
}

// buildZFSSnapCreateArgs returns zfs create command for zfs snapshot
// zfs snapshot <poolname>/<volname>@<snapname>
func buildZFSSnapCreateArgs(snap *apis.ZFSSnapshot) []string {
	var ZFSSnapArg []string

	volname := snap.Labels[ZFSVolKey]
	snapDataset := snap.Spec.PoolName + "/" + volname + "@" + snap.Name

	ZFSSnapArg = append(ZFSSnapArg, ZFSSnapshotArg, snapDataset)

	return ZFSSnapArg
}

// builldZFSSnapDestroyArgs returns zfs destroy command for zfs snapshot
// zfs destroy <poolname>/<volname>@<snapname>
func buildZFSSnapDestroyArgs(snap *apis.ZFSSnapshot) []string {
	var ZFSSnapArg []string

	volname := snap.Labels[ZFSVolKey]
	snapDataset := snap.Spec.PoolName + "/" + volname + "@" + snap.Name

	ZFSSnapArg = append(ZFSSnapArg, ZFSDestroyArg, snapDataset)

	return ZFSSnapArg
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

	if vol.Spec.VolumeType == VOLTYPE_DATASET &&
		len(vol.Spec.RecordSize) != 0 {
		recordsizeProperty := "recordsize=" + vol.Spec.RecordSize
		ZFSVolArg = append(ZFSVolArg, recordsizeProperty)
	}

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

// builldVolumeResizeArgs returns volume set  for resizing the zfs volume
func buildVolumeResizeArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSSetArg)

	if vol.Spec.VolumeType == VOLTYPE_DATASET {
		quotaProperty := "quota=" + vol.Spec.Capacity
		ZFSVolArg = append(ZFSVolArg, quotaProperty)
	} else {
		volsizeProperty := "volsize=" + vol.Spec.Capacity
		ZFSVolArg = append(ZFSVolArg, volsizeProperty)
	}

	ZFSVolArg = append(ZFSVolArg, volume)

	return ZFSVolArg
}

// builldVolumeDestroyArgs returns volume destroy command along with attributes as a string array
func buildVolumeDestroyArgs(vol *apis.ZFSVolume) []string {
	var ZFSVolArg []string

	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSDestroyArg, volume)

	return ZFSVolArg
}

func getVolume(volume string) error {
	var ZFSVolArg []string

	ZFSVolArg = append(ZFSVolArg, ZFSListArg, volume)

	cmd := exec.Command(ZFSVolCmd, ZFSVolArg...)
	_, err := cmd.CombinedOutput()
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
	}

	return nil
}

// CreateClone creates clone for the zvol/dataset as per
// info provided in ZFSVolume object
func CreateClone(vol *apis.ZFSVolume) error {
	volume := vol.Spec.PoolName + "/" + vol.Name

	if err := getVolume(volume); err != nil {
		var args []string
		args = buildCloneCreateArgs(vol)
		cmd := exec.Command(ZFSVolCmd, args...)
		out, err := cmd.CombinedOutput()

		if err != nil {
			logrus.Errorf(
				"zfs: could not clone volume %v cmd %v error: %s", volume, args, string(out),
			)
			return err
		}
		logrus.Infof("created clone %s", volume)
	} else if err == nil {
		logrus.Infof("using existing clone volume %v", volume)
	}

	if vol.Spec.FsType == "xfs" {
		return xfs_generate_uuid(volume)
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

// MountZFSDataset mounts the dataset to the given mountpoint
func MountZFSDataset(vol *apis.ZFSVolume, mountpath string) error {
	volume := vol.Spec.PoolName + "/" + vol.Name

	// set the mountpoint to the path where this volume should be mounted
	err := SetDatasetMountProp(volume, mountpath)
	if err != nil {
		return err
	}

	/*
	 * see if we should attempt to mount the dataset.
	 * Setting the mountpoint is sufficient to mount the zfs dataset,
	 * but if dataset has been unmounted, then setting the mountpoint
	 * is not sufficient, we have to mount the dataset explicitly
	 */
	mounted, err := GetVolumeProperty(vol, "mounted")
	if err != nil {
		return err
	}

	if mounted == "no" {
		var MountVolArg []string
		MountVolArg = append(MountVolArg, "mount", volume)
		cmd := exec.Command(ZFSVolCmd, MountVolArg...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Errorf("zfs: could not mount the dataset %v cmd %v error: %s",
				volume, MountVolArg, string(out))
			return err
		}
	}

	return nil
}

// UmountZFSDataset umounts the dataset
func UmountZFSDataset(vol *apis.ZFSVolume) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	var MountVolArg []string
	MountVolArg = append(MountVolArg, "umount", volume)
	cmd := exec.Command(ZFSVolCmd, MountVolArg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("zfs: could not umount the dataset %v cmd %v error: %s",
			volume, MountVolArg, string(out))
		return err
	}
	// ignoring the failure of setting the mountpoint to none
	// as the dataset has already been umounted, now the new pod
	// can mount it and it will change that to desired mountpath
	// and try to mount it if not mounted
	SetDatasetMountProp(volume, "none")

	return nil
}

// GetVolumeProperty gets zfs properties for the volume
func GetVolumeProperty(vol *apis.ZFSVolume, prop string) (string, error) {
	var ZFSVolArg []string
	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, ZFSGetArg, "-pH", "-o", "value", prop, volume)

	cmd := exec.Command(ZFSVolCmd, ZFSVolArg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("zfs: could not get %s on dataset %v cmd %v error: %s",
			prop, volume, ZFSVolArg, string(out))
		return "", err
	}
	val := out[:len(out)-1]
	return string(val), nil
}

// SetVolumeProp sets the volume property
func SetVolumeProp(vol *apis.ZFSVolume) error {
	var err error
	volume := vol.Spec.PoolName + "/" + vol.Name

	if len(vol.Spec.Compression) == 0 &&
		len(vol.Spec.Dedup) == 0 &&
		(vol.Spec.VolumeType != VOLTYPE_DATASET ||
			len(vol.Spec.RecordSize) == 0) {
		//nothing to set, just return
		return nil
	}
	/* Case: Restart =>
	 * In this case we get the add event but here we don't know which
	 * property has changed when we were down, so firing the zfs set
	 * command with the all property present on the ZFSVolume.

	 * Case: Property Change =>
	 * TODO(pawan) When we get the update event, we make sure at least
	 * one property has changed before adding it to the event queue for
	 * handling. At this stage, since we haven't stored the
	 * ZFSVolume object as it will be too heavy, we are firing the set
	 * command with the all property preset in the ZFSVolume object since
	 * it is guaranteed that at least one property has changed.
	 */

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
		logrus.Errorf(
			"destroy: volume %v is not present, error: %s", volume, err.Error(),
		)
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

// CreateSnapshot creates the zfs volume snapshot
func CreateSnapshot(snap *apis.ZFSSnapshot) error {

	volume := snap.Labels[ZFSVolKey]
	snapDataset := snap.Spec.PoolName + "/" + volume + "@" + snap.Name

	if err := getVolume(snapDataset); err == nil {
		logrus.Infof("snapshot already there %s", snapDataset)
		// snapshot already there just return
		return nil
	}

	args := buildZFSSnapCreateArgs(snap)
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		logrus.Errorf(
			"zfs: could not create snapshot %v@%v cmd %v error: %s", volume, snap.Name, args, string(out),
		)
		return err
	}
	logrus.Infof("created snapshot %s@%s", volume, snap.Name)
	return nil
}

// DestroySnapshot deletes the zfs volume snapshot
func DestroySnapshot(snap *apis.ZFSSnapshot) error {

	volume := snap.Labels[ZFSVolKey]
	snapDataset := snap.Spec.PoolName + "/" + volume + "@" + snap.Name

	if err := getVolume(snapDataset); err != nil {
		logrus.Errorf(
			"destroy: snapshot %v is not present, error: %s", volume, err.Error(),
		)
		return nil
	}

	args := buildZFSSnapDestroyArgs(snap)
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		logrus.Errorf(
			"zfs: could not destroy snapshot %v@%v cmd %v error: %s", volume, snap.Name, args, string(out),
		)
		return err
	}
	logrus.Infof("deleted snapshot %s@%s", volume, snap.Name)
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

func ResizeZFSVolume(vol *apis.ZFSVolume, mountpath string) error {

	volume := vol.Spec.PoolName + "/" + vol.Name
	args := buildVolumeResizeArgs(vol)
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		logrus.Errorf(
			"zfs: could not resize the volume %v cmd %v error: %s", volume, args, string(out),
		)
		return err
	}

	err = handleVolResize(vol, mountpath)
	return err
}
