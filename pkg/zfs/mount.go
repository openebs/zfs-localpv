/*
Copyright © 2020 The OpenEBS Authors
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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	mnt "github.com/openebs/lib-csi/pkg/mount"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

// MountInfo contains the volume related info
// for all types of volumes in ZFSVolume
type MountInfo struct {
	// FSType of a volume will specify the
	// format type - ext4(default), xfs of PV
	FSType string `json:"fsType"`

	// AccessMode of a volume will hold the
	// access mode of the volume
	AccessModes []string `json:"accessModes"`

	// MountPath of the volume will hold the
	// path on which the volume is mounted
	// on that node
	MountPath string `json:"mountPath"`

	// MountOptions specifies the options with
	// which mount needs to be attempted
	MountOptions []string `json:"mountOptions"`
}

// FormatAndMountZvol formats and mounts the created volume to the desired mount path
func FormatAndMountZvol(devicePath string, mountInfo *MountInfo) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}

	err := mounter.FormatAndMount(devicePath, mountInfo.MountPath, mountInfo.FSType, mountInfo.MountOptions)
	if err != nil {
		klog.Errorf(
			"zfspv: failed to mount volume %s [%s] to %s, error %v",
			devicePath, mountInfo.FSType, mountInfo.MountPath, err,
		)
		return err
	}

	return nil
}

// UmountVolume unmounts the volume and the corresponding mount path is removed
func UmountVolume(vol *apis.ZFSVolume, targetPath string,
) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}

	dev, ref, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		klog.Errorf(
			"zfspv umount volume: failed to get device from mnt: %s\nError: %v",
			targetPath, err,
		)
		return err
	}

	// device has already been un-mounted, return successful
	if len(dev) == 0 || ref == 0 {
		klog.Warningf(
			"Warning: Unmount skipped because volume %s not mounted: %v",
			vol.Name, targetPath,
		)
		return nil
	}

	if pathExists, pathErr := mount.PathExists(targetPath); pathErr != nil {
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		klog.Warningf(
			"Warning: Unmount skipped because path does not exist: %v",
			targetPath,
		)
		return nil
	}

	if err = mounter.Unmount(targetPath); err != nil {
		klog.Errorf(
			"zfs: failed to unmount %s: path %s err: %v",
			vol.Name, targetPath, err,
		)
		return err
	}

	if err = SetDatasetLegacyMount(vol); err != nil {
		// ignoring the failure as the volume has already
		// been umounted, now the new pod can mount it
		klog.Warningf(
			"zfs: failed to set legacy mountpoint: %s err: %v",
			vol.Name, err,
		)
	}

	if err := os.Remove(targetPath); err != nil {
		klog.Errorf("zfspv: failed to remove mount path vol %s err : %v", vol.Name, err)
	}

	klog.Infof("umount done %s path %v", vol.Name, targetPath)

	return nil
}

func verifyMountRequest(vol *apis.ZFSVolume, mountpath string) (bool, error) {
	if len(mountpath) == 0 {
		return false, status.Error(codes.InvalidArgument, "verifyMount: mount path missing in request")
	}

	if len(vol.Spec.OwnerNodeID) > 0 &&
		vol.Spec.OwnerNodeID != NodeID {
		return false, status.Error(codes.Internal, "verifyMount: volume is owned by different node")
	}
	if !IsVolumeReady(vol) {
		return false, status.Error(codes.Internal, "verifyMount: volume is not ready to be mounted")
	}

	devicePath, err := GetVolumeDevPath(vol)
	if err != nil {
		klog.Errorf("can not get device for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return false, status.Errorf(codes.Internal, "verifyMount: GetVolumePath failed %s", err.Error())
	}

	/*
	 * This check is the famous *Wall Of North*
	 * It will not let the volume to be mounted
	 * at more than two places. The volume should
	 * be unmounted before proceeding to the mount
	 * operation.
	 */
	currentMounts, err := mnt.GetMounts(devicePath)
	if err != nil {
		klog.Errorf("can not get mounts for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return false, status.Errorf(codes.Internal, "verifyMount: Getmounts failed %s", err.Error())
	} else if len(currentMounts) >= 1 {
		// if device is already mounted at the mount point, return successful
		for _, mp := range currentMounts {
			if mp == mountpath {
				return true, nil
			}
		}

		// if it is not a shared volume, then it should not mounted to more than one path
		if vol.Spec.Shared != "yes" {
			klog.Errorf(
				"can not mount, volume:%s already mounted dev %s mounts: %v",
				vol.Name, devicePath, currentMounts,
			)
			return false, status.Errorf(codes.Internal, "verifyMount: device already mounted at %s", currentMounts)
		}
	}
	return false, nil
}

