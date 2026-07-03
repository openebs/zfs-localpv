package zfs

import (
	"strings"
	"testing"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsVolumeReady(t *testing.T) {
	volGen := func(finalizer string, state string) *apis.ZFSVolume {
		vol := &apis.ZFSVolume{}
		if finalizer != "" {
			vol.Finalizers = append(vol.Finalizers, finalizer)
		}
		if state != "" {
			vol.Status.State = state
		}
		return vol
	}
	tests := []struct {
		name         string
		volFinalizer string
		volState     string
		want         bool
	}{
		{"Older volume is ready", ZFSFinalizer, "", true},
		{"Older volume is not ready", "", "", false},
		{"Newer volume is pending", "", ZFSStatusPending, false},
		{"Newer volume is pending with finalizer", ZFSFinalizer, ZFSStatusPending, false},
		{"Newer volume is ready with finalizer", ZFSFinalizer, ZFSStatusReady, true},
		{"Newer volume is failed", "", ZFSStatusFailed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsVolumeReady(volGen(tt.volFinalizer, tt.volState)); got != tt.want {
				t.Errorf("IsVolumeReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReservationProperty(t *testing.T) {
	tests := []struct {
		name      string
		quotaType string
		capacity  string
		expected  string
	}{
		{"Empty quotaType defaults to quota", "", "10G", "reservation=10G"},
		{"Valid quotaType quota", "quota", "5G", "reservation=5G"},
		{"Valid quotaType refquota", "refquota", "3G", "refreservation=3G"},
		{"Invalid quotaType defaults to quota", "invalid", "2G", "reservation=2G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reservationProperty(tt.quotaType, tt.capacity)
			if result != tt.expected {
				t.Errorf("For quotaType %q and capacity %q, expected %q but got %q",
					tt.quotaType, tt.capacity, tt.expected, result)
			}
		})
	}
}

func TestQuotaProperty(t *testing.T) {
	tests := []struct {
		name      string
		quotaType string
		expected  string
	}{
		{"Empty quotaType defaults to quota", "", "quota"},
		{"Valid quotaType quota", "quota", "quota"},
		{"Valid quotaType refquota", "refquota", "refquota"},
		{"Invalid quotaType defaults to quota", "invalid", "quota"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quotaProperty(tt.quotaType)
			if result != tt.expected {
				t.Errorf("For quotaType %q, expected %q but got %q",
					tt.quotaType, tt.expected, result)
			}
		})
	}
}

func TestBuildCloneCreateArgs(t *testing.T) {
	tests := []struct {
		name     string
		vol      *apis.ZFSVolume
		expected []string
	}{
		{
			name: "Clone without encryption",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vol",
				},
				Spec: apis.VolumeInfo{
					PoolName:      "testpool",
					SnapName:      "testsnap",
					VolumeType:    "DATASET",
					Capacity:      "10G",
					RecordSize:    "128k",
					Compression:   "on",
					Dedup:         "off",
					ThinProvision: "yes",
				},
			},
			expected: []string{"clone", "-o", "quota=10G", "-o", "recordsize=128k", "-o", "mountpoint=legacy", "-o", "dedup=off", "-o", "compression=on", "testpool/testsnap", "testpool/test-vol"},
		},
		{
			name: "Clone with encryption parameters (should be omitted)",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-encrypted-vol",
				},
				Spec: apis.VolumeInfo{
					PoolName:      "testpool",
					SnapName:      "testsnap",
					VolumeType:    "DATASET",
					Capacity:      "10G",
					Compression:   "on",
					Encryption:    "aes-256-gcm",
					KeyLocation:   "file:///etc/zfs/keys/testpool.key",
					KeyFormat:     "raw",
					ThinProvision: "yes",
				},
			},
			expected: []string{"clone", "-o", "quota=10G", "-o", "mountpoint=legacy", "-o", "compression=on", "testpool/testsnap", "testpool/test-encrypted-vol"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCloneCreateArgs(tt.vol)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d arguments, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("Argument %d: expected %q, got %q", i, tt.expected[i], arg)
				}
			}
			// Verify encryption parameters are NOT included
			for _, arg := range result {
				if arg == "encryption=aes-256-gcm" || arg == "keylocation=file:///etc/zfs/keys/testpool.key" || arg == "keyformat=raw" {
					t.Errorf("Clone command should NOT include encryption parameter: %q", arg)
				}
			}
		})
	}
}

