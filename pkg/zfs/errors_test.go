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

// withProbes swaps the package-level state probes for the duration of
// a test and restores them on cleanup. The probes are evaluated as
// "exists?" predicates against the supplied set membership tables.
func withProbes(t *testing.T, pools, datasets map[string]bool) {
	t.Helper()
	origPool, origDS := poolExistsFn, datasetExistsFn
	poolExistsFn = func(name string) bool { return pools[name] }
	datasetExistsFn = func(name string) bool { return datasets[name] }
	t.Cleanup(func() {
		poolExistsFn = origPool
		datasetExistsFn = origDS
	})
}

func TestPoolFromDataset(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"tank":               "tank",
		"tank/foo":           "tank",
		"tank/foo/bar":       "tank",
		"tank/foo@snap":      "tank",
		"tank/foo/bar@snap":  "tank",
		"tank@snap":          "tank",
	}
	for in, want := range cases {
		if got := poolFromDataset(in); got != want {
			t.Errorf("poolFromDataset(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsCreateOp(t *testing.T) {
	cases := map[string]bool{
		"zfs create":           true,
		"zfs clone":            true,
		"zfs snapshot":         true,
		"zfs recv (restore)":   true,
		"zfs destroy":          false,
		"zfs destroy snapshot": false,
		"zfs set":              false,
		"zfs list":             false,
		"zfs get refquota":     false,
	}
	for op, want := range cases {
		if got := isCreateOp(op); got != want {
			t.Errorf("isCreateOp(%q) = %v, want %v", op, got, want)
		}
	}
}

func TestClassifyByState_PoolMissing(t *testing.T) {
	withProbes(t, map[string]bool{}, map[string]bool{})
	if got := classifyByState("zfs create", "tank/foo"); got != ReasonPoolNotFound {
		t.Fatalf("got %q, want %q", got, ReasonPoolNotFound)
	}
}

func TestClassifyByState_CreateButExists(t *testing.T) {
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{"tank/foo": true},
	)
	if got := classifyByState("zfs create", "tank/foo"); got != ReasonDatasetExists {
		t.Fatalf("got %q, want %q", got, ReasonDatasetExists)
	}
}

func TestClassifyByState_DestroyButMissing(t *testing.T) {
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{},
	)
	if got := classifyByState("zfs destroy", "tank/foo"); got != ReasonDatasetNotFound {
		t.Fatalf("got %q, want %q", got, ReasonDatasetNotFound)
	}
}

func TestClassifyByState_SnapshotMissing(t *testing.T) {
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{"tank/foo": true},
	)
	if got := classifyByState("zfs destroy snapshot", "tank/foo@s1"); got != ReasonDatasetNotFound {
		t.Fatalf("got %q, want %q", got, ReasonDatasetNotFound)
	}
}

func TestClassifyByState_StateConsistent(t *testing.T) {
	// e.g. "zfs set" against a present dataset that nonetheless failed
	// (perhaps perms, invalid prop): state is consistent so we have no
	// opinion — caller will surface raw stderr.
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{"tank/foo": true},
	)
	if got := classifyByState("zfs set", "tank/foo"); got != ReasonUnknown {
		t.Fatalf("got %q, want %q", got, ReasonUnknown)
	}
}

func TestClassifyByState_EmptyDataset(t *testing.T) {
	if got := classifyByState("zfs list (pools)", ""); got != ReasonUnknown {
		t.Fatalf("got %q, want %q", got, ReasonUnknown)
	}
}

func TestNewZFSError_NilErr(t *testing.T) {
	if err := NewZFSError("zfs create", "tank/foo", nil, []byte("irrelevant")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewZFSError_PreservesStderrAndUnwrapsToSentinel(t *testing.T) {
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{"tank/foo": true},
	)
	base := fmt.Errorf("exit status 1")
	stderr := []byte("cannot create 'tank/foo': dataset already exists")
	err := NewZFSError("zfs create", "tank/foo", base, stderr)

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
	if zerr.Stderr != string(stderr) {
		t.Fatalf("stderr lost: got %q, want %q", zerr.Stderr, stderr)
	}
}

func TestNewZFSError_UnknownWhenStateConsistent(t *testing.T) {
	withProbes(t,
		map[string]bool{"tank": true},
		map[string]bool{"tank/foo": true},
	)
	base := fmt.Errorf("exit status 1")
	err := NewZFSError("zfs set", "tank/foo", base, []byte("permission denied"))
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
