/*
 Copyright © 2021 The OpenEBS Authors

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

package zfsnode

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"sync"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	informers "github.com/openebs/zfs-localpv/pkg/generated/informer/externalversions"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// Start starts the zfsnode controller.
func Start(controllerMtx *sync.RWMutex, stopCh <-chan struct{}) error {

	// Get in cluster config
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}

	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}

	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building openebs clientset")
	}

	// setup watch only on node we are interested in.
	nodeInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		openebsClient, 0, informers.WithNamespace(zfs.OpenEBSNamespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", zfs.NodeID).String()
		}))

	nodeName := os.Getenv("OPENEBS_NODE_NAME")
	var k8sNode v1.Node

	if len(strings.TrimSpace(zfs.NodeID)) == 0 || nodeName == zfs.NodeID {
		k8sNodeCandidate, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), zfs.NodeID, metav1.GetOptions{})

		if err != nil {
			return errors.Wrapf(err, "fetch k8s node %s", zfs.NodeID)
		}

		k8sNode = *k8sNodeCandidate

	} else {
		topologyRequirement, requirementError := labels.NewRequirement(zfs.ZFSTopologyKey, selection.Equals, []string{zfs.NodeID})
		if requirementError != nil {
			return errors.Wrapf(requirementError, "Unable to generate topology requirement by %s for node id %s", zfs.ZFSTopologyKey, zfs.NodeID)
		}
		topologySelector := labels.NewSelector().Add(*topologyRequirement).String()
		klog.Infof("The topology selector is %s", topologySelector)

		k8sNodeCandidate, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: topologySelector,
		})
		if err != nil {
			return errors.Wrapf(err, "error trying to find node with label %s having value %s", zfs.ZFSTopologyKey, zfs.NodeID)
		}
		if k8sNodeCandidate == nil || len(k8sNodeCandidate.Items) != 1 {
			return fmt.Errorf("unable to retrieve a single node by %s for %s", zfs.ZFSTopologyKey, zfs.NodeID)
		}
		k8sNode = k8sNodeCandidate.Items[0]
	}

	isTrue := true
	// as object returned by client go clears all TypeMeta from it.
	nodeGVK := &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Node",
	}
	ownerRef := metav1.OwnerReference{
		APIVersion: nodeGVK.GroupVersion().String(),
		Kind:       nodeGVK.Kind,
		Name:       k8sNode.Name,
		UID:        k8sNode.GetUID(),
		Controller: &isTrue,
	}

	// Build() fn of all controllers calls AddToScheme to adds all types of this
	// clientset into the given scheme.
	// If multiple controllers happen to call this AddToScheme same time,
	// it causes panic with error saying concurrent map access.
	// This lock is used to serialize the AddToScheme call of all controllers.
	controllerMtx.Lock()

	controller, err := NewNodeControllerBuilder().
		withKubeClient(kubeClient).
		withOpenEBSClient(openebsClient).
		withNodeSynced(nodeInformerFactory).
		withNodeLister(nodeInformerFactory).
		withRecorder(kubeClient).
		withEventHandler(nodeInformerFactory).
		withPollInterval(60 * time.Second).
		withOwnerReference(ownerRef).
		withWorkqueueRateLimiting().Build()

	// blocking call, can't use defer to release the lock
	controllerMtx.Unlock()

	if err != nil {
		return errors.Wrapf(err, "error building controller instance")
	}

	nodeInformerFactory.Start(stopCh)

	// Threadiness defines the number of workers to be launched in Run function
	return controller.Run(1, stopCh)
}
