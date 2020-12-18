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
	"math"

	"github.com/container-storage-interface/spec/lib/go/csi"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// getNodeList gets the nodelist which satisfies the topology info
func getNodeList(topo *csi.TopologyRequirement) ([]string, error) {

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
			if !nodeFiltered {
				nodelist = append(nodelist, node.Name)
				break
			}
		}
	}

	return nodelist, nil
}

// runScheduler goes through the node mapping
// in the topology and picks the node which is less weighted
func runScheduler(nodelist []string, nmap map[string]int64) string {
	var selected string

	var weight int64 = math.MaxInt64

	// schedule it on the node which has less weight
	for _, node := range nodelist {
		if nmap[node] < weight {
			selected = node
			weight = nmap[node]
		}
	}
	return selected
}

// Scheduler schedules the PV as per topology constraints for
// the given node weight.
func Scheduler(req *csi.CreateVolumeRequest, nmap map[string]int64) string {
	topo := req.AccessibilityRequirements
	if topo == nil ||
		len(topo.Preferred) == 0 {
		klog.Errorf("scheduler: topology information not provided")
		return ""
	}

	nodelist, err := getNodeList(topo)
	if err != nil {
		klog.Errorf("scheduler: can not get the nodelist err : %v", err.Error())
		return ""
	} else if len(nodelist) == 0 {
		klog.Errorf("scheduler: nodelist is empty")
		return ""
	}

	// if there is a single node, schedule it on that
	if len(nodelist) == 1 {
		return nodelist[0]
	}

	return runScheduler(nodelist, nmap)
}
