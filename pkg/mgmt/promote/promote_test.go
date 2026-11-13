package promote

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned/fake"
	informers "github.com/openebs/zfs-localpv/v2/pkg/generated/informer/externalversions"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

const (
	testNamespace = "openebs"
	testNode      = "node-1"
)

func testVolume(name string) *apis.ZFSVolume {
	return &apis.ZFSVolume{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
		Spec:       apis.VolumeInfo{PoolName: "pool", OwnerNodeID: testNode},
	}
}

func testSnapshot(name, volume string) *apis.ZFSSnapshot {
	return &apis.ZFSSnapshot{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: testNamespace,
		Labels:    map[string]string{zfs.ZFSVolKey: volume},
	}}
}

func testPromote(volume, node string, status apis.ZFSPromoteStatus) *apis.ZFSPromote {
	return &apis.ZFSPromote{
		ObjectMeta: metav1.ObjectMeta{Name: volume, Namespace: testNamespace},
		Spec:       apis.ZFSPromoteSpec{VolumeName: volume, OwnerNodeID: node},
		Status:     status,
	}
}

// fakeZFS stands in for the zfs commands: it records the volumes promoted
// and lists the snapshots given for each volume
type fakeZFS struct {
	origin     string // "-" for a volume that is not a clone
	promoteErr error
	snapshots  map[string][]string
	promoted   []string
	listed     []string
}

func (f *fakeZFS) ops() zfsOps {
	return zfsOps{
		promote: func(vol *apis.ZFSVolume) (bool, error) {
			if f.promoteErr != nil {
				return false, f.promoteErr
			}
			if f.origin == "-" {
				return false, nil
			}
			f.promoted = append(f.promoted, vol.Name)
			f.origin = "-"
			return true, nil
		},
		listSnapshots: func(vol *apis.ZFSVolume) ([]string, error) {
			f.listed = append(f.listed, vol.Name)
			return f.snapshots[vol.Name], nil
		},
	}
}

type harness struct {
	t        *testing.T
	client   *fake.Clientset
	ctrl     *PromoteController
	zfs      *fakeZFS
	recorder *record.FakeRecorder
}

func newHarness(t *testing.T, z *fakeZFS, objects ...runtime.Object) *harness {
	t.Helper()
	client := fake.NewSimpleClientset(objects...)
	factory := informers.NewSharedInformerFactory(client, 0)
	recorder := record.NewFakeRecorder(10)
	ctrl, err := newPromoteController(client, factory, recorder, testNamespace, testNode)
	if err != nil {
		t.Fatal(err)
	}
	ctrl.zfs = z.ops()
	t.Cleanup(ctrl.workqueue.ShutDown)
	return &harness{t: t, client: client, ctrl: ctrl, zfs: z, recorder: recorder}
}

