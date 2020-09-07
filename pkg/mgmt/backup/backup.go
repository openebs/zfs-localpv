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
	"fmt"
	"k8s.io/klog"
	"time"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	zfs "github.com/openebs/zfs-localpv/pkg/zfs"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

// isDeletionCandidate checks if a zfs backup is a deletion candidate.
func (c *BkpController) isDeletionCandidate(bkp *apis.ZFSBackup) bool {
	return bkp.ObjectMeta.DeletionTimestamp != nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *BkpController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the bkp resource with this namespace/name
	bkp, err := c.bkpLister.ZFSBackups(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("zfs backup '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}
	bkpCopy := bkp.DeepCopy()
	err = c.syncBkp(bkpCopy)
	return err
}

// enqueueBkp takes a ZFSBackup resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than ZFSBackup.
func (c *BkpController) enqueueBkp(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// synBkp is the function which tries to converge to a desired state for the
// ZFSBackup
func (c *BkpController) syncBkp(bkp *apis.ZFSBackup) error {
	var err error = nil
	// ZFSBackup should be deleted. Check if deletion timestamp is set
	if c.isDeletionCandidate(bkp) {
		// reconcile for the Destroy error
		err = zfs.DestoryBackup(bkp)
		if err == nil {
			err = zfs.RemoveBkpFinalizer(bkp)
		}
	} else {
		// if status is init then it means we are creating the zfs backup.
		if bkp.Status == apis.BKPZFSStatusInit {
			err = zfs.CreateBackup(bkp)
			if err == nil {
				klog.Infof("backup %s done %s@%s prevsnap [%s]", bkp.Name, bkp.Spec.VolumeName, bkp.Spec.SnapName, bkp.Spec.PrevSnapName)
				err = zfs.UpdateBkpInfo(bkp, apis.BKPZFSStatusDone)
			} else {
				klog.Errorf("backup %s failed %s@%s err %v", bkp.Name, bkp.Spec.VolumeName, bkp.Spec.SnapName, err)
				err = zfs.UpdateBkpInfo(bkp, apis.BKPZFSStatusFailed)
			}
		}
	}
	return err
}

// addBkp is the add event handler for ZFSBackup
func (c *BkpController) addBkp(obj interface{}) {
	bkp, ok := obj.(*apis.ZFSBackup)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get backup object %#v", obj))
		return
	}

	if zfs.NodeID != bkp.Spec.OwnerNodeID {
		return
	}
	klog.Infof("Got add event for Bkp %s snap %s@%s", bkp.Name, bkp.Spec.VolumeName, bkp.Spec.SnapName)
	c.enqueueBkp(bkp)
}

// updateBkp is the update event handler for ZFSBackup
func (c *BkpController) updateBkp(oldObj, newObj interface{}) {

	newBkp, ok := newObj.(*apis.ZFSBackup)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get bkp object %#v", newBkp))
		return
	}

	if zfs.NodeID != newBkp.Spec.OwnerNodeID {
		return
	}

	if c.isDeletionCandidate(newBkp) {
		klog.Infof("Got update event for Bkp %s snap %s@%s", newBkp.Name, newBkp.Spec.VolumeName, newBkp.Spec.SnapName)
		c.enqueueBkp(newBkp)
	}
}

// deleteBkp is the delete event handler for ZFSBackup
func (c *BkpController) deleteBkp(obj interface{}) {
	bkp, ok := obj.(*apis.ZFSBackup)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		bkp, ok = tombstone.Obj.(*apis.ZFSBackup)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a zfsbackup %#v", obj))
			return
		}
	}

	if zfs.NodeID != bkp.Spec.OwnerNodeID {
		return
	}

	klog.Infof("Got delete event for Bkp %s snap %s@%s", bkp.Name, bkp.Spec.VolumeName, bkp.Spec.SnapName)
	c.enqueueBkp(bkp)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *BkpController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Bkp controller")

	// Wait for the k8s caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.bkpSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	klog.Info("Starting Bkp workers")
	// Launch worker to process Bkp resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started Bkp workers")
	<-stopCh
	klog.Info("Shutting down Bkp workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *BkpController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *BkpController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Bkp resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}
