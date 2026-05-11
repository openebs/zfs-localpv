package zfs

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNewZFSError_NilErr(t *testing.T) {
	if err := NewZFSError("zfs create", "tank/foo", nil, []byte("irrelevant")); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestNewZFSError_PreservesFields(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	stderr := []byte("  cannot create 'tank/foo': dataset already exists\n")
	err := NewZFSError("zfs create", "tank/foo", base, stderr)

	var zerr *Error
	if !errors.As(err, &zerr) {
		t.Fatalf("errors.As *ZFSError failed")
	}
	if zerr.Op != "zfs create" {
		t.Errorf("Op = %q, want %q", zerr.Op, "zfs create")
	}
	if zerr.Dataset != "tank/foo" {
		t.Errorf("Dataset = %q, want %q", zerr.Dataset, "tank/foo")
	}
	wantStderr := strings.TrimSpace(string(stderr))
	if zerr.Stderr != wantStderr {
		t.Errorf("Stderr = %q, want %q", zerr.Stderr, wantStderr)
	}
}

func TestZFSError_ErrorContainsStderr(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	stderr := "cannot create 'tank/foo': out of space"
	err := NewZFSError("zfs create", "tank/foo", base, []byte(stderr))

	msg := err.Error()
	if !strings.Contains(msg, stderr) {
		t.Errorf("error message %q does not contain stderr %q", msg, stderr)
	}
	if !strings.Contains(msg, "tank/foo") {
		t.Errorf("error message %q does not contain dataset", msg)
	}
}

func TestZFSError_Unwrap(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	err := NewZFSError("zfs destroy", "tank/foo", base, []byte("dataset is busy"))
	if !errors.Is(err, base) {
		t.Errorf("errors.Is(err, base) = false, Unwrap should expose original error")
	}
}

func TestZFSError_EmptyStderrFallsBackToErr(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	err := NewZFSError("zfs create", "tank/foo", base, nil)
	msg := err.Error()
	if !strings.Contains(msg, "exit status 1") {
		t.Errorf("expected fallback to Err in message, got %q", msg)
	}
}
