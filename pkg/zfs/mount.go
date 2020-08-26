package zfs

import (
	"fmt"
	"os"
	"os/exec"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// FormatAndMountZvol formats and mounts the created volume to the desired mount path
func FormatAndMountZvol(devicePath string, mountInfo *apis.MountInfo) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

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
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

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
		return status.Error(codes.InvalidArgument, "verifyMount: mount path missing in request")
	}

	if len(vol.Spec.OwnerNodeID) > 0 &&
		vol.Spec.OwnerNodeID != NodeID {
		return status.Error(codes.Internal, "verifyMount: volume is owned by different node")
	}
	if vol.Finalizers == nil {
		return status.Error(codes.Internal, "verifyMount: volume is not ready to be mounted")
	}

	devicePath, err := GetVolumeDevPath(vol)
	if err != nil {
		klog.Errorf("can not get device for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return status.Errorf(codes.Internal, "verifyMount: GetVolumePath failed %s", err.Error())
	}

	// if it is not a shared volume, then make sure it is not mounted to more than one path
	if vol.Spec.Shared != "yes" {
		/*
		 * This check is the famous *Wall Of North*
		 * It will not let the volume to be mounted
		 * at more than two places. The volume should
		 * be unmounted before proceeding to the mount
		 * operation.
		 */
		currentMounts, err := GetMounts(devicePath)
		if err != nil {
			klog.Errorf("can not get mounts for volume:%s dev %s err: %v",
				vol.Name, devicePath, err.Error())
			return status.Errorf(codes.Internal, "verifyMount: Getmounts failed %s", err.Error())
		} else if len(currentMounts) >= 1 {
			klog.Errorf(
				"can not mount, volume:%s already mounted dev %s mounts: %v",
				vol.Name, devicePath, currentMounts,
			)
			return status.Errorf(codes.Internal, "verifyMount: device already mounted at %s", currentMounts)
		}
	}
	return nil
}

// MountZvol mounts the disk to the specified path
func MountZvol(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return err
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
func MountDataset(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	volume := vol.Spec.PoolName + "/" + vol.Name
	err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return err
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
func MountFilesystem(vol *apis.ZFSVolume, mount *apis.MountInfo) error {
	switch vol.Spec.VolumeType {
	case VolTypeDataset:
		return MountDataset(vol, mount)
	default:
		return MountZvol(vol, mount)
	}
}

// MountBlock mounts the block disk to the specified path
func MountBlock(vol *apis.ZFSVolume, mountinfo *apis.MountInfo) error {
	target := mountinfo.MountPath
	devicePath := ZFSDevPath + vol.Spec.PoolName + "/" + vol.Name
	mountopt := []string{"bind"}

	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	// Create the mount point as a file since bind mount device node requires it to be a file
	err := mounter.MakeFile(target)
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
