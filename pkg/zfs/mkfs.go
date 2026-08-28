package zfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"k8s.io/klog/v2"
)

/*
 * mkfs.xfs from xfsprogs 6.5 onwards formats with 64-bit extent counters(the
 * nrext64 incompatible feature) by default, and only kernel 5.19 and above can
 * mount such a filesystem, the older ones refuse it with
 *
 *	XFS (zdN): Superblock has unknown incompatible features (0x20) enabled.
 *
 * The node agent shares the kernel that has to mount the volume, so it asks
 * that kernel whether it takes the feature and turns it off at format time when
 * it does not. The kernel version is only used when that probe can not be run,
 * a custom kernel can have the feature backported or left out.
 */
const (
	// first kernel that can mount xfs with nrext64
	nrext64KernelMajor = 5
	nrext64KernelMinor = 19

	// mkfs.xfs refuses to work with anything smaller than this
	mkfsProbeSize = 512 << 20

	// size of the image the kernel probe formats and mounts
	kernelProbeSize = 320 << 20
)

// nrext64Off is what gets added to the mkfs.xfs command line to drop nrext64.
var nrext64Off = []string{"-i", "nrext64=0"}

// xfsMkfsOptions is decided once, when the node agent starts.
var xfsMkfsOptions []string

// DetectMkfsOptions works out the mkfs options of this node and keeps them for
// the lifetime of the agent. The kernel and the tools do not change under a
// running agent, so nothing is probed per volume. Only the node agent needs it,
// the controller does not format.
func DetectMkfsOptions() {
	xfsMkfsOptions = xfsCompatOptions()
}

// MkfsOptions returns the options the driver adds to the mkfs command for the
// given filesystem type, on top of the ones k8s.io/mount-utils passes.
func MkfsOptions(fstype string) []string {
	if fstype != "xfs" {
		return nil
	}

	return xfsMkfsOptions
}

// xfsCompatOptions returns the mkfs.xfs options that keep the filesystem
// mountable by the kernel of this node.
func xfsCompatOptions() []string {
	// an mkfs.xfs that does not take the option does not set the feature either
	if !xfsSupportsNrext64Opt() {
		return nil
	}

	// what the kernel does with the feature beats what its version says, a
	// custom kernel can have it backported or left out
	if takes, conclusive := kernelTakesNrext64(); conclusive {
		if takes {
			klog.Infof("zfspv: this kernel mounts xfs with nrext64, formatting xfs with the mkfs defaults")
			return nil
		}

		klog.Infof("zfspv: this kernel does not mount xfs with nrext64, formatting without it")

		return nrext64Off
	}

	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		klog.Warningf("zfspv: can not get the kernel version(%v), formatting xfs without nrext64", err)
		return nrext64Off
	}

	release := utsString(uts.Release[:])
	options := xfsOptionsForKernel(release)
	klog.Infof("zfspv: went by the kernel version %s, formatting xfs with mkfs options %v", release, options)

	return options
}

/*
 * kernelTakesNrext64 reports whether the kernel of this node mounts an xfs
 * filesystem that has nrext64 set. It formats a throwaway image with the
 * feature and tries to mount it read only, which is the only answer the kernel
 * gives for sure, there is nothing in sysfs listing the xfs features it knows.
 *
 * The second return value is false when the probe could not be carried out, the
 * caller then has to fall back on the kernel version.
 */
func kernelTakesNrext64() (bool, bool) {
	dir, err := os.MkdirTemp("", "zfspv-xfs-probe")
	if err != nil {
		klog.Warningf("zfspv: can not create the kernel probe directory: %v", err)
		return false, false
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			klog.Warningf("zfspv: can not remove the kernel probe directory %s: %v", dir, err)
		}
	}()

	image := filepath.Join(dir, "image")
	if err := os.WriteFile(image, nil, 0600); err != nil {
		klog.Warningf("zfspv: can not create the kernel probe image: %v", err)
		return false, false
	}

	if err := os.Truncate(image, kernelProbeSize); err != nil {
		klog.Warningf("zfspv: can not size the kernel probe image: %v", err)
		return false, false
	}

	if out, err := exec.Command("mkfs.xfs", "-q", "-f", "-i", "nrext64=1", image).CombinedOutput(); err != nil {
		klog.Warningf("zfspv: can not format the kernel probe image(%v: %s)", err, strings.TrimSpace(string(out)))
		return false, false
	}

	device, err := attachLoop(image)
	if err != nil {
		klog.Warningf("zfspv: can not attach the kernel probe image: %v", err)
		return false, false
	}
	defer detachLoop(device)

	mountpoint := filepath.Join(dir, "mnt")
	if err := os.Mkdir(mountpoint, 0750); err != nil {
		klog.Warningf("zfspv: can not create the kernel probe mountpoint: %v", err)
		return false, false
	}

	err = syscall.Mount(device, mountpoint, "xfs", syscall.MS_RDONLY, "")
	switch {
	case err == nil:
		if err := syscall.Unmount(mountpoint, 0); err != nil {
			klog.Warningf("zfspv: can not unmount the kernel probe %s: %v", mountpoint, err)
		}

		return true, true

	case errors.Is(err, syscall.EINVAL):
		// the superblock was written by the mkfs of this image and mkfs was
		// happy with it, so the kernel turning it down is the feature check
		return false, true

	default:
		klog.Warningf("zfspv: the kernel probe could not be mounted: %v", err)
		return false, false
	}
}

