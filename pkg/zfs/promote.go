/*
Copyright 2026 The OpenEBS Authors.

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
	"strings"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"k8s.io/klog/v2"
)

// noOrigin is the value of the origin property of a dataset that is not a clone
const noOrigin = "-"

// PromoteVolume runs `zfs promote` on the volume, so that it takes over the
// snapshot it was cloned from, and every earlier snapshot of its origin. It
// returns false without running anything when the volume has no origin, as
// after an earlier promote, so it is safe to call again.
func PromoteVolume(vol *apis.ZFSVolume) (bool, error) {
	origin, err := GetVolumeProperty(vol, "origin")
	if err != nil {
		return false, err
	}
	volume := vol.Spec.PoolName + "/" + vol.Name
	if origin == noOrigin {
		klog.Infof("promote: volume %s has no origin, nothing to promote", volume)
		return false, nil
	}

	args := []string{ZFSPromoteArg, volume}
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := runCmd(cmd, volume)
	if err != nil {
		zerr := NewZFSError("zfs promote", volume, err, out)
		klog.Errorf("zfs: could not promote volume %v cmd %v error: %s", volume, args, zerr)
		return false, zerr
	}
	klog.Infof("promoted volume %s, its origin was %s", volume, origin)
	return true, nil
}

// ListVolumeSnapshots returns the names (the part after the @) of the
// snapshots of the volume itself, not of any child dataset
func ListVolumeSnapshots(vol *apis.ZFSVolume) ([]string, error) {
	volume := vol.Spec.PoolName + "/" + vol.Name
	args := []string{ZFSListArg, "-H", "-o", "name", "-t", "snapshot", "-d", "1", volume}
	cmd := exec.Command(ZFSVolCmd, args...)
	out, err := runCmd(cmd, volume)
	if err != nil {
		zerr := NewZFSError("zfs list (snapshots)", volume, err, out)
		klog.Errorf("zfs: could not list snapshots of %v cmd %v error: %s", volume, args, zerr)
		return nil, zerr
	}
	return parseSnapshotNames(out, volume), nil
}

// parseSnapshotNames picks the snapshot names of volume out of
// `zfs list -H -o name -t snapshot` output
func parseSnapshotNames(out []byte, volume string) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		dataset, snap, ok := strings.Cut(line, "@")
		if ok && dataset == volume && snap != "" {
			names = append(names, snap)
		}
	}
	return names
}
