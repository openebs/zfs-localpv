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
)

// The node keys used by all the helpers below are the node IDs (the value of
// the openebs.io/nodeid topology label), which is what ZFSVolume records in
// Spec.OwnerNodeID and what the ZFSNode CR is named after. Note that lib-csi's
// scheduler works with kubernetes node *names*, so the caller has to map a node
// name to its node ID (zfs.GetNodeID) before looking it up in the suitable set
// or asking resolvePool for the concrete pool.

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

// getVolumeWeightedMap goes through all the volumes provisioned into a pool
// matching the pattern and creates the node mapping of the volume count.
// It returns a map which has nodes as key and volumes present
// on the nodes as corresponding value.
//
// This is an ordering input for the scheduler only, it does not decide whether
// a node can hold the volume, see getSuitableNodes for that.
func getVolumeWeightedMap(pattern *regexp.Regexp) (map[string]int64, error) {
	zvlist, err := volbuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return map[string]int64{}, err
	}

	return volumeWeightedMap(zvlist.Items, pattern), nil
}

func volumeWeightedMap(zvlist []apis.ZFSVolume, pattern *regexp.Regexp) map[string]int64 {
	nmap := map[string]int64{}

	// create the map of the volume count for the matching pools
	for _, zv := range zvlist {
		if pattern.MatchString(poolRoot(zv.Spec.PoolName)) {
			nmap[zv.Spec.OwnerNodeID]++
		}
	}

	return nmap
}

// getCapacityWeightedMap goes through the pools advertised by every node and
// creates the node mapping of the capacity used by the pools matching the
// pattern. It returns a map which has nodes as key and the used capacity of
// the matching pools as corresponding value. The scheduler will use this map
// and picks the node which is less weighted.
//
// The weight is the pool's real on disk usage as reported by the ZFSNode CR,
// so it also accounts for the data which was not provisioned by this driver.
// Every node with a matching pool gets an entry, even when the pools are empty,
// so that the ordering is not distorted: lib-csi's scheduler moves the nodes
// missing from the map to the front of the list.
func getCapacityWeightedMap(pattern *regexp.Regexp) (map[string]int64, error) {
	nodelist, err := listZFSNodes()
	if err != nil {
		return map[string]int64{}, err
	}

	return capacityWeightedMap(nodelist, pattern), nil
}

func capacityWeightedMap(nodelist []apis.ZFSNode, pattern *regexp.Regexp) map[string]int64 {
	nmap := map[string]int64{}

	for _, node := range nodelist {
		for _, pool := range node.Pools {
			if pattern.MatchString(pool.Name) {
				nmap[node.Name] += pool.Used.Value()
			}
		}
	}

	return nmap
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, pattern *regexp.Regexp) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(pattern)
	case CapacityWeighted:
		return getCapacityWeightedMap(pattern)
	}
	// return CapacityWeighted(default) if not specified
	return getCapacityWeightedMap(pattern)
}

// getSuitableNodes returns the set of nodes which have a pool matching the
// pattern with more than `size` bytes free, along with `matched`, which tells
// whether any pool matched the pattern at all, the fit aside.
//
// A single pool has to hold the whole reservation, the free capacity is not
// summed across the pools of a node. The caller intersects this set with the
// node list returned by the scheduler for the volumes which reserve space (see
// reservesSpace), and uses `matched` to tell an exhausted pool (matched, but
// nothing fits) apart from a storageclass which names a pool that does not
// exist anywhere (not matched).
//
// The check is best effort: the free capacity on the ZFSNode CR is a periodic
// snapshot and two concurrent creates can both pass it, so "zfs create" stays
// the final arbiter.
func getSuitableNodes(pattern *regexp.Regexp, size int64) (map[string]bool, bool, error) {
	nodelist, err := listZFSNodes()
	if err != nil {
		return nil, false, err
	}

	suitable, matched := suitableNodes(nodelist, pattern, size)
	return suitable, matched, nil
}

func suitableNodes(nodelist []apis.ZFSNode, pattern *regexp.Regexp, size int64) (map[string]bool, bool) {
	suitable := map[string]bool{}
	matched := false

	for _, node := range nodelist {
		var maxFree int64
		for _, pool := range node.Pools {
			if !pattern.MatchString(pool.Name) {
				continue
			}
			matched = true
			if free := pool.Free.Value(); free > maxFree {
				maxFree = free
			}
		}
		if maxFree > size {
			suitable[node.Name] = true
		}
	}

	return suitable, matched
}

// resolvePool returns the pool matching the pattern on the given node which has
// the most free capacity, or an empty string when no pool on the node matches.
//
// This is how a poolpattern is turned into the concrete pool stored in
// ZFSVolume.Spec.PoolName. It is not used for the fixed poolname case, where
// the parameter is used as it is, since it can carry a dataset path
// ("zpool/k8s/localpv") which the ZFSNode CR does not advertise.
func resolvePool(node string, pattern *regexp.Regexp) (string, error) {
	zfsNode, err := nodebuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		Get(node, metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	return poolForNode(zfsNode.Pools, pattern), nil
}

func poolForNode(pools []apis.Pool, pattern *regexp.Regexp) string {
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

	return selected
}
