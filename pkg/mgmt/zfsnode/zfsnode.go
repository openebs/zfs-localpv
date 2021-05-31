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
	"fmt"
	"reflect"
	"time"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/nodebuilder"
	"github.com/openebs/zfs-localpv/pkg/equality"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func (c *NodeController) listZFSPool() ([]apis.Pool, error) {
	return zfs.ListZFSPool()
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *NodeController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	return c.syncNode(namespace, name)
}

// syncNode is the function which tries to converge to a desired state for the
// ZFSNode
func (c *NodeController) syncNode(namespace string, name string) error {
	// Get the node resource with this namespace/name
	cachedNode, err := c.NodeLister.ZFSNodes(namespace).Get(name)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}
	var node *apis.ZFSNode
	if cachedNode != nil {
		node = cachedNode.DeepCopy()
	}

	pools, err := c.listZFSPool()
	if err != nil {
		return err
	}

	if node == nil { // if it doesn't exists, create zfs node object
		if node, err = nodebuilder.NewBuilder().
			WithNamespace(namespace).WithName(name).
			WithPools(pools).
			WithOwnerReferences(c.ownerRef).
			Build(); err != nil {
			return err
		}

		klog.Infof("zfs node controller: creating new node object for %+v", node)
		if node, err = nodebuilder.NewKubeclient().WithNamespace(namespace).Create(node); err != nil {
			return fmt.Errorf("create zfs node %s/%s: %v", namespace, name, err)
		}
		klog.Infof("zfs node controller: created node object %s/%s", namespace, name)
		return nil
	}

	// zfs node already exists check if we need to update it.
	var updateRequired bool
	// validate if owner reference updated.
	if ownerRefs, req := c.isOwnerRefsUpdateRequired(node.OwnerReferences); req {
		klog.Infof("zfs node controller: node owner references updated current=%+v, required=%+v",
			node.OwnerReferences, ownerRefs)
		node.OwnerReferences = ownerRefs
		updateRequired = true
	}

	// validate if node pools are upto date.
	if !equality.Semantic.DeepEqual(node.Pools, pools) {
		klog.Infof("zfs node controller: node pools updated current=%+v, required=%+v",
			node.Pools, pools)
		node.Pools = pools
		updateRequired = true
	}

	if !updateRequired {
		return nil
	}

	klog.Infof("zfs node controller: updating node object with %+v", node)
	if _, err = nodebuilder.NewKubeclient().WithNamespace(namespace).Update(node); err != nil {
		return fmt.Errorf("update zfs node %s/%s: %v", namespace, name, err)
	}
	klog.Infof("zfs node controller: updated node object %s/%s", namespace, name)

	return nil
}

// addNode is the add event handler for ZFSNode
func (c *NodeController) addNode(obj interface{}) {
	node, ok := obj.(*apis.ZFSNode)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get node object %#v", obj))
		return
	}

	klog.Infof("Got add event for zfs node %s/%s", node.Namespace, node.Name)
	c.enqueueNode(node)
}

// updateNode is the update event handler for ZFSNode
func (c *NodeController) updateNode(oldObj, newObj interface{}) {
	newNode, ok := newObj.(*apis.ZFSNode)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get node object %#v", newNode))
		return
	}

	klog.Infof("Got update event for zfs node %s/%s", newNode.Namespace, newNode.Name)
	c.enqueueNode(newNode)
}

// deleteNode is the delete event handler for ZFSNode
func (c *NodeController) deleteNode(obj interface{}) {
	node, ok := obj.(*apis.ZFSNode)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		node, ok = tombstone.Obj.(*apis.ZFSNode)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a ZFSNode %#v", obj))
			return
		}
	}

	klog.Infof("Got delete event for node %s/%s", node.Namespace, node.Name)
	c.enqueueNode(node)
}

// enqueueNode takes a ZFSNode resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than ZFSNode.
func (c *NodeController) enqueueNode(node *apis.ZFSNode) {
	// node must exists in openebs namespace & must equal to the node id.
	if node.Namespace != zfs.OpenEBSNamespace ||
		node.Name != zfs.NodeID {
		klog.Warningf("skipping zfs node object %s/%s", node.Namespace, node.Name)
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(node)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *NodeController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Node controller")

	// Wait for the k8s caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.NodeSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting Node workers")
	// Launch worker to process Node resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started Node workers")

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
		case <-stopCh:
			klog.Info("Shutting down Node controller")
			return nil
		}
		item := zfs.OpenEBSNamespace + "/" + zfs.NodeID
		c.workqueue.Add(item) // add the item to worker queue.
		timer.Reset(c.pollInterval)
	}
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *NodeController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *NodeController) processNextWorkItem() bool {
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
		// Node resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.V(5).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// isOwnerRefUpdateRequired validates if relevant owner references is being
// set for zfs node. If not, it returns the final owner references that needs
// to be set.
func (c *NodeController) isOwnerRefsUpdateRequired(ownerRefs []metav1.OwnerReference) ([]metav1.OwnerReference, bool) {
	updated := false
	reqOwnerRef := c.ownerRef
	for idx := range ownerRefs {
		if ownerRefs[idx].UID != reqOwnerRef.UID {
			continue
		}
		// in case owner reference exists, validate
		// if controller field is set correctly or not.
		if !reflect.DeepEqual(ownerRefs[idx].Controller, reqOwnerRef.Controller) {
			updated = true
			ownerRefs[idx].Controller = reqOwnerRef.Controller
		}
		return ownerRefs, updated
	}
	updated = true
	ownerRefs = append(ownerRefs, reqOwnerRef)
	return ownerRefs, updated
}