// attachLoop puts the given file behind a loop device and returns it.
func attachLoop(image string) (string, error) {
	out, err := exec.Command("losetup", "--find", "--show", image).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("losetup: %v: %s", err, strings.TrimSpace(string(out)))
	}

	device := strings.TrimSpace(string(out))
	if device == "" {
		return "", fmt.Errorf("losetup returned no device")
	}

	return device, nil
}

// detachLoop frees a loop device. Loop devices are shared with the host, so one
// left behind is a leak.
func detachLoop(device string) {
	if out, err := exec.Command("losetup", "--detach", device).CombinedOutput(); err != nil {
		klog.Warningf("zfspv: can not detach %s(%v: %s)", device, err, strings.TrimSpace(string(out)))
	}
}

// xfsOptionsForKernel returns the mkfs.xfs options needed by the given kernel
// release. A release that can not be read is treated as an old one.
func xfsOptionsForKernel(release string) []string {
	major, minor, err := parseVersion(release)
	if err != nil {
		klog.Warningf("zfspv: can not parse the kernel release %q(%v), formatting xfs without nrext64", release, err)
		return nrext64Off
	}

	if versionAtLeast(major, minor, nrext64KernelMajor, nrext64KernelMinor) {
		return nil
	}

	return nrext64Off
}

// xfsSupportsNrext64Opt asks mkfs.xfs whether it takes "-i nrext64=". mkfs.xfs
// -N only parses the command line and prints the geometry it would use, so the
// probe writes nothing and needs no real device, a sparse file is enough.
func xfsSupportsNrext64Opt() bool {
	probe, err := os.CreateTemp("", "zfspv-mkfs-probe")
	if err != nil {
		klog.Warningf("zfspv: can not create the mkfs.xfs probe file(%v), formatting xfs with the mkfs defaults", err)
		return false
	}

	name := probe.Name()
	defer func() {
		if err := os.Remove(name); err != nil {
			klog.Warningf("zfspv: can not remove the mkfs.xfs probe file %s: %v", name, err)
		}
	}()

	if err := probe.Truncate(mkfsProbeSize); err != nil {
		_ = probe.Close()
		klog.Warningf("zfspv: can not size the mkfs.xfs probe file(%v), formatting xfs with the mkfs defaults", err)
		return false
	}

	if err := probe.Close(); err != nil {
		klog.Warningf("zfspv: can not close the mkfs.xfs probe file(%v), formatting xfs with the mkfs defaults", err)
		return false
	}

	args := append([]string{"-N"}, append(nrext64Off, name)...)
	if out, err := exec.Command("mkfs.xfs", args...).CombinedOutput(); err != nil {
		klog.Infof("zfspv: this mkfs.xfs does not take nrext64(%v: %s), formatting xfs with the mkfs defaults",
			err, strings.TrimSpace(string(out)))
		return false
	}

	return true
}

// utsString turns a null terminated uname field into a string.
func utsString(field []int8) string {
	out := make([]byte, 0, len(field))
	for _, c := range field {
		if c == 0 {
			break
		}
		out = append(out, byte(c))
	}

	return string(out)
}

// parseVersion picks the major and the minor version out of a version string,
// coping with the kernel releases of the distros, "5.15.0-190-generic",
// "4.18.0-513.9.1.el8_9.x86_64" and "6.18.0-rc4" are all valid input.
func parseVersion(version string) (int, int, error) {
	fields := strings.SplitN(strings.TrimSpace(version), ".", 3)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("can not parse version %q", version)
	}

	major, err := strconv.Atoi(leadingDigits(fields[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("can not parse version %q", version)
	}

	minor, err := strconv.Atoi(leadingDigits(fields[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("can not parse version %q", version)
	}

	return major, minor, nil
}

// leadingDigits returns the digits at the start of a version field, the rest is
// the distro specific part.
func leadingDigits(field string) string {
	for i, c := range field {
		if c < '0' || c > '9' {
			return field[:i]
		}
	}

	return field
}

// versionAtLeast tells if major.minor is not older than the wanted version.
func versionAtLeast(major, minor, wantMajor, wantMinor int) bool {
	if major != wantMajor {
		return major > wantMajor
	}

	return minor >= wantMinor
}
