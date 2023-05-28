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
	"time"

	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	openebsScheme "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset/scheme"
	informers "github.com/openebs/zfs-localpv/pkg/generated/informer/externalversions"
	listers "github.com/openebs/zfs-localpv/pkg/generated/lister/zfs/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "zfsnode-controller"

// NodeController is the controller implementation for zfs node resources
type NodeController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	NodeLister listers.ZFSNodeLister

	// NodeSynced is used for caches sync to get populated
	NodeSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// pollInterval controls the polling frequency of syncing up the vg metadata.
	pollInterval time.Duration

	// ownerRef is used to set the owner reference to zfsnode objects.
	ownerRef metav1.OwnerReference
}

// NodeControllerBuilder is the builder object for controller.
type NodeControllerBuilder struct {
	NodeController *NodeController
}

// NewNodeControllerBuilder returns an empty instance of controller builder.
func NewNodeControllerBuilder() *NodeControllerBuilder {
	return &NodeControllerBuilder{
		NodeController: &NodeController{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *NodeControllerBuilder) withKubeClient(ks kubernetes.Interface) *NodeControllerBuilder {
	cb.NodeController.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *NodeControllerBuilder) withOpenEBSClient(cs clientset.Interface) *NodeControllerBuilder {
	cb.NodeController.clientset = cs
	return cb
}

// withNodeLister fills Node lister to controller object.
func (cb *NodeControllerBuilder) withNodeLister(sl informers.SharedInformerFactory) *NodeControllerBuilder {
	NodeInformer := sl.Zfs().V1().ZFSNodes()
	cb.NodeController.NodeLister = NodeInformer.Lister()
	return cb
}

// withNodeSynced adds object sync information in cache to controller object.
func (cb *NodeControllerBuilder) withNodeSynced(sl informers.SharedInformerFactory) *NodeControllerBuilder {
	NodeInformer := sl.Zfs().V1().ZFSNodes()
	cb.NodeController.NodeSynced = NodeInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *NodeControllerBuilder) withWorkqueueRateLimiting() *NodeControllerBuilder {
	cb.NodeController.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Node")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *NodeControllerBuilder) withRecorder(ks kubernetes.Interface) *NodeControllerBuilder {
	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.NodeController.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *NodeControllerBuilder) withEventHandler(cvcInformerFactory informers.SharedInformerFactory) *NodeControllerBuilder {
	cvcInformer := cvcInformerFactory.Zfs().V1().ZFSNodes()
	// Set up an event handler for when zfs node vg change.
	// Note: rather than setting up the resync period at informer level,
	// we are controlling the syncing based on pollInternal. See
	// NodeController#Run func for more details.
	cvcInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.NodeController.addNode,
		UpdateFunc: cb.NodeController.updateNode,
		DeleteFunc: cb.NodeController.deleteNode,
	}, 0)
	return cb
}

func (cb *NodeControllerBuilder) withPollInterval(interval time.Duration) *NodeControllerBuilder {
	cb.NodeController.pollInterval = interval
	return cb
}

func (cb *NodeControllerBuilder) withOwnerReference(ownerRef metav1.OwnerReference) *NodeControllerBuilder {
	cb.NodeController.ownerRef = ownerRef
	return cb
}

// Build returns a controller instance.
func (cb *NodeControllerBuilder) Build() (*NodeController, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.NodeController, nil
}
