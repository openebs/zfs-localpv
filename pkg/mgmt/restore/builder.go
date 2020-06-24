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

package restore

import (
	"k8s.io/klog"

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

const controllerAgentName = "zfsrestore-controller"

// RstrController is the controller implementation for Restore resources
type RstrController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	rstrLister listers.ZFSRestoreLister

	// backupSynced is used for caches sync to get populated
	rstrSynced cache.InformerSynced

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

// RstrControllerBuilder is the builder object for controller.
type RstrControllerBuilder struct {
	RstrController *RstrController
}

// NewRstrControllerBuilder returns an empty instance of controller builder.
func NewRstrControllerBuilder() *RstrControllerBuilder {
	return &RstrControllerBuilder{
		RstrController: &RstrController{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *RstrControllerBuilder) withKubeClient(ks kubernetes.Interface) *RstrControllerBuilder {
	cb.RstrController.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *RstrControllerBuilder) withOpenEBSClient(cs clientset.Interface) *RstrControllerBuilder {
	cb.RstrController.clientset = cs
	return cb
}

// withRestoreLister fills rstr lister to controller object.
func (cb *RstrControllerBuilder) withRestoreLister(sl informers.SharedInformerFactory) *RstrControllerBuilder {
	rstrInformer := sl.Zfs().V1().ZFSRestores()
	cb.RstrController.rstrLister = rstrInformer.Lister()
	return cb
}

// withRestoreSynced adds object sync information in cache to controller object.
func (cb *RstrControllerBuilder) withRestoreSynced(sl informers.SharedInformerFactory) *RstrControllerBuilder {
	rstrInformer := sl.Zfs().V1().ZFSRestores()
	cb.RstrController.rstrSynced = rstrInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *RstrControllerBuilder) withWorkqueueRateLimiting() *RstrControllerBuilder {
	cb.RstrController.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Restore")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *RstrControllerBuilder) withRecorder(ks kubernetes.Interface) *RstrControllerBuilder {
	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.RstrController.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *RstrControllerBuilder) withEventHandler(cvcInformerFactory informers.SharedInformerFactory) *RstrControllerBuilder {
	cvcInformer := cvcInformerFactory.Zfs().V1().ZFSRestores()
	// Set up an event handler for when Restore resources change
	cvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.RstrController.addRestore,
		UpdateFunc: cb.RstrController.updateRestore,
		DeleteFunc: cb.RstrController.deleteRestore,
	})
	return cb
}

// Build returns a controller instance.
func (cb *RstrControllerBuilder) Build() (*RstrController, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.RstrController, nil
}
