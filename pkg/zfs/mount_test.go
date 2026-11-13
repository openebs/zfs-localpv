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
	"path/filepath"
	"testing"

	"k8s.io/mount-utils"
)

// TestMountBlockIdempotent checks that publishing a block volume again at a
// target which is already bind mounted does not mount it a second time.
func TestMountBlockIdempotent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "pvc-1")
	mounter := mount.NewFakeMounter(nil)
	device := ZFSDevPath + "zfspv-pool/pvc-1"

	for i := range 2 {
		if err := mountBlock(mounter, device, target); err != nil {
			t.Fatalf("publish %d: unexpected error: %v", i+1, err)
		}
	}

	mounts, err := mounter.List()
	if err != nil {
		t.Fatalf("list mounts: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount at %s, got %d: %v", target, len(mounts), mounts)
	}
	if mounts[0].Device != device {
		t.Errorf("expected device %s, got %s", device, mounts[0].Device)
	}
	if got := mounter.GetLog(); len(got) != 1 || got[0].Action != mount.FakeActionMount {
		t.Errorf("expected a single mount action, got %v", got)
	}
}
