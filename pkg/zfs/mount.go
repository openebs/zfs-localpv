package zfs

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
)

// FormatAndMountZvol formats and mounts the created volume to the desired mount path
func FormatAndMountZvol(devicePath string, mountInfo *apis.MountInfo) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	err := mounter.FormatAndMount(devicePath, mountInfo.MountPath, mountInfo.FSType, mountInfo.MountOptions)
	if err != nil {
		logrus.Errorf(
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
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	dev, ref, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		logrus.Errorf(
			"zfspv umount volume: failed to get device from mnt: %s\nError: %v",
			targetPath, err,
		)
		return err
	}

	// device has already been un-mounted, return successful
	if len(dev) == 0 || ref == 0 {
		logrus.Warningf(
			"Warning: Unmount skipped because volume %s not mounted: %v",
			vol.Name, targetPath,
		)
		return nil
	}

	if pathExists, pathErr := mount.PathExists(targetPath); pathErr != nil {
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		logrus.Warningf(
			"Warning: Unmount skipped because path does not exist: %v",
			targetPath,
		)
		return nil
	}

	if vol.Spec.VolumeType == VOLTYPE_DATASET {
		if err = UmountZFSDataset(vol); err != nil {
			logrus.Errorf(
				"zfspv failed to umount dataset: path %s Error: %v",
				targetPath, err,
			)
			return err
		}
	} else {
		if err = mounter.Unmount(targetPath); err != nil {
			logrus.Errorf(
				"zfspv failed to unmount zvol: path %s Error: %v",
				targetPath, err,
			)
			return err
		}
	}

	if err := os.Remove(targetPath); err != nil {
		logrus.Errorf("zfspv: failed to remove mount path Error: %v", err)
	}

	logrus.Infof("umount done path %v", targetPath)

	return nil
}

// GetMounts gets mountpoints for the specified volume
func GetMounts(dev string) ([]string, error) {

	var (
		currentMounts []string
		err           error
		mountList     []mount.MountPoint
	)

	mounter := mount.New("")
	// Get list of mounted paths present with the node
	if mountList, err = mounter.List(); err != nil {
		return nil, err
	}
	for _, mntInfo := range mountList {
		if mntInfo.Device == dev {
			currentMounts = append(currentMounts, mntInfo.Path)
		}
	}
	return currentMounts, nil
}

// IsMountPath returns true if path is a mount path
func IsMountPath(path string) bool {

	var (
		err       error
		mountList []mount.MountPoint
	)

	mounter := mount.New("")
	// Get list of mounted paths present with the node
	if mountList, err = mounter.List(); err != nil {
		return false
	}
	for _, mntInfo := range mountList {
		if mntInfo.Path == path {
			return true
		}
	}
	return false
}

func verifyMountRequest(vol *apis.ZFSVolume, mountpath string) error {
	if len(mountpath) == 0 {
		return status.Error(codes.InvalidArgument, "mount path missing in request")
	}

	if len(vol.Spec.OwnerNodeID) > 0 &&
		vol.Spec.OwnerNodeID != NodeID {
		return status.Error(codes.Internal, "volume is owned by different node")
	}

	devicePath, err := GetVolumeDevPath(vol)
	if err != nil {
		logrus.Errorf("can not get device for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return err
	}

	/*
	 * This check is the famous *Wall Of North*
	 * It will not let the volume to be mounted
	 * at more than two places. The volume should
	 * be unmounted before proceeding to the mount
	 * operation.
	 */
	currentMounts, err := GetMounts(devicePath)
	if err != nil {
		logrus.Errorf("can not get mounts for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return err
	} else if len(currentMounts) >= 1 {
		logrus.Errorf(
			"can not mount, volume:%s already mounted dev %s mounts: %v",
			vol.Name, devicePath, currentMounts,
		)
		return status.Error(codes.Internal, "device already mounted")
	}
	return nil
}

// MountZvol mounts the disk to the specified path
func MountZvol(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return status.Error(codes.Internal, "zvol can not be mounted")
	}

	devicePath := ZFS_DEVPATH + volume

	err = FormatAndMountZvol(devicePath, mount)
	if err != nil {
		return status.Error(codes.Internal, "not able to format and mount the zvol")
	}

	logrus.Infof("zvol %v mounted %v fs %v", volume, mount.MountPath, mount.FSType)

	return err
}

// MountDataset mounts the zfs dataset to the specified path
func MountDataset(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return status.Error(codes.Internal, "dataset can not be mounted")
	}

	err = MountZFSDataset(vol, mount.MountPath)
	if err != nil {
		return status.Error(codes.Internal, "not able to mount the dataset")
	}

	logrus.Infof("dataset %v mounted %v", volume, mount.MountPath)

	return nil
}

// MountVolume mounts the disk to the specified path
func MountVolume(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	switch vol.Spec.VolumeType {
	case VOLTYPE_DATASET:
		return MountDataset(vol, mount)
	default:
		return MountZvol(vol, mount)
	}
}
