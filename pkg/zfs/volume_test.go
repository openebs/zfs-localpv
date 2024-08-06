package zfs

import (
	"testing"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
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
