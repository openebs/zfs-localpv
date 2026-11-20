package driver

import (
	"context"
	"testing"
	"time"

	zfsapi "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/builder/volbuilder"
	"github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned/fake"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

const reaperTestNamespace = "openebs"

func reaperTestVolume(name string, marked bool) *zfsapi.ZFSVolume {
	zv := &zfsapi.ZFSVolume{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: reaperTestNamespace, UID: types.UID("uid-" + name)},
	}
	if marked {
		zv.Annotations = map[string]string{volbuilder.MarkForDeletionAnnotation: "true"}
	}
	return zv
}

func reaperTestSnapshot(name, volName string) *zfsapi.ZFSSnapshot {
	return &zfsapi.ZFSSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: reaperTestNamespace,
			Labels:    map[string]string{zfs.ZFSVolKey: volName},
		},
	}
}

func volumeExists(t *testing.T, cs *fake.Clientset, name string) bool {
	t.Helper()
	_, err := cs.ZfsV1().ZFSVolumes(reaperTestNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func TestMarkedVolumeReaperReap(t *testing.T) {
	tests := map[string]struct {
		objects    []runtime.Object
		wantExists bool
	}{
		"marked, no snapshots left: deleted": {
			objects:    []runtime.Object{reaperTestVolume("pvc-a", true)},
			wantExists: false,
		},
		"marked, snapshots remain: kept": {
			objects: []runtime.Object{
				reaperTestVolume("pvc-a", true),
				reaperTestSnapshot("snap-1", "pvc-a"),
			},
			wantExists: true,
		},
		"not marked, no snapshots: kept": {
			objects:    []runtime.Object{reaperTestVolume("pvc-a", false)},
			wantExists: true,
		},
		"marked, only another volume's snapshots: deleted": {
			objects: []runtime.Object{
				reaperTestVolume("pvc-a", true),
				reaperTestVolume("pvc-b", false),
				reaperTestSnapshot("snap-1", "pvc-b"),
			},
			wantExists: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset(tt.objects...)
			r := newMarkedVolumeReaper(cs, reaperTestNamespace, newVolumeLock())
			require.NoError(t, r.reap(context.TODO(), "pvc-a"))
			assert.Equal(t, tt.wantExists, volumeExists(t, cs, "pvc-a"))
		})
	}
}

func TestMarkedVolumeReaperReapMissingVolume(t *testing.T) {
	cs := fake.NewSimpleClientset()
	r := newMarkedVolumeReaper(cs, reaperTestNamespace, newVolumeLock())
	assert.NoError(t, r.reap(context.TODO(), "pvc-gone"))
}

func TestMarkedVolumeReaperWaitsForVolumeLock(t *testing.T) {
	cs := fake.NewSimpleClientset(reaperTestVolume("pvc-a", true))
	lock := newVolumeLock()
	r := newMarkedVolumeReaper(cs, reaperTestNamespace, lock)

	// A concurrent CreateSnapshot holds the volume lock and adds a snapshot
	// before releasing it; the reaper must see that snapshot.
	unlock := lock.LockVolume("pvc-a")
	done := make(chan error)
	go func() { done <- r.reap(context.TODO(), "pvc-a") }()
	_, err := cs.ZfsV1().ZFSSnapshots(reaperTestNamespace).Create(context.TODO(),
		reaperTestSnapshot("snap-1", "pvc-a"), metav1.CreateOptions{})
	require.NoError(t, err)
	unlock()

	require.NoError(t, <-done)
	assert.True(t, volumeExists(t, cs, "pvc-a"))
}

func TestMarkedVolumeReaperEventHandlers(t *testing.T) {
	marked := reaperTestVolume("pvc-a", true)
	deleting := reaperTestVolume("pvc-b", true)
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	unmarked := reaperTestVolume("pvc-c", false)

	r := newMarkedVolumeReaper(fake.NewSimpleClientset(), reaperTestNamespace, newVolumeLock())
	r.volumeChanged(marked)
	r.volumeChanged(deleting)
	r.volumeChanged(unmarked)
	assert.Equal(t, 1, r.queue.Len(), "only the marked volume not yet being deleted is queued")

	r = newMarkedVolumeReaper(fake.NewSimpleClientset(), reaperTestNamespace, newVolumeLock())
	snap := reaperTestSnapshot("snap-1", "pvc-a")
	r.snapshotUpdated(snap, snap)
	assert.Equal(t, 0, r.queue.Len(), "an update that keeps the volume label queues nothing")
	r.snapshotUpdated(snap, reaperTestSnapshot("snap-1", "pvc-b"))
	assert.Equal(t, 1, r.queue.Len(), "a relabelled snapshot queues its old volume")
	item, _ := r.queue.Get()
	assert.Equal(t, "pvc-a", item)
}

// startTestReaper runs the reaper against cs until the test ends.
func startTestReaper(t *testing.T, cs *fake.Clientset) {
	t.Helper()
	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })
	r := newMarkedVolumeReaper(cs, reaperTestNamespace, newVolumeLock())
	require.NoError(t, r.Start(stopCh))
}

func waitForVolumeGone(t *testing.T, cs *fake.Clientset, name string) {
	t.Helper()
	err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Millisecond, 5*time.Second, true,
		func(context.Context) (bool, error) { return !volumeExists(t, cs, name), nil })
	require.NoError(t, err, "volume %s was not deleted", name)
}

func TestMarkedVolumeReaperDeletesAfterLastSnapshotDestroyed(t *testing.T) {
	cs := fake.NewSimpleClientset(
		reaperTestVolume("pvc-a", true),
		reaperTestSnapshot("snap-1", "pvc-a"),
		reaperTestSnapshot("snap-2", "pvc-a"),
	)
	startTestReaper(t, cs)
	snaps := cs.ZfsV1().ZFSSnapshots(reaperTestNamespace)

	require.NoError(t, snaps.Delete(context.TODO(), "snap-1", metav1.DeleteOptions{}))
	time.Sleep(200 * time.Millisecond)
	assert.True(t, volumeExists(t, cs, "pvc-a"), "kept while snap-2 remains")

	require.NoError(t, snaps.Delete(context.TODO(), "snap-2", metav1.DeleteOptions{}))
	waitForVolumeGone(t, cs, "pvc-a")
}

func TestMarkedVolumeReaperDeletesAlreadyOrphanedVolumeOnStart(t *testing.T) {
	// The state #776 leaves behind: marked, no snapshots, nothing pending.
	cs := fake.NewSimpleClientset(
		reaperTestVolume("pvc-a", true),
		reaperTestVolume("pvc-b", false),
	)
	startTestReaper(t, cs)
	waitForVolumeGone(t, cs, "pvc-a")
	assert.True(t, volumeExists(t, cs, "pvc-b"))
}