func assertArgs(t *testing.T, got, expected []string) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("Expected %d arguments, got %d\nExpected: %v\nGot: %v",
			len(expected), len(got), expected, got)
	}
	for i, arg := range got {
		if arg != expected[i] {
			t.Errorf("Argument %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}

func TestBuildDatasetCreateArgs(t *testing.T) {
	tests := []struct {
		name     string
		vol      *apis.ZFSVolume
		expected []string
	}{
		{
			name: "atime and logbias are set on dataset creation",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vol"},
				Spec: apis.VolumeInfo{
					PoolName:      "testpool",
					VolumeType:    "DATASET",
					Capacity:      "10G",
					RecordSize:    "16k",
					ATime:         "off",
					Compression:   "lz4",
					LogBias:       "throughput",
					ThinProvision: "yes",
				},
			},
			expected: []string{"create", "-o", "quota=10G", "-o", "recordsize=16k", "-o", "atime=off", "-o", "compression=lz4", "-o", "logbias=throughput", "-o", "mountpoint=legacy", "testpool/test-vol"},
		},
		{
			name: "atime and logbias omitted when unset",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "plain-vol"},
				Spec: apis.VolumeInfo{
					PoolName:      "testpool",
					VolumeType:    "DATASET",
					Capacity:      "10G",
					ThinProvision: "yes",
				},
			},
			expected: []string{"create", "-o", "quota=10G", "-o", "mountpoint=legacy", "testpool/plain-vol"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertArgs(t, buildDatasetCreateArgs(tt.vol), tt.expected)
		})
	}
}

func TestBuildZvolCreateArgs(t *testing.T) {
	vol := &apis.ZFSVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "test-zvol"},
		Spec: apis.VolumeInfo{
			PoolName:      "testpool",
			VolumeType:    "ZVOL",
			Capacity:      "10G",
			VolBlockSize:  "8k",
			Compression:   "lz4",
			ATime:         "off", // must be ignored for zvols
			LogBias:       "throughput",
			ThinProvision: "yes",
		},
	}
	expected := []string{"create", "-s", "-V", "10G", "-b", "8k", "-o", "compression=lz4", "-o", "logbias=throughput", "testpool/test-zvol"}
	assertArgs(t, buildZvolCreateArgs(vol), expected)
}

