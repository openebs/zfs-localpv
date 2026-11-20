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

package driver

import (
	"context"
	"time"

	"github.com/openebs/lib-csi/pkg/common/errors"
	zfsapi "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/builder/volbuilder"
	clientset "github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned"
	informers "github.com/openebs/zfs-localpv/v2/pkg/generated/informer/externalversions"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// markedVolumeResync is how often every ZFSVolume is re-delivered to the
// reaper, as a safety net for events it missed.
const markedVolumeResync = 10 * time.Minute

// markedVolumeReaper finishes deleting the ZFSVolumes that DeleteVolume only
// marked for deletion because they still had snapshots.
//
// A ZFSVolume that is marked for deletion and has no ZFSSnapshot left is
// deleted, however the last snapshot went away. The check runs when a
// ZFSSnapshot is deleted or moves to another volume, when a ZFSVolume is
// added or updated (which includes being marked), and on every resync.
type markedVolumeReaper struct {
	client     clientset.Interface
	namespace  string
	volumeLock *volumeLock
	queue      workqueue.TypedRateLimitingInterface[string]
}

func newMarkedVolumeReaper(client clientset.Interface, namespace string, lock *volumeLock) *markedVolumeReaper {
	return &markedVolumeReaper{
		client:     client,
		namespace:  namespace,
		volumeLock: lock,
		queue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{Name: "MarkedZV"},
		),
	}
}

// Start registers the event handlers, waits for the informer caches and
// starts the worker. The worker stops when stopCh is closed.
func (r *markedVolumeReaper) Start(stopCh <-chan struct{}) error {
	factory := informers.NewSharedInformerFactoryWithOptions(r.client,
		markedVolumeResync, informers.WithNamespace(r.namespace))
	zvInformer := factory.Zfs().V1().ZFSVolumes().Informer()
	snapInformer := factory.Zfs().V1().ZFSSnapshots().Informer()

	if _, err := zvInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.volumeChanged,
		UpdateFunc: func(_, newObj interface{}) { r.volumeChanged(newObj) },
	}); err != nil {
		return err
	}
	if _, err := snapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: r.snapshotUpdated,
		DeleteFunc: r.snapshotDeleted,
	}); err != nil {
		return err
	}

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, zvInformer.HasSynced, snapInformer.HasSynced) {
		return errors.New("timed out waiting for zfs volume and snapshot informer caches to sync")
	}

	go func() {
		<-stopCh
		r.queue.ShutDown()
	}()
	go wait.Until(r.runWorker, time.Second, stopCh)
	return nil
}

// volumeChanged queues a ZFSVolume that is marked for deletion and not yet
// being deleted.
func (r *markedVolumeReaper) volumeChanged(obj interface{}) {
	zv, ok := obj.(*zfsapi.ZFSVolume)
	if !ok {
		return
	}
	if isMarkedForDeletion(zv) && zv.DeletionTimestamp == nil {
		r.queue.Add(zv.Name)
	}
}

// snapshotUpdated queues the volume a ZFSSnapshot no longer belongs to.
func (r *markedVolumeReaper) snapshotUpdated(oldObj, newObj interface{}) {
	oldSnap, ok := oldObj.(*zfsapi.ZFSSnapshot)
	if !ok {
		return
	}
	newSnap, ok := newObj.(*zfsapi.ZFSSnapshot)
	if !ok {
		return
	}
	if oldVol := oldSnap.Labels[zfs.ZFSVolKey]; oldVol != "" && oldVol != newSnap.Labels[zfs.ZFSVolKey] {
		r.queue.Add(oldVol)
	}
}

// snapshotDeleted queues the volume of a ZFSSnapshot that is gone.
func (r *markedVolumeReaper) snapshotDeleted(obj interface{}) {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = tombstone.Obj
	}
	snap, ok := obj.(*zfsapi.ZFSSnapshot)
	if !ok {
		return
	}
	if vol := snap.Labels[zfs.ZFSVolKey]; vol != "" {
		r.queue.Add(vol)
	}
}

func (r *markedVolumeReaper) runWorker() {
	for r.processNextItem() {
	}
}

func (r *markedVolumeReaper) processNextItem() bool {
	volName, shutdown := r.queue.Get()
	if shutdown {
		return false
	}
	defer r.queue.Done(volName)

	if err := r.reap(context.TODO(), volName); err != nil {
		klog.Errorf("failed to finish deletion of marked volume %s, requeuing: %v", volName, err)
		r.queue.AddRateLimited(volName)
		return true
	}
	r.queue.Forget(volName)
	return true
}

// reap deletes the ZFSVolume volName if it is marked for deletion and no
// ZFSSnapshot belongs to it any more. It holds the volume lock, so it does not
// interleave with DeleteVolume, CreateSnapshot or DeleteSnapshot, and it reads
// from the API server rather than the informer cache.
func (r *markedVolumeReaper) reap(ctx context.Context, volName string) error {
	unlock := r.volumeLock.LockVolume(volName)
	defer unlock()

	zv, err := r.client.ZfsV1().ZFSVolumes(r.namespace).Get(ctx, volName, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !isMarkedForDeletion(zv) || zv.DeletionTimestamp != nil {
		return nil
	}

	snaps, err := r.client.ZfsV1().ZFSSnapshots(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: zfs.ZFSVolKey + "=" + volName,
	})
	if err != nil {
		return err
	}
	if len(snaps.Items) > 0 {
		return nil
	}

	propagation := metav1.DeletePropagationForeground
	err = r.client.ZfsV1().ZFSVolumes(r.namespace).Delete(ctx, volName, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
		Preconditions:     &metav1.Preconditions{UID: &zv.UID},
	})
	if k8serror.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	klog.Infof("deleted volume %s, which was marked for deletion and has no snapshots left", volName)
	return nil
}

func isMarkedForDeletion(zv *zfsapi.ZFSVolume) bool {
	return zv.Annotations[volbuilder.MarkForDeletionAnnotation] == "true"
}
