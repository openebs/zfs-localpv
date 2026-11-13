/*
Copyright 2026 The OpenEBS Authors

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

package promote

import (
	"sync"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	clientset "github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned"
	openebsScheme "github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned/scheme"
	informers "github.com/openebs/zfs-localpv/v2/pkg/generated/informer/externalversions"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const controllerAgentName = "zfspromote-controller"

// Start starts the zfspromote controller.
func Start(controllerMtx *sync.RWMutex, stopCh <-chan struct{}) error {
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building openebs clientset")
	}

	factory := informers.NewSharedInformerFactory(openebsClient, time.Second*30)

	// AddToScheme is not safe to call concurrently; every controller takes
	// this lock around it.
	controllerMtx.Lock()
	err = openebsScheme.AddToScheme(scheme.Scheme)
	controllerMtx.Unlock()
	if err != nil {
		return errors.Wrap(err, "error adding openebs types to the scheme")
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller, err := newPromoteController(openebsClient, factory, recorder, zfs.OpenEBSNamespace, zfs.NodeID)
	if err != nil {
		return errors.Wrap(err, "error building controller instance")
	}

	go factory.Start(stopCh)

	// Threadiness defines the number of workers to be launched in Run function
	return controller.Run(2, stopCh)
}
