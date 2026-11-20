package zfs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakePromoteZFS is a zfs command on PATH that answers `zfs get origin`
// from its origin file, fails `zfs promote` when the file promote-error
// exists, and prints its snapshots file for `zfs list`
const fakePromoteZFS = `#!/bin/sh
dir="$(dirname "$0")"
echo "$*" >> "$dir/calls"
case "$1" in
get)
	cat "$dir/origin"
	;;
promote)
	if [ -f "$dir/promote-error" ]; then
		cat "$dir/promote-error" >&2
		exit 1
	fi
	echo "-" > "$dir/origin"
	;;
list)
	cat "$dir/snapshots"
	;;
*)
	exit 2
	;;
esac
`

// withFakePromoteZFS puts fakePromoteZFS first on PATH and returns its
// directory, for the test to fill and to read the calls back from
func withFakePromoteZFS(t *testing.T, origin string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "zfs"), []byte(fakePromoteZFS), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "origin"), []byte(origin+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func fakeCalls(t *testing.T, dir string) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(raw)), "\n")
}

func promoteTestVolume() *apis.ZFSVolume {
	return &apis.ZFSVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "pvc-c"},
		Spec:       apis.VolumeInfo{PoolName: "pool"},
	}
}

func TestPromoteVolume(t *testing.T) {
	t.Run("clone", func(t *testing.T) {
		dir := withFakePromoteZFS(t, "pool/pvc-t@snapshot-1")
		promoted, err := PromoteVolume(promoteTestVolume())
		if err != nil || !promoted {
			t.Fatalf("want promoted, got %v %v", promoted, err)
		}
		want := []string{"get -pH -o value origin pool/pvc-c", "promote pool/pvc-c"}
		if got := fakeCalls(t, dir); !reflect.DeepEqual(got, want) {
			t.Fatalf("calls: want %q, got %q", want, got)
		}
	})

	t.Run("already promoted", func(t *testing.T) {
		dir := withFakePromoteZFS(t, "-")
		promoted, err := PromoteVolume(promoteTestVolume())
		if err != nil || promoted {
			t.Fatalf("want nothing done, got %v %v", promoted, err)
		}
		want := []string{"get -pH -o value origin pool/pvc-c"}
		if got := fakeCalls(t, dir); !reflect.DeepEqual(got, want) {
			t.Fatalf("calls: want %q, got %q", want, got)
		}
	})

	t.Run("promote fails", func(t *testing.T) {
		dir := withFakePromoteZFS(t, "pool/pvc-t@snapshot-1")
		stderr := "cannot promote 'pool/pvc-c': out of space"
		if err := os.WriteFile(filepath.Join(dir, "promote-error"), []byte(stderr), 0o644); err != nil {
			t.Fatal(err)
		}
		promoted, err := PromoteVolume(promoteTestVolume())
		if err == nil || promoted {
			t.Fatalf("want an error, got %v %v", promoted, err)
		}
		if !strings.Contains(err.Error(), stderr) {
			t.Fatalf("want the zfs stderr in the error, got %v", err)
		}
	})
}

func TestListVolumeSnapshots(t *testing.T) {
	dir := withFakePromoteZFS(t, "-")
	out := "pool/pvc-c@snapshot-1\npool/pvc-c@snapshot-2\n"
	if err := os.WriteFile(filepath.Join(dir, "snapshots"), []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ListVolumeSnapshots(promoteTestVolume())
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"snapshot-1", "snapshot-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %q, got %q", want, got)
	}
	want := []string{"list -H -o name -t snapshot -d 1 pool/pvc-c"}
	if calls := fakeCalls(t, dir); !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls: want %q, got %q", want, calls)
	}
}

func TestParseSnapshotNames(t *testing.T) {
	out := []byte("pool/pvc-c@snapshot-1\npool/pvc-c/child@snapshot-2\npool/pvc-cc@snapshot-3\npool/pvc-c@snapshot-4\n")
	got := parseSnapshotNames(out, "pool/pvc-c")
	if want := []string{"snapshot-1", "snapshot-4"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %q, got %q", want, got)
	}
	if got := parseSnapshotNames(nil, "pool/pvc-c"); got != nil {
		t.Fatalf("want none, got %q", got)
	}
}
