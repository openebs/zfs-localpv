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

package snapshot

import (
	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	openebsScheme "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset/scheme"
	informers "github.com/openebs/zfs-localpv/pkg/generated/informer/externalversions"
	listers "github.com/openebs/zfs-localpv/pkg/generated/lister/zfs/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const controllerAgentName = "zfssnap-controller"

// SnapController is the controller implementation for Snap resources
type SnapController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	snapLister listers.ZFSSnapshotLister

	// snapSynced is used for caches sync to get populated
	snapSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// SnapControllerBuilder is the builder object for controller.
type SnapControllerBuilder struct {
	SnapController *SnapController
}

// NewSnapControllerBuilder returns an empty instance of controller builder.
func NewSnapControllerBuilder() *SnapControllerBuilder {
	return &SnapControllerBuilder{
		SnapController: &SnapController{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *SnapControllerBuilder) withKubeClient(ks kubernetes.Interface) *SnapControllerBuilder {
	cb.SnapController.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *SnapControllerBuilder) withOpenEBSClient(cs clientset.Interface) *SnapControllerBuilder {
	cb.SnapController.clientset = cs
	return cb
}

// withSnapLister fills snap lister to controller object.
func (cb *SnapControllerBuilder) withSnapLister(sl informers.SharedInformerFactory) *SnapControllerBuilder {
	snapInformer := sl.Zfs().V1().ZFSSnapshots()
	cb.SnapController.snapLister = snapInformer.Lister()
	return cb
}

// withSnapSynced adds object sync information in cache to controller object.
func (cb *SnapControllerBuilder) withSnapSynced(sl informers.SharedInformerFactory) *SnapControllerBuilder {
	snapInformer := sl.Zfs().V1().ZFSSnapshots()
	cb.SnapController.snapSynced = snapInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *SnapControllerBuilder) withWorkqueueRateLimiting() *SnapControllerBuilder {
	cb.SnapController.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Snap")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *SnapControllerBuilder) withRecorder(ks kubernetes.Interface) *SnapControllerBuilder {
	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.SnapController.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *SnapControllerBuilder) withEventHandler(cvcInformerFactory informers.SharedInformerFactory) *SnapControllerBuilder {
	cvcInformer := cvcInformerFactory.Zfs().V1().ZFSSnapshots()
	// Set up an event handler for when Snap resources change
	cvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.SnapController.addSnap,
		UpdateFunc: cb.SnapController.updateSnap,
		DeleteFunc: cb.SnapController.deleteSnap,
	})
	return cb
}

// Build returns a controller instance.
func (cb *SnapControllerBuilder) Build() (*SnapController, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.SnapController, nil
}
