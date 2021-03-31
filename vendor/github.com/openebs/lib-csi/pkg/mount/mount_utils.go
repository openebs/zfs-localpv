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

package mount

import (
	"k8s.io/utils/mount"
)

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
