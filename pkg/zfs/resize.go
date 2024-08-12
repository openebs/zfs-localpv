/*
Copyright 2020 The Kubernetes Authors.

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

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
)

// ResizeExtn can be used to run a resize command on the ext2/3/4 filesystem
// to expand the filesystem to the actual size of the device
func ResizeExtn(devpath string) error {
	cmd := exec.Command("resize2fs", devpath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("zfspv: ResizeExtn failed error: %s", string(out))
		return err
	}
	return nil
}

// ResizeXFS can be used to run a resize command on the xfs filesystem
// to expand the filesystem to the actual size of the device
func ResizeXFS(path string) error {
	cmd := exec.Command("xfs_growfs", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("zfspv: ResizeXFS failed error: %s", string(out))
		return err
	}
	return nil
}

// handleVolResize resizes the filesystem, it is called after quota
// has been set on the volume. It takes care of expanding the filesystem.
func handleVolResize(vol *apis.ZFSVolume, volumePath string) error {
	var err error

	devpath, err := GetVolumeDevPath(vol)
	if err != nil {
		return err
	}

	fsType := vol.Spec.FsType

	mounter := mount.New("")
	list, _ := mounter.List()
	for _, mpt := range list {
		if mpt.Path == volumePath {
			switch fsType {
			case "xfs":
				err = ResizeXFS(volumePath)
			case "zfs":
				// just setting the quota is suffcient
				// nothing to handle here
				err = nil
			default:
				err = ResizeExtn(devpath)
			}
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}
