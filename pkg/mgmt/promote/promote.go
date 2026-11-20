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
	"context"
	"fmt"
	"time"

	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	clientset "github.com/openebs/zfs-localpv/v2/pkg/generated/clientset/versioned"
	informers "github.com/openebs/zfs-localpv/v2/pkg/generated/informer/externalversions"
	listers "github.com/openebs/zfs-localpv/v2/pkg/generated/lister/zfs/v1"
	"github.com/openebs/zfs-localpv/v2/pkg/zfs"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// zfsOps are the zfs commands the controller runs, replaceable in tests
type zfsOps struct {
	promote       func(vol *apis.ZFSVolume) (bool, error)
	listSnapshots func(vol *apis.ZFSVolume) ([]string, error)
}

// PromoteController handles the ZFSPromote resources of this node
type PromoteController struct {
	clientset clientset.Interface
	// namespace holds the ZFSVolume and ZFSSnapshot records
	namespace string
	nodeID    string
	zfs       zfsOps

	promoteLister listers.ZFSPromoteLister
	promoteSynced cache.InformerSynced
	workqueue     workqueue.TypedRateLimitingInterface[string]
	recorder      record.EventRecorder
}

func newPromoteController(
	cs clientset.Interface,
	factory informers.SharedInformerFactory,
	recorder record.EventRecorder,
	namespace, nodeID string,
) (*PromoteController, error) {
	informer := factory.Zfs().V1().ZFSPromotes()
	c := &PromoteController{
		clientset: cs,
		namespace: namespace,
		nodeID:    nodeID,
		zfs: zfsOps{
			promote:       zfs.PromoteVolume,
			listSnapshots: zfs.ListVolumeSnapshots,
		},
		promoteLister: informer.Lister(),
		promoteSynced: informer.Informer().HasSynced,
		workqueue: workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.DefaultTypedControllerRateLimiter[string](),
			workqueue.TypedRateLimitingQueueConfig[string]{Name: "Promote"},
		),
		recorder: recorder,
	}
	_, err := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueuePromote,
		UpdateFunc: func(_, newObj interface{}) { c.enqueuePromote(newObj) },
	})
	return c, err
}

// isPending tells whether the ZFSPromote is this node's and still to do.
// InProgress counts: an agent that died midway finds it so on restart.
func (c *PromoteController) isPending(p *apis.ZFSPromote) bool {
	return p.Spec.OwnerNodeID == c.nodeID &&
		p.DeletionTimestamp == nil &&
		(p.Status == apis.PromoteZFSStatusInit || p.Status == apis.PromoteZFSStatusInProgress)
}

