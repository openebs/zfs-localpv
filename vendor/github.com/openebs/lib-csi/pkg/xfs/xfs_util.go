/*
Copyright 2020 The OpenEBS Authors

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

package xfs

import (
	"github.com/openebs/lib-csi/pkg/mount"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"k8s.io/klog"
)

func xfsTempMount(device string) error {
	pvol := strings.Split(device, "/")
	volname := pvol[len(pvol)-1]

	// evaluate the symlink to get the dev path for volume
	dev, err := filepath.EvalSymlinks(device)
	if err != nil {
		return err
	}

	// create a temporary directory to mount the xfs file system
	tmpdir := "/tmp/" + volname
	err = os.Mkdir(tmpdir, 0755)
	if os.IsNotExist(err) {
		klog.Errorf("xfs: failed to create tmpdir %s error: %s", tmpdir, err.Error())
		return err
	}

	/*
	 * Device might have already mounted at the tmp path but umount might have failed
	 * in previous attempt. Checking here if device is not mounted then only attempt
	 * to mount it, otherwise proceed with the umount.
	 */
	curMounts, err := mount.GetMounts(dev)
	if err != nil {
		klog.Errorf("xfs: get mounts failed dev: %s err: %v", device, err.Error())
		return err
	} else if len(curMounts) == 0 {
		// mount with nouuid, so that it can play the log
		cmd := exec.Command("mount", "-o", "nouuid", device, tmpdir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			klog.Errorf("xfs: failed to mount device %s => %s error: %s", device, tmpdir, string(out))
			return err
		}
	} else {
		klog.Infof("xfs: device already mounted %s => [%v]", device, curMounts)
	}

	// log has been replayed, unmount the volume
	cmd := exec.Command("umount", tmpdir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("xfs: failed to umount tmpdir %s error: %s", tmpdir, string(out))
		return err
	}

	// remove the tmp directory
	err = os.Remove(tmpdir)
	if err != nil {
		// don't return error, reconciliation is not needed as umount is done
		klog.Errorf("xfs: failed to remove tmpdir %s error: %s", tmpdir, err.Error())
	}
	return nil
}

/*
* We have to generate a new UUID for the cloned volumes with xfs filesystem
* otherwise system won't let anyone mount it if UUID is same. Here, since cloned
* volume refers to the same block because of the way ZFS clone works, it will
* also have the same UUID.
* There might be something there in the xfs log, we have to clear them
* so that filesystem is clean and we can generate the UUID for it.
 */

// GenerateUUID generates a new UUID for the given device
func GenerateUUID(device string) error {
	// temporary mount the volume with nouuid to replay the logs
	err := xfsTempMount(device)
	if err != nil {
		return err
	}

	// for mounting the cloned volume for xfs, a new UUID has to be generated
	cmd := exec.Command("xfs_admin", "-U", "generate", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("xfs: uuid generate failed for device %s error: %s", device, string(out))
		return err
	}
	klog.Infof("xfs: generated UUID for the device %s \n %v", device, string(out))
	return nil
}
