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
	"k8s.io/klog"
	"os/exec"
)

/*
* We have to generate a new UUID for the cloned volumes with btrfs filesystem
* otherwise system will mount the same volume if UUID is same. Here, since cloned
* volume refers to the same block because of the way ZFS clone works, it will
* also have the same UUID.
 */
func btrfsGenerateUuid(volume string) error {
	device := ZFS_DEVPATH + volume

	// for mounting the cloned volume for btrfs, a new UUID has to be generated
	cmd := exec.Command("btrfstune", "-f", "-u", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("btrfs: uuid generate failed %s error: %s", volume, string(out))
		return err
	}
	klog.Infof("btrfs: generated UUID for the cloned volume %s \n %v", volume, string(out))
	return nil
}
