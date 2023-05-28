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

package backup

import (
	"k8s.io/klog/v2"

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
)

const controllerAgentName = "zfsbackup-controller"

// BkpController is the controller implementation for Bkp resources
type BkpController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	bkpLister listers.ZFSBackupLister

	// backupSynced is used for caches sync to get populated
	bkpSynced cache.InformerSynced

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

// BkpControllerBuilder is the builder object for controller.
type BkpControllerBuilder struct {
	BkpController *BkpController
}

// NewBkpControllerBuilder returns an empty instance of controller builder.
func NewBkpControllerBuilder() *BkpControllerBuilder {
	return &BkpControllerBuilder{
		BkpController: &BkpController{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *BkpControllerBuilder) withKubeClient(ks kubernetes.Interface) *BkpControllerBuilder {
	cb.BkpController.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *BkpControllerBuilder) withOpenEBSClient(cs clientset.Interface) *BkpControllerBuilder {
	cb.BkpController.clientset = cs
	return cb
}

// withBkpLister fills bkp lister to controller object.
func (cb *BkpControllerBuilder) withBkpLister(sl informers.SharedInformerFactory) *BkpControllerBuilder {
	bkpInformer := sl.Zfs().V1().ZFSBackups()
	cb.BkpController.bkpLister = bkpInformer.Lister()
	return cb
}

// withBkpSynced adds object sync information in cache to controller object.
func (cb *BkpControllerBuilder) withBkpSynced(sl informers.SharedInformerFactory) *BkpControllerBuilder {
	bkpInformer := sl.Zfs().V1().ZFSBackups()
	cb.BkpController.bkpSynced = bkpInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *BkpControllerBuilder) withWorkqueueRateLimiting() *BkpControllerBuilder {
	cb.BkpController.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Bkp")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *BkpControllerBuilder) withRecorder(ks kubernetes.Interface) *BkpControllerBuilder {
	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.BkpController.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *BkpControllerBuilder) withEventHandler(cvcInformerFactory informers.SharedInformerFactory) *BkpControllerBuilder {
	cvcInformer := cvcInformerFactory.Zfs().V1().ZFSBackups()
	// Set up an event handler for when Bkp resources change
	cvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.BkpController.addBkp,
		UpdateFunc: cb.BkpController.updateBkp,
		DeleteFunc: cb.BkpController.deleteBkp,
	})
	return cb
}

// Build returns a controller instance.
func (cb *BkpControllerBuilder) Build() (*BkpController, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.BkpController, nil
}
