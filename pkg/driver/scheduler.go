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
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	k8sapi "github.com/openebs/zfs-localpv/pkg/client/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given pool
	// this will be the default scheduler when none provided
	VolumeWeighted = "VolumeWeighted"
)

// GetNodeList gets the nodelist which satisfies the topology info
func GetNodeList(topo *csi.TopologyRequirement) ([]string, error) {

	var nodelist []string

	list, err := k8sapi.ListNodes(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range list.Items {
		for _, prf := range topo.Preferred {
			nodeFiltered := false
			for key, value := range prf.Segments {
				if node.Labels[key] != value {
					nodeFiltered = true
					break
				}
			}
			if nodeFiltered == false {
				nodelist = append(nodelist, node.Name)
				break
			}
		}
	}

	return nodelist, nil
}

// volumeWeightedScheduler goes through all the pools on the nodes mentioned
// in the topology and picks the node which has less volume on
// the given zfs pool.
func volumeWeightedScheduler(nodelist []string, pool string) string {
	var selected string

	zvlist, err := volbuilder.NewKubeclient().
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
	for _, node := range nodelist {
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

	nodelist, err := GetNodeList(topo)
	if err != nil {
		logrus.Errorf("can not get the ndelist err : %v", err.Error())
		return ""
	} else if len(nodelist) == 0 {
		logrus.Errorf("nodelist is empty")
		return ""
	}

	// if there is a single node, schedule it on that
	if len(nodelist) == 1 {
		return nodelist[0]
	}

	switch schld {
	case VolumeWeighted:
		return volumeWeightedScheduler(nodelist, pool)
	default:
		return volumeWeightedScheduler(nodelist, pool)
	}

	return ""
}