// enqueuePromote is the add and update event handler for ZFSPromote
func (c *PromoteController) enqueuePromote(obj interface{}) {
	p, ok := obj.(*apis.ZFSPromote)
	if !ok {
		runtime.HandleError(fmt.Errorf("couldn't get promote object %#v", obj))
		return
	}
	if !c.isPending(p) {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(p)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	klog.Infof("Got event for Promote %s vol %s status %s", key, p.Spec.VolumeName, p.Status)
	c.workqueue.Add(key)
}

// syncHandler handles the ZFSPromote of the given namespace/name key
func (c *PromoteController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	p, err := c.promoteLister.ZFSPromotes(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		klog.Infof("zfs promote '%s' has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}
	return c.syncPromote(context.TODO(), p.DeepCopy())
}

// syncPromote promotes the volume of a pending ZFSPromote and records the
// result in its status. An error return means the status could not be
// written, and the key is retried.
func (c *PromoteController) syncPromote(ctx context.Context, p *apis.ZFSPromote) error {
	if !c.isPending(p) {
		return nil
	}

	if p.Status == apis.PromoteZFSStatusInit {
		var err error
		if p, err = c.setStatus(ctx, p, apis.PromoteZFSStatusInProgress); err != nil {
			return err
		}
	}

	if err := c.promote(ctx, p.Spec.VolumeName); err != nil {
		klog.Errorf("promote %s of volume %s failed: %v", p.Name, p.Spec.VolumeName, err)
		zfs.EmitFailureEvent(c.recorder, p, zfs.ReasonPromoteFailed, err)
		_, err = c.setStatus(ctx, p, apis.PromoteZFSStatusFailed)
		return err
	}

	klog.Infof("promote %s of volume %s done", p.Name, p.Spec.VolumeName)
	zfs.EmitSuccessEvent(c.recorder, p, zfs.ReasonPromoted, "volume promoted")
	_, err := c.setStatus(ctx, p, apis.PromoteZFSStatusDone)
	return err
}

// promote runs `zfs promote` on the volume, if it is still a clone, and
// relabels the snapshot records of every snapshot it now holds. Both steps
// are safe to rerun.
func (c *PromoteController) promote(ctx context.Context, volumeName string) error {
	vol, err := c.clientset.ZfsV1().ZFSVolumes(c.namespace).Get(ctx, volumeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get ZFSVolume %s: %w", volumeName, err)
	}
	if vol.Spec.OwnerNodeID != c.nodeID {
		return fmt.Errorf("ZFSVolume %s is on node %s, not %s", volumeName, vol.Spec.OwnerNodeID, c.nodeID)
	}

	if _, err := c.zfs.promote(vol); err != nil {
		return err
	}

	snapshots, err := c.zfs.listSnapshots(vol)
	if err != nil {
		return err
	}
	for _, name := range snapshots {
		if err := c.relabel(ctx, name, volumeName); err != nil {
			return err
		}
	}
	return nil
}

// relabel points the ZFSSnapshot record of the named snapshot at volumeName.
// A snapshot without a record, such as the one a volume-to-volume clone is
// made from, is left alone.
func (c *PromoteController) relabel(ctx context.Context, snapName, volumeName string) error {
	snaps := c.clientset.ZfsV1().ZFSSnapshots(c.namespace)
	snap, err := snaps.Get(ctx, snapName, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get ZFSSnapshot %s: %w", snapName, err)
	}
	old := snap.Labels[zfs.ZFSVolKey]
	if old == volumeName {
		return nil
	}
	if snap.Labels == nil {
		snap.Labels = map[string]string{}
	}
	snap.Labels[zfs.ZFSVolKey] = volumeName
	if _, err := snaps.Update(ctx, snap, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("relabel ZFSSnapshot %s: %w", snapName, err)
	}
	klog.Infof("promote: snapshot %s moved from volume %s to %s", snapName, old, volumeName)
	return nil
}

// setStatus writes the status of the ZFSPromote and returns the result
func (c *PromoteController) setStatus(ctx context.Context, p *apis.ZFSPromote, status apis.ZFSPromoteStatus) (*apis.ZFSPromote, error) {
	p = p.DeepCopy()
	p.Status = status
	updated, err := c.clientset.ZfsV1().ZFSPromotes(p.Namespace).Update(ctx, p, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("could not set promote %s status to %s: %v", p.Name, status, err)
		return nil, err
	}
	return updated, nil
}

// Run waits for the informer cache and starts threadiness workers. It
// blocks until stopCh is closed.
func (c *PromoteController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting Promote controller")
	if ok := cache.WaitForCacheSync(stopCh, c.promoteSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Started Promote workers")
	<-stopCh
	klog.Info("Shutting down Promote workers")
	return nil
}

func (c *PromoteController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *PromoteController) processNextWorkItem() bool {
	key, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	defer c.workqueue.Done(key)

	if err := c.syncHandler(key); err != nil {
		c.workqueue.AddRateLimited(key)
		runtime.HandleError(fmt.Errorf("error syncing '%s': %w, requeuing", key, err))
		return true
	}
	c.workqueue.Forget(key)
	return true
}