// sync runs syncPromote on the stored ZFSPromote of the volume
func (h *harness) sync(name string) error {
	h.t.Helper()
	p, err := h.client.ZfsV1().ZFSPromotes(testNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.ctrl.syncPromote(context.TODO(), p)
}

func (h *harness) status(name string) apis.ZFSPromoteStatus {
	h.t.Helper()
	p, err := h.client.ZfsV1().ZFSPromotes(testNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		h.t.Fatal(err)
	}
	return p.Status
}

func (h *harness) label(snap string) string {
	h.t.Helper()
	s, err := h.client.ZfsV1().ZFSSnapshots(testNamespace).Get(context.TODO(), snap, metav1.GetOptions{})
	if err != nil {
		h.t.Fatal(err)
	}
	return s.Labels[zfs.ZFSVolKey]
}

func (h *harness) events() []string {
	var events []string
	for {
		select {
		case e := <-h.recorder.Events:
			events = append(events, e)
		default:
			return events
		}
	}
}

// statusWrites returns the statuses written to ZFSPromotes, in order
func (h *harness) statusWrites() []apis.ZFSPromoteStatus {
	var writes []apis.ZFSPromoteStatus
	for _, a := range h.client.Actions() {
		if a.GetVerb() == "update" && a.GetResource().Resource == "zfspromotes" {
			obj := a.(interface{ GetObject() runtime.Object }).GetObject()
			writes = append(writes, obj.(*apis.ZFSPromote).Status)
		}
	}
	return writes
}

func TestPromoteClone(t *testing.T) {
	// pvc-c was cloned from pvc-t@snapshot-2; promoting it moves snapshot-1
	// and snapshot-2 from pvc-t, while snapshot-3 was taken of pvc-c and
	// pvc-x is a volume-to-volume clone's snapshot, without a record
	z := &fakeZFS{
		origin:    "pool/pvc-t@snapshot-2",
		snapshots: map[string][]string{"pvc-c": {"snapshot-1", "snapshot-2", "pvc-x", "snapshot-3"}},
	}
	h := newHarness(t, z,
		testVolume("pvc-c"), testVolume("pvc-t"),
		testSnapshot("snapshot-1", "pvc-t"),
		testSnapshot("snapshot-2", "pvc-t"),
		testSnapshot("snapshot-3", "pvc-c"),
		testSnapshot("snapshot-9", "pvc-t"), // taken after the clone point, stays
		testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit),
	)

	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(z.promoted, []string{"pvc-c"}) {
		t.Fatalf("promoted: want [pvc-c], got %v", z.promoted)
	}
	for snap, want := range map[string]string{
		"snapshot-1": "pvc-c", "snapshot-2": "pvc-c", "snapshot-3": "pvc-c", "snapshot-9": "pvc-t",
	} {
		if got := h.label(snap); got != want {
			t.Errorf("%s: want label %s, got %s", snap, want, got)
		}
	}
	want := []apis.ZFSPromoteStatus{apis.PromoteZFSStatusInProgress, apis.PromoteZFSStatusDone}
	if got := h.statusWrites(); !reflect.DeepEqual(got, want) {
		t.Fatalf("status writes: want %v, got %v", want, got)
	}
	if ev := h.events(); len(ev) != 1 || !strings.HasPrefix(ev[0], "Normal "+zfs.ReasonPromoted) {
		t.Fatalf("events: want one %s, got %q", zfs.ReasonPromoted, ev)
	}

	// snapshot-3 already carried pvc-c and is not written
	for _, a := range h.client.Actions() {
		if a.GetVerb() == "update" && a.GetResource().Resource == "zfssnapshots" {
			obj := a.(interface{ GetObject() runtime.Object }).GetObject()
			if name := obj.(*apis.ZFSSnapshot).Name; name == "snapshot-3" {
				t.Errorf("snapshot-3 was updated though its label was right")
			}
		}
	}
}

// An agent that died between the promote and the relabel finds the
// ZFSPromote InProgress and the volume without an origin; it relabels and
// finishes without promoting again
func TestPromoteAlreadyPromoted(t *testing.T) {
	z := &fakeZFS{
		origin:    "-",
		snapshots: map[string][]string{"pvc-c": {"snapshot-1"}},
	}
	h := newHarness(t, z,
		testVolume("pvc-c"),
		testSnapshot("snapshot-1", "pvc-t"),
		testPromote("pvc-c", testNode, apis.PromoteZFSStatusInProgress),
	)

	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}
	if len(z.promoted) != 0 {
		t.Fatalf("promoted again: %v", z.promoted)
	}
	if got := h.label("snapshot-1"); got != "pvc-c" {
		t.Fatalf("snapshot-1: want label pvc-c, got %s", got)
	}
	if got := h.status("pvc-c"); got != apis.PromoteZFSStatusDone {
		t.Fatalf("want Done, got %s", got)
	}
}

func TestPromoteFails(t *testing.T) {
	z := &fakeZFS{
		origin:     "pool/pvc-t@snapshot-1",
		promoteErr: errors.New("cannot promote 'pool/pvc-c': out of space"),
		snapshots:  map[string][]string{"pvc-c": {"snapshot-1"}},
	}
	h := newHarness(t, z,
		testVolume("pvc-c"),
		testSnapshot("snapshot-1", "pvc-t"),
		testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit),
	)

	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}
	if got := h.status("pvc-c"); got != apis.PromoteZFSStatusFailed {
		t.Fatalf("want Failed, got %s", got)
	}
	if got := h.label("snapshot-1"); got != "pvc-t" {
		t.Fatalf("snapshot-1 relabelled after a failed promote: %s", got)
	}
	ev := h.events()
	if len(ev) != 1 || !strings.HasPrefix(ev[0], "Warning "+zfs.ReasonPromoteFailed) ||
		!strings.Contains(ev[0], "out of space") {
		t.Fatalf("events: want one %s with the zfs error, got %q", zfs.ReasonPromoteFailed, ev)
	}

	// Failed is final: another sync does nothing
	z.promoteErr = nil
	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}
	if len(z.promoted) != 0 || h.status("pvc-c") != apis.PromoteZFSStatusFailed {
		t.Fatalf("a Failed promote was retried")
	}
}

