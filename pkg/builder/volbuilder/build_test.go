package volbuilder

import (
	"fmt"
	"testing"
)

// TestWithVolumeError checks that the error that failed a volume creation is
// recorded in the volume status, and cleared once the volume is created.
func TestWithVolumeError(t *testing.T) {
	createErr := fmt.Errorf(`zfs create on "tank/pvc-1": cannot create 'tank/pvc-1': out of space`)

	vol, err := NewBuilder().
		WithName("pvc-1").
		WithVolumeStatus("Failed").
		WithVolumeError(createErr).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vol.Status.Error == nil {
		t.Fatalf("status error not set")
	}
	if vol.Status.Error.Message != createErr.Error() {
		t.Errorf("status error message = %q, want %q", vol.Status.Error.Message, createErr.Error())
	}

	vol, err = BuildFrom(vol).
		WithVolumeStatus("Ready").
		WithVolumeError(nil).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vol.Status.Error != nil {
		t.Errorf("status error = %+v, want it cleared", vol.Status.Error)
	}
}
