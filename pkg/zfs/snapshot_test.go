package zfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeZFS is a zfs command on PATH that knows the datasets in its state
// file and handles `zfs list` and `zfs destroy` of them
const fakeZFS = `#!/bin/sh
state="$(dirname "$0")/datasets"
echo "$*" >> "$(dirname "$0")/calls"
case "$1" in
list)
	if [ "$2" = "-H" ]; then
		for pool; do :; done
		grep "^$pool/[^/]*@" "$state"
		exit 0
	fi
	grep -qx "$2" "$state" && exit 0
	echo "cannot open '$2': dataset does not exist" >&2
	exit 1
	;;
destroy)
	grep -qx "$2" "$state" || { echo "cannot open '$2': dataset does not exist" >&2; exit 1; }
	grep -vx "$2" "$state" > "$state.new"
	mv "$state.new" "$state"
	;;
*)
	exit 2
	;;
esac
`

// withFakeZFS puts fakeZFS first on PATH with the given datasets and returns
// a function reading back the datasets left
func withFakeZFS(t *testing.T, datasets ...string) func() []string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "zfs"), []byte(fakeZFS), 0o755); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(dir, "datasets")
	if err := os.WriteFile(state, []byte(strings.Join(datasets, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return func() []string {
		raw, err := os.ReadFile(state)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Fields(string(raw))
	}
}

func TestDestroySnapshot(t *testing.T) {
	tests := map[string]struct {
		datasets []string
		wantLeft []string
		wantErr  bool
	}{
		"at the labelled volume": {
			datasets: []string{"pool", "pool/pvc-a", "pool/pvc-a@snapshot-1", "pool/pvc-a@snapshot-2"},
			wantLeft: []string{"pool", "pool/pvc-a", "pool/pvc-a@snapshot-2"},
		},
		// zfs promote pvc-b moved snapshot-1 from pvc-a to pvc-b
		"moved by a promote": {
			datasets: []string{"pool", "pool/pvc-a", "pool/pvc-b", "pool/pvc-b@snapshot-1", "pool/pvc-b@snapshot-2"},
			wantLeft: []string{"pool", "pool/pvc-a", "pool/pvc-b", "pool/pvc-b@snapshot-2"},
		},
		"already gone": {
			datasets: []string{"pool", "pool/pvc-a", "pool/pvc-b@snapshot-2", "pool/pvc-b/nested@snapshot-1"},
			wantLeft: []string{"pool", "pool/pvc-a", "pool/pvc-b@snapshot-2", "pool/pvc-b/nested@snapshot-1"},
		},
		"ambiguous": {
			datasets: []string{"pool", "pool/pvc-b@snapshot-1", "pool/pvc-c@snapshot-1"},
			wantLeft: []string{"pool", "pool/pvc-b@snapshot-1", "pool/pvc-c@snapshot-1"},
			wantErr:  true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			left := withFakeZFS(t, test.datasets...)
			snap := &apis.ZFSSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "snapshot-1",
					Labels: map[string]string{ZFSVolKey: "pvc-a"},
				},
				Spec: apis.VolumeInfo{PoolName: "pool"},
			}

			err := DestroySnapshot(snap)
			if (err != nil) != test.wantErr {
				t.Fatalf("want error %v, got %v", test.wantErr, err)
			}
			if got := strings.Join(left(), " "); got != strings.Join(test.wantLeft, " ") {
				t.Fatalf("datasets left: want %v, got %s", test.wantLeft, got)
			}
		})
	}
}

func TestMatchSnapshot(t *testing.T) {
	out := []byte("pool/sub/pvc-a@snapshot-1\npool/pvc-b@snapshot-10\npool/pvc-b@snapshot-1\npool/pvc-b/x@snapshot-1\n")
	if got, err := matchSnapshot(out, "pool", "snapshot-1"); err != nil || got != "pool/pvc-b@snapshot-1" {
		t.Fatalf("want pool/pvc-b@snapshot-1, got %q %v", got, err)
	}
	if got, err := matchSnapshot(out, "pool/sub", "snapshot-1"); err != nil || got != "pool/sub/pvc-a@snapshot-1" {
		t.Fatalf("want pool/sub/pvc-a@snapshot-1, got %q %v", got, err)
	}
	if got, err := matchSnapshot(nil, "pool", "snapshot-1"); err != nil || got != "" {
		t.Fatalf("want no match, got %q %v", got, err)
	}
}
