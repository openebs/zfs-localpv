package driver

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// Builds a ZFSNode pool entry with the given free/used capacity in bytes.
func newPool(name string, free, used int64) apis.Pool {
	return apis.Pool{
		Name: name,
		UUID: name + "-uuid",
		Free: *resource.NewQuantity(free, resource.BinarySI),
		Used: *resource.NewQuantity(used, resource.BinarySI),
	}
}

// Builds a ZFSNode CR as the node agent does without a custom node
// id: the CR is named after the node and owned by it, so name == id.
func newZFSNode(name string, pools ...apis.Pool) apis.ZFSNode {
	return newCustomIDZFSNode(name, name, pools...)
}

// Builds a ZFSNode CR for a node running with a custom node
// id: the CR is named after the id, while the owner reference still names the
// kubernetes node, which is what the scheduler works with.
func newCustomIDZFSNode(nodeid, nodename string, pools ...apis.Pool) apis.ZFSNode {
	return apis.ZFSNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeid,
			Namespace: "openebs",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1",
				Kind:       "Node",
				Name:       nodename,
			}},
		},
		Pools: pools,
	}
}

func newZFSVolume(node, poolname string) apis.ZFSVolume {
	return apis.ZFSVolume{
		Spec: apis.VolumeInfo{
			OwnerNodeID: node,
			PoolName:    poolname,
		},
	}
}

func TestPoolRoot(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"pool name":            {input: "zfspv-pool", expected: "zfspv-pool"},
		"dataset path":         {input: "zpool/k8s/localpv", expected: "zpool"},
		"single child dataset": {input: "zpool/localpv", expected: "zpool"},
		"empty":                {input: "", expected: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, poolRoot(test.input))
		})
	}
}

func TestCompilePoolPattern(t *testing.T) {
	tests := map[string]struct {
		poolname    string
		poolpattern string
		matches     []string
		notMatches  []string
		wantErr     bool
	}{
		"exact poolname": {
			poolname:   "zfspv-pool",
			matches:    []string{"zfspv-pool"},
			notMatches: []string{"zfspv-pool-a", "my-zfspv-pool", "zfspv"},
		},
		"poolname metacharacters are quoted": {
			poolname:   "tank.prod",
			matches:    []string{"tank.prod"},
			notMatches: []string{"tankXprod", "tank-prod"},
		},
		"poolname is matched against its root": {
			poolname:   "zpool/k8s/localpv",
			matches:    []string{"zpool"},
			notMatches: []string{"zpool/k8s/localpv", "zpool2"},
		},
		"poolpattern is unanchored": {
			poolpattern: "zfspv-pool",
			matches:     []string{"zfspv-pool", "zfspv-pool-a", "my-zfspv-pool"},
			notMatches:  []string{"zfspv"},
		},
		"poolpattern anchored by the user": {
			poolpattern: "^zfspv-pool-[ab]$",
			matches:     []string{"zfspv-pool-a", "zfspv-pool-b"},
			notMatches:  []string{"zfspv-pool", "zfspv-pool-c"},
		},
		"both set is rejected":    {poolname: "zfspv-pool", poolpattern: "zfspv-pool.*", wantErr: true},
		"neither set is rejected": {wantErr: true},
		"invalid regex":           {poolpattern: "zfspv-pool[", wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pattern, err := compilePoolPattern(test.poolname, test.poolpattern)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, p := range test.matches {
				assert.True(t, pattern.MatchString(p), "%q should match %s", p, pattern)
			}
			for _, p := range test.notMatches {
				assert.False(t, pattern.MatchString(p), "%q should not match %s", p, pattern)
			}
		})
	}
}

func TestReservesSpace(t *testing.T) {
	zvol := zfs.VolTypeZVol
	dataset := zfs.VolTypeDataset

	tests := map[string]struct {
		vtype         string
		thinProvision string
		expected      bool
	}{
		// a zvol gets ZFS's default refreservation unless it is sparse
		"zvol, thinprovision unset": {vtype: zvol, thinProvision: "", expected: true},
		"zvol, thinprovision no":    {vtype: zvol, thinProvision: "no", expected: true},
		"zvol, thinprovision yes":   {vtype: zvol, thinProvision: "yes", expected: false},
		// a dataset only gets a reservation on top of its quota when asked for
		"dataset, thinprovision unset": {vtype: dataset, thinProvision: "", expected: false},
		"dataset, thinprovision no":    {vtype: dataset, thinProvision: "no", expected: true},
		"dataset, thinprovision yes":   {vtype: dataset, thinProvision: "yes", expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, reservesSpace(test.vtype, test.thinProvision))
		})
	}
}

