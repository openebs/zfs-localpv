package promotebuilder

import (
	"testing"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
)

func TestBuild(t *testing.T) {
	p, err := NewBuilder().WithName("pvc-c").WithNamespace("openebs").
		WithVolume("pvc-c").WithNode("node-1").Build()
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "pvc-c" || p.Namespace != "openebs" || p.Spec.VolumeName != "pvc-c" ||
		p.Spec.OwnerNodeID != "node-1" || p.Status != apis.PromoteZFSStatusInit {
		t.Fatalf("unexpected object %+v", p)
	}

	if _, err := NewBuilder().WithName("pvc-c").WithVolume("").Build(); err == nil {
		t.Fatal("want an error for a missing volume name")
	}
	if _, err := BuildFrom(nil).Build(); err == nil {
		t.Fatal("want an error for a nil promote")
	}

	done, err := BuildFrom(p).WithStatus(apis.PromoteZFSStatusDone).Build()
	if err != nil || done.Status != apis.PromoteZFSStatusDone || p.Status != apis.PromoteZFSStatusInit {
		t.Fatalf("BuildFrom must copy: got %v %v, original %v", done, err, p.Status)
	}
}