// MountZvol mounts the disk to the specified path
func MountZvol(vol *apis.ZFSVolume, mount *MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	mounted, err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return err
	}

	if mounted {
		klog.Infof("zvol : already mounted %s => %s", volume, mount.MountPath)
		return nil
	}

	devicePath := ZFSDevPath + volume

	err = FormatAndMountZvol(devicePath, mount)
	if err != nil {
		return status.Error(codes.Internal, "not able to format and mount the zvol")
	}

	klog.Infof("zvol %v mounted %v fs %v", volume, mount.MountPath, mount.FSType)

	return err
}

// MountDataset mounts the zfs dataset to the specified path
func MountDataset(vol *apis.ZFSVolume, mount *MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	mounted, err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return err
	}

	if mounted {
		klog.Infof("dataset : already mounted %s => %s", volume, mount.MountPath)
		return nil
	}

	val, err := GetVolumeProperty(vol, "mountpoint")
	if err != nil {
		return err
	}

	if val == "legacy" {
		var MountVolArg []string
		var mntopt string

		for _, option := range mount.MountOptions {
			mntopt += option + ","
		}

		MountVolArg = append(MountVolArg, "-o", mntopt, "-t", "zfs", volume, mount.MountPath)
		cmd := exec.Command("mount", MountVolArg...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			klog.Errorf("zfs: could not mount the dataset %v cmd %v error: %s",
				volume, MountVolArg, string(out))
			return status.Errorf(codes.Internal, "dataset: mount failed err : %s", string(out))
		}
		klog.Infof("dataset : legacy mounted %s => %s", volume, mount.MountPath)
	} else {
		/*
		 * We might have created volumes and then upgraded the node agent before
		 * getting the mount request for that volume. In this case volume will
		 * not be created with mountpoint as legacy. Handling the mount in old way.
		 */
		err = MountZFSDataset(vol, mount.MountPath)
		if err != nil {
			return status.Errorf(codes.Internal, "zfs: mount failed err : %s", err.Error())
		}
		klog.Infof("dataset : mounted %s => %s", volume, mount.MountPath)
	}

	return nil
}

// MountFilesystem mounts the disk to the specified path
func MountFilesystem(vol *apis.ZFSVolume, mount *MountInfo) error {
	// creating the directory with 0750 permission so that it can be accessed by other person.
	// if the directory already exist(old k8s), the creator should set the proper permission.
	if err := os.MkdirAll(mount.MountPath, 0750); err != nil {
		return status.Errorf(codes.Internal, "Could not create dir {%q}, err: %v", mount.MountPath, err)
	}

	switch vol.Spec.VolumeType {
	case VolTypeDataset:
		return MountDataset(vol, mount)
	default:
		return MountZvol(vol, mount)
	}
}

// MountBlock mounts the block disk to the specified path
func MountBlock(vol *apis.ZFSVolume, mountinfo *MountInfo) error {
	target := mountinfo.MountPath
	devicePath := ZFSDevPath + vol.Spec.PoolName + "/" + vol.Name
	mountopt := []string{"bind"}

	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}

	// Create the mount point as a file since bind mount device node requires it to be a file
	err := makeFile(target)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not create target file %q: %v", target, err)
	}

	// do the bind mount of the zvol device at the target path
	if err := mounter.Mount(devicePath, target, "", mountopt); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "mount failed at %v err : %v", target, err)
	}

	klog.Infof("NodePublishVolume mounted block device %s at %s", devicePath, target)

	return nil
}

func makeFile(pathname string) error {
	f, err := os.OpenFile(filepath.Clean(pathname), os.O_CREATE, os.FileMode(0644))
	defer f.Close()
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
