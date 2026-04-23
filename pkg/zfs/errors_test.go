/*
Copyright 2024 The OpenEBS Authors

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
	"errors"
	"fmt"
	"testing"
)

func TestClassifyStderr(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		want   string
	}{
		{"pool not found", "cannot open 'tank': no such pool", ReasonPoolNotFound},
		{"dataset not found", "cannot open 'tank/foo': dataset does not exist", ReasonDatasetNotFound},
		{"snapshot not found", "cannot open 'tank/foo@s1': snapshot does not exist", ReasonDatasetNotFound},
		{"dataset exists", "cannot create 'tank/foo': dataset already exists", ReasonDatasetExists},
		{"snapshot exists", "cannot create snapshot 'tank/foo@s1': snapshot already exists", ReasonDatasetExists},
		{"dataset busy", "cannot destroy 'tank/foo': dataset is busy", ReasonDatasetBusy},
		{"dependent clones", "cannot destroy 'tank/foo@s': snapshot has dependent clones", ReasonDatasetBusy},
		{"out of space", "cannot create 'tank/foo': out of space", ReasonInsufficientSpace},
		{"quota exceeded", "cannot receive new filesystem stream: quota exceeded", ReasonInsufficientSpace},
		{"permission", "cannot mount 'tank/foo': permission denied", ReasonPermissionDenied},
		{"not permitted", "zfs: operation not permitted", ReasonPermissionDenied},
		{"invalid arg", "invalid argument for 'volsize'", ReasonInvalidArgument},
		{"invalid property", "invalid property: compression=foo", ReasonInvalidArgument},
		{"unknown", "some random unexpected error message", ReasonUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyStderr(tc.stderr)
			if got != tc.want {
				t.Fatalf("classifyStderr(%q) = %q, want %q", tc.stderr, got, tc.want)
			}
		})
	}
}

func TestNewZFSError_NilErr(t *testing.T) {
	if err := NewZFSError("zfs create", "tank/foo", nil, []byte("irrelevant")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewZFSError_UnwrapsToSentinel(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	err := NewZFSError("zfs create", "tank/foo", base, []byte("cannot create 'tank/foo': dataset already exists"))
	if !errors.Is(err, ErrDatasetExists) {
		t.Fatalf("errors.Is(err, ErrDatasetExists) = false, want true; err=%v", err)
	}
	if ReasonOf(err) != ReasonDatasetExists {
		t.Fatalf("ReasonOf = %q, want %q", ReasonOf(err), ReasonDatasetExists)
	}
	var zerr *ZFSError
	if !errors.As(err, &zerr) {
		t.Fatalf("errors.As *ZFSError failed")
	}
	if zerr.Op != "zfs create" || zerr.Dataset != "tank/foo" {
		t.Fatalf("fields lost: %+v", zerr)
	}
}

func TestNewZFSError_UnknownFallsBack(t *testing.T) {
	base := fmt.Errorf("exit status 1")
	err := NewZFSError("zfs create", "tank/foo", base, []byte("totally novel stderr"))
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("expected errors.Is ErrUnknown, got err=%v", err)
	}
}

func TestReasonOf_NonZFSError(t *testing.T) {
	if got := ReasonOf(nil); got != "" {
		t.Fatalf("ReasonOf(nil) = %q, want empty", got)
	}
	if got := ReasonOf(errors.New("plain")); got != ReasonUnknown {
		t.Fatalf("ReasonOf(plain) = %q, want %q", got, ReasonUnknown)
	}
}


