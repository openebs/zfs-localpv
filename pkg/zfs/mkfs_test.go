package zfs

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

// TestXfsOptionsForKernel walks the kernel releases of the distros through the
// format decision. Everything below 5.19 has to lose nrext64, everything from
// 5.19 keeps the mkfs defaults.
func TestXfsOptionsForKernel(t *testing.T) {
	tests := map[string]struct {
		release  string
		expected []string
	}{
		"rhel 8":              {release: "4.18.0-513.9.1.el8_9.x86_64", expected: nrext64Off},
		"debian 10":           {release: "4.19.0-25-amd64", expected: nrext64Off},
		"sles 15 sp2":         {release: "5.3.18-150300.59.87-default", expected: nrext64Off},
		"ubuntu 20.04":        {release: "5.4.0-192-generic", expected: nrext64Off},
		"rhel 9":              {release: "5.14.0-427.13.1.el9_4.x86_64", expected: nrext64Off},
		"ubuntu 22.04":        {release: "5.15.0-190-generic", expected: nrext64Off},
		"last kernel without": {release: "5.18.19-051819-generic", expected: nrext64Off},
		"first kernel with":   {release: "5.19.17-051917-generic", expected: nil},
		"debian 12":           {release: "6.1.0-13-amd64", expected: nil},
		"ubuntu 24.04":        {release: "6.8.0-137-generic", expected: nil},
		"aws 5.19":            {release: "5.19.0-1025-aws", expected: nil},
		"talos":               {release: "6.12.13-talos", expected: nil},
		"release candidate":   {release: "6.18.0-rc4", expected: nil},
		"next major":          {release: "7.0.0-generic", expected: nil},
		"unreadable release":  {release: "linux", expected: nrext64Off},
		"empty release":       {release: "", expected: nrext64Off},
		"major only":          {release: "6", expected: nrext64Off},
		"non numeric":         {release: "v5.19.0", expected: nrext64Off},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := xfsOptionsForKernel(test.release)
			if !reflect.DeepEqual(got, test.expected) {
				t.Fatalf("kernel %q: expected %v, got %v", test.release, test.expected, got)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := map[string]struct {
		version   string
		major     int
		minor     int
		expectErr bool
	}{
		"ubuntu kernel":     {version: "5.15.0-190-generic", major: 5, minor: 15},
		"rhel kernel":       {version: "4.18.0-513.9.1.el8_9.x86_64", major: 4, minor: 18},
		"rc kernel":         {version: "6.18.0-rc4", major: 6, minor: 18},
		"two digit minor":   {version: "5.19.0-1025-aws", major: 5, minor: 19},
		"no patch level":    {version: "7.1", major: 7, minor: 1},
		"trailing newline":  {version: " 6.5.0\n", major: 6, minor: 5},
		"no minor":          {version: "6", expectErr: true},
		"empty":             {version: "", expectErr: true},
		"not a version":     {version: "linux", expectErr: true},
		"no leading digits": {version: "v6.5.0", expectErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			major, minor, err := parseVersion(test.version)
			if test.expectErr {
				if err == nil {
					t.Fatalf("expected an error for %q, got %d.%d", test.version, major, minor)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", test.version, err)
			}
			if major != test.major || minor != test.minor {
				t.Fatalf("expected %d.%d for %q, got %d.%d",
					test.major, test.minor, test.version, major, minor)
			}
		})
	}
}

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		major, minor int
		expected     bool
	}{
		{4, 18, false},
		{5, 4, false},
		{5, 15, false},
		{5, 18, false},
		{5, 19, true},
		{5, 20, true},
		{6, 0, true},
		{6, 12, true},
		{7, 0, true},
	}

	for _, test := range tests {
		got := versionAtLeast(test.major, test.minor, nrext64KernelMajor, nrext64KernelMinor)
		if got != test.expected {
			t.Fatalf("expected %t for kernel %d.%d, got %t", test.expected, test.major, test.minor, got)
		}
	}
}