func TestPromoteMissingVolume(t *testing.T) {
	h := newHarness(t, &fakeZFS{origin: "pool/pvc-t@snapshot-1"},
		testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit),
	)
	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}
	if got := h.status("pvc-c"); got != apis.PromoteZFSStatusFailed {
		t.Fatalf("want Failed, got %s", got)
	}
	if ev := h.events(); len(ev) != 1 || !strings.Contains(ev[0], "pvc-c") {
		t.Fatalf("events: want one naming the volume, got %q", ev)
	}
}

// A volume on another node than the ZFSPromote claims is refused, as this
// node's zfs does not have it
func TestPromoteVolumeOnOtherNode(t *testing.T) {
	z := &fakeZFS{origin: "pool/pvc-t@snapshot-1"}
	vol := testVolume("pvc-c")
	vol.Spec.OwnerNodeID = "node-2"
	h := newHarness(t, z, vol, testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit))
	if err := h.sync("pvc-c"); err != nil {
		t.Fatal(err)
	}
	if len(z.promoted) != 0 || h.status("pvc-c") != apis.PromoteZFSStatusFailed {
		t.Fatalf("want Failed without a promote, got %v %s", z.promoted, h.status("pvc-c"))
	}
}

func TestPromoteIgnored(t *testing.T) {
	deleting := testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit)
	deleting.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	deleting.Finalizers = []string{"test"}

	tests := map[string]*apis.ZFSPromote{
		"another node": testPromote("pvc-c", "node-2", apis.PromoteZFSStatusInit),
		"done":         testPromote("pvc-c", testNode, apis.PromoteZFSStatusDone),
		"failed":       testPromote("pvc-c", testNode, apis.PromoteZFSStatusFailed),
		"pending":      testPromote("pvc-c", testNode, apis.PromoteZFSStatusPending),
		"deleting":     deleting,
	}
	for name, p := range tests {
		t.Run(name, func(t *testing.T) {
			z := &fakeZFS{origin: "pool/pvc-t@snapshot-1"}
			h := newHarness(t, z, testVolume("pvc-c"), p)

			h.ctrl.enqueuePromote(p)
			if n := h.ctrl.workqueue.Len(); n != 0 {
				t.Fatalf("enqueued %d items", n)
			}
			if err := h.sync("pvc-c"); err != nil {
				t.Fatal(err)
			}
			if len(z.promoted) != 0 || len(z.listed) != 0 || len(h.statusWrites()) != 0 {
				t.Fatalf("acted on it: promoted %v listed %v writes %v", z.promoted, z.listed, h.statusWrites())
			}
		})
	}
}

func TestEnqueuePending(t *testing.T) {
	for _, status := range []apis.ZFSPromoteStatus{apis.PromoteZFSStatusInit, apis.PromoteZFSStatusInProgress} {
		h := newHarness(t, &fakeZFS{})
		h.ctrl.enqueuePromote(testPromote("pvc-c", testNode, status))
		if n := h.ctrl.workqueue.Len(); n != 1 {
			t.Fatalf("%s: want 1 item queued, got %d", status, n)
		}
		key, _ := h.ctrl.workqueue.Get()
		if key != testNamespace+"/pvc-c" {
			t.Fatalf("%s: want key %s/pvc-c, got %s", status, testNamespace, key)
		}
		h.ctrl.workqueue.Done(key)
	}
}

// The worker path: the lister finds the ZFSPromote and the controller
// carries it to Done
func TestSyncHandler(t *testing.T) {
	z := &fakeZFS{origin: "pool/pvc-t@snapshot-1", snapshots: map[string][]string{}}
	p := testPromote("pvc-c", testNode, apis.PromoteZFSStatusInit)
	h := newHarness(t, z, testVolume("pvc-c"), p)
	informer := informers.NewSharedInformerFactory(h.client, 0).Zfs().V1().ZFSPromotes()
	if err := informer.Informer().GetIndexer().Add(p); err != nil {
		t.Fatal(err)
	}
	h.ctrl.promoteLister = informer.Lister()

	if err := h.ctrl.syncHandler(testNamespace + "/pvc-c"); err != nil {
		t.Fatal(err)
	}
	if got := h.status("pvc-c"); got != apis.PromoteZFSStatusDone {
		t.Fatalf("want Done, got %s", got)
	}
	if err := h.ctrl.syncHandler(testNamespace + "/gone"); err != nil {
		t.Fatalf("a deleted ZFSPromote must not be retried: %v", err)
	}
}