func TestK8sNodeName(t *testing.T) {
	tests := map[string]struct {
		node     apis.ZFSNode
		expected string
	}{
		"no custom node id, name == id": {
			node:     newZFSNode("node1"),
			expected: "node1",
		},
		"custom node id, the owner reference names the node": {
			node:     newCustomIDZFSNode("node1-custom-id", "node1"),
			expected: "node1",
		},
		"no owner reference falls back to the CR name": {
			node:     apis.ZFSNode{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
			expected: "node1",
		},
		"a non node owner reference is ignored": {
			node: apis.ZFSNode{ObjectMeta: metav1.ObjectMeta{
				Name:            "node1-custom-id",
				OwnerReferences: []metav1.OwnerReference{{Kind: "Pod", Name: "some-pod"}},
			}},
			expected: "node1-custom-id",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, k8sNodeName(test.node))
		})
	}
}

func TestVolumeWeightedMap(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newZFSNode("node1"),
		newZFSNode("node2"),
		newZFSNode("node3"),
	}

	zvlist := []apis.ZFSVolume{
		newZFSVolume("node1", "zfspv-pool-a"),
		newZFSVolume("node1", "zfspv-pool-a"),
		newZFSVolume("node1", "other-pool"),
		newZFSVolume("node2", "zfspv-pool-b"),
		// dataset paths are counted against their pool root
		newZFSVolume("node2", "zfspv-pool-b/k8s/localpv"),
		newZFSVolume("node3", "other-pool"),
	}

	tests := map[string]struct {
		pattern  string
		expected map[string]int64
	}{
		"pattern spanning both pools": {
			pattern:  "^zfspv-pool-.*",
			expected: map[string]int64{"node1": 2, "node2": 2},
		},
		"exact pool": {
			pattern:  "^zfspv-pool-a$",
			expected: map[string]int64{"node1": 1 + 1},
		},
		"no match": {
			pattern:  "^nosuch$",
			expected: map[string]int64{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			nmap := volumeWeightedMap(zvlist, nodelist, regexp.MustCompile(test.pattern))
			assert.Equal(t, test.expected, nmap)
		})
	}
}

// the volume count has to land under the node NAME, since that is what lib-csi's
// scheduler looks up, while a volume records the node ID of its node
func TestVolumeWeightedMapCustomNodeID(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newCustomIDZFSNode("node1-custom-id", "node1"),
		newCustomIDZFSNode("node2-custom-id", "node2"),
	}
	zvlist := []apis.ZFSVolume{
		newZFSVolume("node1-custom-id", "zfspv-pool-a"),
		newZFSVolume("node1-custom-id", "zfspv-pool-a"),
		newZFSVolume("node2-custom-id", "zfspv-pool-b"),
		// a volume whose node has no ZFSNode CR falls back to its node id
		newZFSVolume("node3-custom-id", "zfspv-pool-c"),
	}

	nmap := volumeWeightedMap(zvlist, nodelist, regexp.MustCompile("^zfspv-pool-"))

	assert.Equal(t, map[string]int64{
		"node1":           2,
		"node2":           1,
		"node3-custom-id": 1,
	}, nmap)
}

func TestCapacityWeightedMap(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newZFSNode("node1", newPool("zfspv-pool-a", 90*Gi, 10*Gi), newPool("other-pool", 50*Gi, 50*Gi)),
		newZFSNode("node2", newPool("zfspv-pool-b", 70*Gi, 30*Gi), newPool("zfspv-pool-c", 95*Gi, 5*Gi)),
		// a node with a matching but completely empty pool still gets an entry,
		// otherwise the scheduler would move it to the front of the list
		newZFSNode("node3", newPool("zfspv-pool-d", 100*Gi, 0)),
		newZFSNode("node4", newPool("other-pool", 100*Gi, 0)),
	}

	tests := map[string]struct {
		pattern  string
		expected map[string]int64
	}{
		"used capacity is summed over the matching pools": {
			pattern: "^zfspv-pool-",
			expected: map[string]int64{
				"node1": 10 * Gi,
				"node2": 30*Gi + 5*Gi,
				"node3": 0,
			},
		},
		"only the matching pool counts": {
			pattern:  "^other-pool$",
			expected: map[string]int64{"node1": 50 * Gi, "node4": 0},
		},
		"no match": {
			pattern:  "^nosuch$",
			expected: map[string]int64{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			nmap := capacityWeightedMap(nodelist, regexp.MustCompile(test.pattern))
			assert.Equal(t, test.expected, nmap)
		})
	}
}

