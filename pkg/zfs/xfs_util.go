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

package zfs

import (
	"os"
	"os/exec"

	"strings"

	"k8s.io/klog"
)

func xfsTempMount(volume string) error {
	device := ZFS_DEVPATH + volume
	pvol := strings.Split(volume, "/")

	// create a temporary directory to mount the xfs file system
	tmpdir := "/tmp/" + pvol[1]
	err := os.Mkdir(tmpdir, 0755)
	if err != nil {
		klog.Errorf("xfs: failed to create tmpdir %s error: %s", tmpdir, err.Error())
		return err
	}

	// mount with nouuid, so that it can play the log
	cmd := exec.Command("mount", "-o", "nouuid", device, tmpdir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("xfs: failed to mount volume %s=>%s error: %s", device, tmpdir, string(out))
		return err
	}

	// log has been replayed, unmount the volume
	cmd = exec.Command("umount", tmpdir)
	out, err = cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("xfs: failed to umount tmpdir %s error: %s", tmpdir, string(out))
		return err
	}

	// remove the directory
	err = os.Remove(tmpdir)
	if err != nil {
		klog.Errorf("xfs: failed to remove tmpdir %s error: %s", tmpdir, err.Error())
		return err
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
func xfsGenerateUuid(volume string) error {
	device := ZFS_DEVPATH + volume

	// temporary mount the volume with nouuid to replay the logs
	err := xfsTempMount(volume)
	if err != nil {
		return err
	}

	// for mounting the cloned volume for xfs, a new UUID has to be generated
	cmd := exec.Command("xfs_admin", "-U", "generate", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("xfs: uuid generate failed %s error: %s", volume, string(out))
		return err
	}
	klog.Infof("xfs: generated UUID for the cloned volume %s \n %v", volume, string(out))
	return nil
}
