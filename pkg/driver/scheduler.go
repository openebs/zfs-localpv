/*
Copyright © 2019 The OpenEBS Authors

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
	"github.com/Sirupsen/logrus"
	"math"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/zfs-localpv/pkg/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given pool
	// this will be the default scheduler when none provided
	VolumeWeighted = "VolumeWeighted"
)

// volumeWeightedScheduler goes through all the pools on the nodes mentioned
// in the topology and picks the node which has less volume on
// the given zfs pool.
func volumeWeightedScheduler(topo *csi.TopologyRequirement, pool string) string {
	var selected string

	zvlist, err := builder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return ""
	}

	volmap := map[string]int{}

	// create the map of the volume count
	// for the given pool
	for _, zv := range zvlist.Items {
		if zv.Spec.PoolName == pool {
			volmap[zv.Spec.OwnerNodeID]++
		}
	}

	var numVol int = math.MaxInt32

	// schedule it on the node which has less
	// number of volume for the given pool
	for _, prf := range topo.Preferred {
		node := prf.Segments[zfs.ZFSTopologyKey]
		if volmap[node] < numVol {
			selected = node
			numVol = volmap[node]
		}
	}
	return selected
}

// scheduler schedules the PV as per topology constraints for
// the given zfs pool.
func scheduler(topo *csi.TopologyRequirement, schld string, pool string) string {

	if topo == nil ||
		len(topo.Preferred) == 0 {
		logrus.Errorf("topology information not provided")
		return ""
	}
	// if there is a single node, schedule it on that
	if len(topo.Preferred) == 1 {
		return topo.Preferred[0].Segments[zfs.ZFSTopologyKey]
	}

	switch schld {
	case VolumeWeighted:
		return volumeWeightedScheduler(topo, pool)
	default:
		return volumeWeightedScheduler(topo, pool)
	}

	return ""
}
