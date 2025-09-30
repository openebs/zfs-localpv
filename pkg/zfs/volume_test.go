package zfs

import (
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
