package zfs

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/mount-utils"
)

// bindMounter records mounts like FakeMounter and makes target resolve to the
// same file as source, as a bind mount of a device node does.
type bindMounter struct {
	*mount.FakeMounter
}

func (m *bindMounter) Mount(source, target, fstype string, options []string) error {
	if err := os.Remove(target); err != nil {
		return err
	}
	if err := os.Link(source, target); err != nil {
		return err
	}
	return m.FakeMounter.Mount(source, target, fstype, options)
}

func newFakeDevice(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "zvol")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	device := filepath.Join(dir, name)
	if err := os.WriteFile(device, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return device
}

// TestMountBlockIdempotent checks that publishing a block volume again at a
// target which is already bind mounted does not mount it a second time.
func TestMountBlockIdempotent(t *testing.T) {
	target := filepath.Join(t.TempDir(), "pvc-1")
	mounter := &bindMounter{mount.NewFakeMounter(nil)}
	device := newFakeDevice(t, "pvc-1")

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

// TestMountBlockIdempotentSymlink checks that a device path given through a
// symlink, like /dev/zvol/<pool>/<vol>, matches the device mounted at target.
func TestMountBlockIdempotentSymlink(t *testing.T) {
	target := filepath.Join(t.TempDir(), "pvc-1")
	mounter := &bindMounter{mount.NewFakeMounter(nil)}
	device := newFakeDevice(t, "zd0")
	link := filepath.Join(t.TempDir(), "pvc-1")
	if err := os.Symlink(device, link); err != nil {
		t.Fatal(err)
	}

	if err := mountBlock(mounter, link, target); err != nil {
		t.Fatalf("first publish: unexpected error: %v", err)
	}
	if err := mountBlock(mounter, device, target); err != nil {
		t.Fatalf("second publish: unexpected error: %v", err)
	}
	if got := mounter.GetLog(); len(got) != 1 {
		t.Errorf("expected a single mount action, got %v", got)
	}
}

// TestMountBlockOtherDevice checks that publishing at a target which is
// already bind mounted with another device fails instead of reporting success.
func TestMountBlockOtherDevice(t *testing.T) {
	target := filepath.Join(t.TempDir(), "pvc-1")
	mounter := &bindMounter{mount.NewFakeMounter(nil)}
	deviceA := newFakeDevice(t, "pvc-a")
	deviceB := newFakeDevice(t, "pvc-b")

	if err := mountBlock(mounter, deviceA, target); err != nil {
		t.Fatalf("publish volume A: unexpected error: %v", err)
	}
	if err := mountBlock(mounter, deviceB, target); err == nil {
		t.Fatal("publish volume B at the target of volume A: expected an error")
	}
	if got := mounter.GetLog(); len(got) != 1 {
		t.Errorf("expected only the mount of volume A, got %v", got)
	}
}
