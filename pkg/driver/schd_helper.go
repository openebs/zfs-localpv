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
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given pool
	// this will be the default scheduler when none provided
	VolumeWeighted = "VolumeWeighted"
)

// getVolumeWeightedMap goes through all the pools on all the nodes
// and creats the node mapping of the volume for all the nodes.
// It returns a map which has nodes as key and volumes present
// on the nodes as corresponding value.
func getVolumeWeightedMap(pool string) (map[string]int64, error) {
	nmap := map[string]int64{}

	zvlist, err := volbuilder.NewKubeclient().
		WithNamespace(zfs.OpenEBSNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	// create the map of the volume count
	// for the given pool
	for _, zv := range zvlist.Items {
		if zv.Spec.PoolName == pool {
			nmap[zv.Spec.OwnerNodeID]++
		}
	}

	return nmap, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, pool string) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(pool)
	}
	// return VolumeWeighted(default) if not specified
	return getVolumeWeightedMap(pool)
}
