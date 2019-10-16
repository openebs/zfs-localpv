package zfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/core/v1alpha1"
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

	logrus.Infof("created zvol %v and mounted %v fs %v", devicePath, mountInfo.MountPath, mountInfo.FSType)
	return nil
}

// UmountVolume unmounts the volume and the corresponding mount path is removed
func UmountVolume(vol *apis.ZFSVolume, targetPath string,
) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	_, _, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		logrus.Errorf(
			"zfspv umount volume: failed to get device from mnt: %s\nError: %v",
			targetPath, err,
		)
		return err
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

	if err = mounter.Unmount(targetPath); err != nil {
		logrus.Errorf(
			"zfspv umount volume: failed to unmount: %s\nError: %v",
			targetPath, err,
		)
		return err
	}

	if err := os.RemoveAll(targetPath); err != nil {
		logrus.Errorf("zfspv: failed to remove mount path Error: %v", err)
		return err
	}

	logrus.Infof("umount done path %v", targetPath)

	return nil
}

// GetMounts gets mountpoints for the specified volume
func GetMounts(devicepath string) ([]string, error) {

	var (
		currentMounts []string
		err           error
		mountList     []mount.MountPoint
	)

	dev, err := filepath.EvalSymlinks(devicepath)
	if err != nil {
		return nil, err
	}
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

// CreateAndMountZvol creates the zfs Volume
// and mounts the disk to the specified path
func CreateAndMountZvol(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	if len(mount.MountPath) == 0 {
		return status.Error(codes.InvalidArgument, "mount path missing in request")
	}

	if len(vol.Spec.OwnerNodeID) > 0 &&
		vol.Spec.OwnerNodeID != NodeID {
		return status.Error(codes.Internal, "volume is owned by different node")
	}

	devicePath, err := GetDevicePath(vol)
	if err != nil {
		return status.Error(codes.Internal, "not able to get the device path")
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
		return err
	} else if len(currentMounts) >= 1 {
		logrus.Errorf(
			"can not mount, more than one mounts for volume:%s dev %s mounts: %v",
			vol.Name, devicePath, currentMounts,
		)
		return status.Error(codes.Internal, "device already mounted")
	}
	err = FormatAndMountZvol(devicePath, mount)
	if err != nil {
		return status.Error(codes.Internal, "not able to mount the volume")
	}

	return err
}