// with a custom node id the ZFSNode CR is named after the id, but the weight
// still has to be keyed by the node name, otherwise lib-csi finds no key, treats
// every node as least loaded and the weighting silently stops working
func TestCapacityWeightedMapCustomNodeID(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newCustomIDZFSNode("node1-custom-id", "node1", newPool("zfspv-pool-a", 90*Gi, 10*Gi)),
		newCustomIDZFSNode("node2-custom-id", "node2", newPool("zfspv-pool-b", 70*Gi, 30*Gi)),
	}

	nmap := capacityWeightedMap(nodelist, regexp.MustCompile("^zfspv-pool-"))

	assert.Equal(t, map[string]int64{"node1": 10 * Gi, "node2": 30 * Gi}, nmap)
}

func TestSuitableNodes(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newZFSNode("node1", newPool("zfspv-pool-a", 10*Gi, 90*Gi)),
		// free capacity is not summed across the pools of a node, a single pool
		// has to hold the whole reservation
		newZFSNode("node2", newPool("zfspv-pool-b", 15*Gi, 85*Gi), newPool("zfspv-pool-c", 15*Gi, 85*Gi)),
		newZFSNode("node3", newPool("zfspv-pool-d", 5*Gi, 95*Gi), newPool("zfspv-pool-e", 40*Gi, 60*Gi)),
		newZFSNode("node4", newPool("other-pool", 100*Gi, 0)),
	}

	tests := map[string]struct {
		pattern     string
		size        int64
		expected    map[string]bool
		wantMatched bool
	}{
		"largest free pool of a node decides": {
			pattern:     "^zfspv-pool-",
			size:        20 * Gi,
			expected:    map[string]bool{"node3": true},
			wantMatched: true,
		},
		"every node fits a small volume": {
			pattern:     "^zfspv-pool-",
			size:        1 * Gi,
			expected:    map[string]bool{"node1": true, "node2": true, "node3": true},
			wantMatched: true,
		},
		"a pool matches but nothing fits": {
			pattern:     "^zfspv-pool-",
			size:        100 * Gi,
			expected:    map[string]bool{},
			wantMatched: true,
		},
		"a volume of exactly the free capacity does not fit": {
			pattern:     "^zfspv-pool-a$",
			size:        10 * Gi,
			expected:    map[string]bool{},
			wantMatched: true,
		},
		"no pool matches the pattern": {
			pattern:     "^nosuch",
			size:        1 * Gi,
			expected:    map[string]bool{},
			wantMatched: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			suitable, matched := suitableNodes(nodelist, regexp.MustCompile(test.pattern), test.size)
			assert.Equal(t, test.expected, suitable)
			assert.Equal(t, test.wantMatched, matched)
		})
	}
}

// the suitable set is intersected with the scheduler's output, which holds node
// names, so a custom node id must not leak into it — keying it by the id would
// intersect to nothing and fail every space reserving volume
func TestSuitableNodesCustomNodeID(t *testing.T) {
	nodelist := []apis.ZFSNode{
		newCustomIDZFSNode("node1-custom-id", "node1", newPool("zfspv-pool-a", 40*Gi, 60*Gi)),
		newCustomIDZFSNode("node2-custom-id", "node2", newPool("zfspv-pool-b", 5*Gi, 95*Gi)),
	}

	suitable, matched := suitableNodes(nodelist, regexp.MustCompile("^zfspv-pool-"), 20*Gi)

	assert.Equal(t, map[string]bool{"node1": true}, suitable)
	assert.True(t, matched)
}

func TestPoolForNode(t *testing.T) {
	tests := map[string]struct {
		pools    []apis.Pool
		pattern  string
		expected string
	}{
		"the matching pool with the most free capacity wins": {
			pools:    []apis.Pool{newPool("zfspv-pool-a", 10*Gi, 90*Gi), newPool("zfspv-pool-b", 60*Gi, 40*Gi), newPool("zfspv-pool-c", 30*Gi, 70*Gi)},
			pattern:  "^zfspv-pool-",
			expected: "zfspv-pool-b",
		},
		"a bigger non matching pool is ignored": {
			pools:    []apis.Pool{newPool("zfspv-pool-a", 10*Gi, 90*Gi), newPool("other-pool", 90*Gi, 10*Gi)},
			pattern:  "^zfspv-pool-",
			expected: "zfspv-pool-a",
		},
		"a full matching pool is still selected": {
			pools:    []apis.Pool{newPool("zfspv-pool-a", 0, 100*Gi)},
			pattern:  "^zfspv-pool-",
			expected: "zfspv-pool-a",
		},
		"no pool matches": {
			pools:    []apis.Pool{newPool("other-pool", 90*Gi, 10*Gi)},
			pattern:  "^zfspv-pool-",
			expected: "",
		},
		"node without pools": {
			pools:    nil,
			pattern:  "^zfspv-pool-",
			expected: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, poolForNode(test.pools, regexp.MustCompile(test.pattern)))
		})
	}
}