func TestBuildVolumeRestoreArgs(t *testing.T) {
	rstr := &apis.ZFSRestore{
		Spec: apis.ZFSRestoreSpec{RestoreSrc: "10.0.0.1:9000", VolumeName: "test-vol"},
		VolSpec: apis.VolumeInfo{
			PoolName:    "testpool",
			VolumeType:  "DATASET",
			Capacity:    "10G",
			RecordSize:  "16k",
			ATime:       "off",
			Compression: "lz4",
			LogBias:     "throughput",
		},
	}
	ncArgs, recvArgs, err := buildVolumeRestoreArgs(rstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nc := strings.Join(ncArgs, " ")
	if want := "-w 3 10.0.0.1 9000"; nc != want {
		t.Errorf("nc args = %q, want %q", nc, want)
	}
	cmd := strings.Join(recvArgs, " ")
	for _, want := range []string{"-o atime=off", "-o logbias=throughput"} {
		if !strings.Contains(cmd, want) {
			t.Errorf("restore recv command missing %q\ngot: %s", want, cmd)
		}
	}
}

// TestBuildVolumeBackupArgsSnapTokenization checks that a prevSnapName
// containing special characters ends up as a single argv element of `zfs send`,
// never split into separate tokens.
func TestBuildVolumeBackupArgsSnapTokenization(t *testing.T) {
	name := "snap with spaces; and & chars"
	bkp := &apis.ZFSBackup{
		Spec: apis.ZFSBackupSpec{
			BackupDest:   "10.0.0.1:9000",
			SnapName:     "snap1",
			PrevSnapName: name,
		},
	}
	vol := &apis.ZFSVolume{
		Spec: apis.VolumeInfo{PoolName: "zfspv-pool"},
	}
	vol.Name = "pvc-test"

	sendArgs, ncArgs, err := buildVolumeBackupArgs(bkp, vol)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The value must appear verbatim as exactly one argv element.
	wantSnap := "zfspv-pool/pvc-test@" + name
	found := false
	for _, a := range sendArgs {
		if a == wantSnap {
			found = true
		}
		if a == name {
			t.Errorf("value leaked as a bare argv element: %q", a)
		}
	}
	if !found {
		t.Errorf("expected value contained in a single snap argument %q, got args: %#v", wantSnap, sendArgs)
	}
	if sendArgs[0] != ZFSSendArg {
		t.Errorf("send args[0] = %q, want %q", sendArgs[0], ZFSSendArg)
	}
	if want := "-w 3 10.0.0.1 9000"; strings.Join(ncArgs, " ") != want {
		t.Errorf("nc args = %q, want %q", strings.Join(ncArgs, " "), want)
	}
}

// TestBuildVolumeRestoreArgsPropertyTokenization checks that a keylocation
// value containing special characters stays a single argv element following its
// own "-o" flag.
func TestBuildVolumeRestoreArgsPropertyTokenization(t *testing.T) {
	value := "file:///tmp/key; and & chars"
	rstr := &apis.ZFSRestore{
		Spec: apis.ZFSRestoreSpec{RestoreSrc: "10.0.0.1:9000", VolumeName: "test-vol"},
		VolSpec: apis.VolumeInfo{
			PoolName:    "zfspv-pool",
			KeyLocation: value,
		},
	}

	_, recvArgs, err := buildVolumeRestoreArgs(rstr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// keylocation=<value> must be a single argv element following its own "-o".
	want := "keylocation=" + value
	found := false
	for i, a := range recvArgs {
		if a == want {
			found = true
			if i == 0 || recvArgs[i-1] != "-o" {
				t.Errorf("keylocation value not preceded by its own -o flag: %#v", recvArgs)
			}
		}
	}
	if !found {
		t.Errorf("expected keylocation value as a single argv element %q, got: %#v", want, recvArgs)
	}
}

func TestBuildVolumeSetArgs(t *testing.T) {
	tests := []struct {
		name     string
		vol      *apis.ZFSVolume
		expected []string
	}{
		{
			name: "atime (dataset) and logbias are emitted on set",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vol"},
				Spec: apis.VolumeInfo{
					PoolName:    "testpool",
					VolumeType:  "DATASET",
					RecordSize:  "16k",
					ATime:       "off",
					Compression: "lz4",
					LogBias:     "throughput",
				},
			},
			expected: []string{"set", "recordsize=16k", "atime=off", "compression=lz4", "logbias=throughput", "testpool/test-vol"},
		},
		{
			name: "atime is skipped for zvol, logbias still applies",
			vol: &apis.ZFSVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "zvol-vol"},
				Spec: apis.VolumeInfo{
					PoolName:   "testpool",
					VolumeType: "ZVOL",
					ATime:      "off",
					LogBias:    "throughput",
				},
			},
			expected: []string{"set", "logbias=throughput", "testpool/zvol-vol"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertArgs(t, buildVolumeSetArgs(tt.vol), tt.expected)
		})
	}
}
