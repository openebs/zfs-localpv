/*
Copyright 2020 The OpenEBS Authors

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

package driver

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/nodebuilder"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given pool
	VolumeWeighted = "VolumeWeighted"

	// pick the node where the matching pools have occupied less capacity
	// this will be the default scheduler when none provided
	CapacityWeighted = "CapacityWeighted"

	// pick the node whose matching pool has the most free space
	SpaceWeighted = "SpaceWeighted"
)

// Returns the name of the kubernetes node this ZFSNode belongs to,
// read from the owner reference the node agent maintains and falling back to the
// CR name.
func k8sNodeName(node apis.ZFSNode) string {
	for _, ref := range node.OwnerReferences {
		if ref.Kind == "Node" {
			return ref.Name
		}
	}
	return node.Name
}

// Returns the pool part of a "poolname" parameter, which may name a
// pool ("zpool") or a dataset below it ("zpool/k8s/localpv").
func poolRoot(poolname string) string {
	return strings.SplitN(poolname, "/", 2)[0]
}

// Folds the "poolname" and "poolpattern" storageclass
// parameters into one regular expression matched against pool roots. Exactly one
// of them must be set: poolpattern compiles as given, poolname as an anchored
// exact match.
func compilePoolPattern(poolname, poolpattern string) (*regexp.Regexp, error) {
	switch {
	case poolname != "" && poolpattern != "":
		return nil, errors.New("poolname and poolpattern are mutually exclusive, set only one of them")
	case poolpattern != "":
		pattern, err := regexp.Compile(poolpattern)
		if err != nil {
			return nil, fmt.Errorf("invalid poolpattern %q : %w", poolpattern, err)
		}
		return pattern, nil
	case poolname != "":
		// quoting is required for correctness: legal ZFS pool names contain
		// regex metacharacters, so an unescaped "tank.prod" would also match
		// "tankXprod" and select the wrong pool
		return regexp.MustCompile("^" + regexp.QuoteMeta(poolRoot(poolname)) + "$"), nil
	}
	return nil, errors.New("either poolname or poolpattern must be set")
}

// Reports whether the volume gets a ZFS reservation and so has to
// fit in a pool's free capacity at create time. A zvol reserves unless it is
// created sparse; a dataset carries a quota, which is a limit and not a
// reservation, and reserves only when thinprovision is "no".
func reservesSpace(vtype, thinProvision string) bool {
	if thinProvision == "no" {
		return true
	}
	return vtype != zfs.VolTypeDataset && thinProvision != "yes"
}

// Returns the ZFSNode CRs, which advertise the pools present on
// every node along with their free and used capacity.
func listZFSNodes() ([]apis.ZFSNode, error) {
	nodelist, err := nodebuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return nodelist.Items, nil
}

// Returns the node name to volume count mapping for the
// volumes provisioned into a pool matching the pattern.
func getVolumeWeightedMap(pattern *regexp.Regexp) (map[string]int64, error) {
	zvlist, err := volbuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return map[string]int64{}, err
	}

	// a volume records the node id of its node, the ZFSNode CRs are what maps
	// that back to the node name the scheduler works with
	nodelist, err := listZFSNodes()
	if err != nil {
		return map[string]int64{}, err
	}

	return volumeWeightedMap(zvlist.Items, nodelist, pattern), nil
}

// Counts the volumes in a matching pool per node, translating
// the node id a volume records back to its node name.
func volumeWeightedMap(zvlist []apis.ZFSVolume, nodelist []apis.ZFSNode, pattern *regexp.Regexp) map[string]int64 {
	nodename := map[string]string{}
	for _, node := range nodelist {
		nodename[node.Name] = k8sNodeName(node)
	}

	nmap := map[string]int64{}

	// create the map of the volume count for the matching pools
	for _, zv := range zvlist {
		if !pattern.MatchString(poolRoot(zv.Spec.PoolName)) {
			continue
		}
		node, ok := nodename[zv.Spec.OwnerNodeID]
		if !ok {
			// no ZFSNode CR for the volume's node, its node id is the best
			// guess at the node name
			node = zv.Spec.OwnerNodeID
		}
		nmap[node]++
	}

	return nmap
}

// Returns the node name to used capacity mapping for the
// pools matching the pattern, as the ZFSNode CR reports it, so it also accounts
// for data which was not provisioned by this driver.
func getCapacityWeightedMap(pattern *regexp.Regexp) (map[string]int64, error) {
	nodelist, err := listZFSNodes()
	if err != nil {
		return map[string]int64{}, err
	}

	return capacityWeightedMap(nodelist, pattern), nil
}

// Sums the used capacity of a node's matching pools.
func capacityWeightedMap(nodelist []apis.ZFSNode, pattern *regexp.Regexp) map[string]int64 {
	nmap := map[string]int64{}

	for _, node := range nodelist {
		for _, pool := range node.Pools {
			if pattern.MatchString(pool.Name) {
				nmap[k8sNodeName(node)] += pool.Used.Value()
			}
		}
	}

	return nmap
}

// Returns the node name to weight mapping which orders the
// nodes by the free capacity of their roomiest matching pool, so that the volume
// goes to the node with the most room left rather than the least written to.
func getSpaceWeightedMap(pattern *regexp.Regexp) (map[string]int64, error) {
	nodelist, err := listZFSNodes()
	if err != nil {
		return map[string]int64{}, err
	}

	return spaceWeightedMap(nodelist, pattern), nil
}

// Inverts the free capacity of each node's roomiest matching
// pool into a weight.
func spaceWeightedMap(nodelist []apis.ZFSNode, pattern *regexp.Regexp) map[string]int64 {
	nmap := map[string]int64{}

	for _, node := range nodelist {
		pool, free := maxFreePool(node.Pools, pattern)
		if pool == "" {
			continue
		}
		// the scheduler prefers the node with the lowest weight, so the free
		// capacity is inverted: the more space a node has left, the less loaded
		// it looks. A node whose matching pool is full keeps its entry, at
		// math.MaxInt64, so that it sorts last instead of being front loaded.
		nmap[k8sNodeName(node)] = math.MaxInt64 - free
	}

	return nmap
}

// Returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, pattern *regexp.Regexp) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(pattern)
	case CapacityWeighted:
		return getCapacityWeightedMap(pattern)
	case SpaceWeighted:
		return getSpaceWeightedMap(pattern)
	}
	// return CapacityWeighted(default) if not specified
	return getCapacityWeightedMap(pattern)
}

// Returns the names of the nodes whose roomiest matching pool
// has more than `size` bytes free, and whether any pool matched the pattern at
// all. The fit is best effort, as the free capacity on the ZFSNode CR is a
// periodic snapshot, so "zfs create" stays the final arbiter.
func getSuitableNodes(pattern *regexp.Regexp, size int64) (map[string]bool, bool, error) {
	nodelist, err := listZFSNodes()
	if err != nil {
		return nil, false, err
	}

	suitable, matched := suitableNodes(nodelist, pattern, size)
	return suitable, matched, nil
}

// Reports which nodes have a matching pool with more than `size`
// bytes free, and whether any pool matched the pattern at all. A single pool has
// to hold the whole reservation, so the free capacity is not summed across the
// pools of a node.
func suitableNodes(nodelist []apis.ZFSNode, pattern *regexp.Regexp, size int64) (map[string]bool, bool) {
	suitable := map[string]bool{}
	matched := false

	for _, node := range nodelist {
		pool, free := maxFreePool(node.Pools, pattern)
		if pool == "" {
			continue
		}
		matched = true
		if free > size {
			suitable[k8sNodeName(node)] = true
		}
	}

	return suitable, matched
}

// Returns the matching pool with the most free capacity on the given
// node, or an empty string when no pool on the node matches, turning a
// poolpattern into the concrete pool stored in ZFSVolume.Spec.PoolName. `nodeid`
// is the node id, since the ZFSNode CR is named after it.
func resolvePool(nodeid string, pattern *regexp.Regexp) (string, error) {
	zfsNode, err := nodebuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		Get(nodeid, metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	pool, _ := maxFreePool(zfsNode.Pools, pattern)
	return pool, nil
}

// Returns the pool matching the pattern with the most free capacity
// along with that capacity, or an empty name when no pool matches. It is the one
// definition of "the node's roomiest matching pool", shared by resolvePool,
// suitableNodes and spaceWeightedMap.
func maxFreePool(pools []apis.Pool, pattern *regexp.Regexp) (string, int64) {
	var (
		selected string
		maxFree  int64
	)

	for _, pool := range pools {
		if !pattern.MatchString(pool.Name) {
			continue
		}
		if free := pool.Free.Value(); selected == "" || free > maxFree {
			selected, maxFree = pool.Name, free
		}
	}

	return selected, maxFree
}
