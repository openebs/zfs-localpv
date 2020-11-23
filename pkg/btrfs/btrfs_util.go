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

package btrfs

import (
	"os/exec"

	"k8s.io/klog"
)

/*
* We have to generate a new UUID for the cloned volumes with btrfs filesystem
* otherwise system will mount the same volume if UUID is same. Here, since cloned
* volume refers to the same block because of the way ZFS clone works, it will
* also have the same UUID.
 */

// GenerateUUID generates a new btrfs UUID for the given device
func GenerateUUID(device string) error {
	// for mounting the cloned volume for btrfs, a new UUID has to be generated
	cmd := exec.Command("btrfstune", "-f", "-u", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("btrfs: uuid generate failed for device %s error: %s", device, string(out))
		return err
	}
	klog.Infof("btrfs: generated UUID for the device %s \n %v", device, string(out))
	return nil
}
