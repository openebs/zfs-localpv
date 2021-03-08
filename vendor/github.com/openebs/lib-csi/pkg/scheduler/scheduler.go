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

package scheduler

import (
	"sort"

	"github.com/container-storage-interface/spec/lib/go/csi"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// key value struct for creating the filtered list
type kv struct {
	Key   string
	Value int64
}

// getNodeList gets the nodelist which satisfies the topology info
func getNodeList(topo []*csi.Topology) ([]string, error) {

	var nodelist []string

	list, err := k8sapi.ListNodes(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range list.Items {
		for _, prf := range topo {
			nodeFiltered := false
			for key, value := range prf.Segments {
				if node.Labels[key] != value {
					nodeFiltered = true
					break
				}
			}
			if !nodeFiltered {
				nodelist = append(nodelist, node.Name)
				break
			}
		}
	}

	return nodelist, nil
}

// runScheduler goes through the node mapping in the topology
// and creates the list of preferred nodes as per their weight
func runScheduler(nodelist []string, nmap map[string]int64) []string {
	var preferred []string
	var fmap []kv

	// go though the filtered node and prepare the preferred list
	for _, node := range nodelist {
		if val, ok := nmap[node]; ok {
			// create the filtered node map
			fmap = append(fmap, kv{node, val})
		} else {
			// put the non occupied nodes in beginning of the list
			preferred = append(preferred, node)
		}
	}

	// sort the filtered node map
	sort.Slice(fmap, func(i, j int) bool {
		return fmap[i].Value < fmap[j].Value
	})

	// put the occupied nodes in the sorted order at the end
	for _, kv := range fmap {
		preferred = append(preferred, kv.Key)
	}

	return preferred
}

// Scheduler schedules the PV as per topology constraints for
// the given node weight.
func Scheduler(req *csi.CreateVolumeRequest, nmap map[string]int64) []string {
	var nodelist []string
	areq := req.AccessibilityRequirements

	if areq == nil {
		klog.Errorf("scheduler: Accessibility Requirements not provided")
		return nodelist
	}

	topo := areq.Preferred
	if len(topo) == 0 {
		// if preferred list is empty, use the requisite
		topo = areq.Requisite
	}

	if len(topo) == 0 {
		klog.Errorf("scheduler: topology information not provided")
		return nodelist
	}

	nodelist, err := getNodeList(topo)
	if err != nil {
		klog.Errorf("scheduler: can not get the nodelist err : %v", err.Error())
		return nodelist
	} else if len(nodelist) == 0 {
		klog.Errorf("scheduler: nodelist is empty")
		return nodelist
	}

	// if there is a single node, schedule it on that
	if len(nodelist) == 1 {
		return nodelist
	}

	return runScheduler(nodelist, nmap)
}