func TestMkfsOptions(t *testing.T) {
	defer func() { xfsMkfsOptions = nil }()
	xfsMkfsOptions = nrext64Off

	for _, fstype := range []string{"ext2", "ext3", "ext4", "btrfs", "zfs", ""} {
		if got := MkfsOptions(fstype); got != nil {
			t.Fatalf("expected no mkfs options for %q, got %v", fstype, got)
		}
	}

	if got := MkfsOptions("xfs"); !reflect.DeepEqual(got, nrext64Off) {
		t.Fatalf("expected %v for xfs, got %v", nrext64Off, got)
	}
}

// TestMkfsOptionsBeforeDetect checks that a format before the detection ran,
// which should not happen, falls back to the mkfs defaults.
func TestMkfsOptionsBeforeDetect(t *testing.T) {
	if got := MkfsOptions("xfs"); got != nil {
		t.Fatalf("expected no options before the detection ran, got %v", got)
	}
}

// TestXfsSupportsNrext64Opt runs the probe against the mkfs.xfs of the machine
// running the test and checks that it leaves no file behind.
func TestXfsSupportsNrext64Opt(t *testing.T) {
	if _, err := exec.LookPath("mkfs.xfs"); err != nil {
		t.Skip("no mkfs.xfs on this machine")
	}

	before := probeFiles(t)

	if !xfsSupportsNrext64Opt() {
		t.Log("this mkfs.xfs does not take -i nrext64=, expected only below xfsprogs 6.0")
	}

	if after := probeFiles(t); len(after) != len(before) {
		t.Fatalf("the probe left files behind: %v", after)
	}
}

func TestDetectMkfsOptions(t *testing.T) {
	defer func() { xfsMkfsOptions = nil }()

	DetectMkfsOptions()
	t.Logf("detected xfs mkfs options on this machine: %v", MkfsOptions("xfs"))
}

func probeFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("can not read %s: %v", os.TempDir(), err)
	}

	var found []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "zfspv-mkfs-probe") || strings.HasPrefix(entry.Name(), "zfspv-xfs-probe") {
			found = append(found, filepath.Join(os.TempDir(), entry.Name()))
		}
	}

	return found
}

// TestKernelTakesNrext64 runs the kernel probe on the machine running the test.
// It needs mkfs.xfs, a free loop device and the right to mount, so an
// inconclusive result is not a failure, it is what makes the driver fall back
// on the kernel version.
func TestKernelTakesNrext64(t *testing.T) {
	takes, conclusive := kernelTakesNrext64()
	if !conclusive {
		t.Skip("the kernel probe could not run here, the version fallback takes over")
	}

	t.Logf("this kernel mounts xfs with nrext64: %t", takes)

	release := probeRelease(t)
	major, minor, err := parseVersion(release)
	if err != nil {
		return
	}

	// an upstream kernel has to agree with its own version
	if byVersion := versionAtLeast(major, minor, nrext64KernelMajor, nrext64KernelMinor); byVersion != takes {
		t.Logf("kernel %s says %t by version but %t by probe, expected only on a custom kernel",
			release, byVersion, takes)
	}
}

// TestProbesLeaveNothingBehind checks that neither probe leaves a file or a
// loop device behind, loop devices are shared with the host.
func TestProbesLeaveNothingBehind(t *testing.T) {
	loopsBefore := attachedLoops(t)

	xfsSupportsNrext64Opt()
	kernelTakesNrext64()

	if left := probeFiles(t); len(left) != 0 {
		t.Fatalf("probe files left behind: %v", left)
	}

	if loopsAfter := attachedLoops(t); loopsAfter != loopsBefore {
		t.Fatalf("loop devices leaked: %d before, %d after", loopsBefore, loopsAfter)
	}
}

func probeRelease(t *testing.T) string {
	t.Helper()

	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		t.Fatalf("can not get the kernel release: %v", err)
	}

	return utsString(uts.Release[:])
}

func attachedLoops(t *testing.T) int {
	t.Helper()

	out, err := exec.Command("losetup", "--list", "--noheadings").CombinedOutput()
	if err != nil {
		return 0
	}

	return len(strings.Fields(strings.TrimSpace(string(out))))
}
