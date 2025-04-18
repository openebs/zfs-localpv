/*
Copyright © 2025 The OpenEBS Authors

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
	"fmt"

	"github.com/openebs/lib-csi/pkg/common/errors"
	zfsapi "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/builder/bkpbuilder"
	clientset "github.com/openebs/zfs-localpv/pkg/generated/clientset/internalclientset"
	informers "github.com/openebs/zfs-localpv/pkg/generated/informer/externalversions"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// BackupSnapshotIndex is the name of the index that allows lookup of ZFSBackup resources
// by the combination of VolumeName, OwnerNodeID, and SnapName
const BackupSnapshotIndex = "BackupSnapshotIndex"

// BackupSnapshotIndexFunc creates an index based on the combination of VolumeName, OwnerNodeID, and SnapName
// This allows efficient lookup of backups by these three fields combined
func BackupSnapshotIndexFunc(obj interface{}) ([]string, error) {
	backup, ok := obj.(*zfsapi.ZFSBackup)
	if !ok {
		return nil, nil
	}

	// Combine the three fields with a dot separator for unique key generation
	composedKey := backup.Spec.VolumeName + "." + backup.Spec.OwnerNodeID + "." + backup.Spec.SnapName
	return []string{composedKey}, nil
}

// BackupPreviousSnapshotIndex is the name of the index that allows lookup of ZFSBackup resources
// by their referenced previous snapshot information (VolumeName, OwnerNodeID, and PrevSnapName)
const BackupPreviousSnapshotIndex = "BackupPreviousSnapshotIndex"

// BackupPreviousSnapshotIndexFunc creates an index based on the previous snapshot reference
// This allows efficient lookup of backups that reference a particular snapshot as their previous snapshot
func BackupPreviousSnapshotIndexFunc(obj interface{}) ([]string, error) {
	backup, ok := obj.(*zfsapi.ZFSBackup)
	if !ok {
		return nil, nil
	}

	// If there's no previous snapshot reference, don't index this backup
	if backup.Spec.PrevSnapName == "" {
		return nil, nil
	}

	// Combine the three fields with a dot separator for unique key generation
	composedKey := backup.Spec.VolumeName + "." + backup.Spec.OwnerNodeID + "." + backup.Spec.PrevSnapName
	return []string{composedKey}, nil
}

// BackupGarbageCollector manages the lifecycle of ZFS backup resources by
// ensuring orphaned backups (those whose previous snapshots have been deleted)
// are cleaned up properly to maintain backup chain integrity.
type BackupGarbageCollector struct {
	zfsBackupInformer cache.SharedIndexInformer
}

// Initialize sets up the BackupGarbageCollector by configuring clients, informers,
// event handlers and indexers required for monitoring backup resources
func (bgc *BackupGarbageCollector) Initialize(openebsClient *clientset.Clientset, stopCh <-chan struct{}) error {
	openebsInformerFactory := informers.NewSharedInformerFactoryWithOptions(openebsClient,
		0, informers.WithNamespace(zfs.OpenEBSNamespace))

	bgc.zfsBackupInformer = openebsInformerFactory.Zfs().V1().ZFSBackups().Informer()

	if err := bgc.setupIndexers(); err != nil {
		return errors.Wrapf(err, "failed to set up informer indexers")
	}

	bgc.registerEventHandlers()

	return bgc.startAndWaitForInformer(stopCh)
}

// InitializeForTesting sets up the BackupGarbageCollector for testing purposes
// by allowing direct clientset and namespace injection
func (bgc *BackupGarbageCollector) InitializeForTesting(openebsClient clientset.Interface, namespace string) error {
	// Create informer factory with namespace filtering
	openebsInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		openebsClient,
		0,
		informers.WithNamespace(namespace),
	)

	stopCh := signals.SetupSignalHandler()

	bgc.zfsBackupInformer = openebsInformerFactory.Zfs().V1().ZFSBackups().Informer()

	// Add indexers to the informer for efficient lookup
	if err := bgc.setupIndexers(); err != nil {
		return err
	}

	// Register event handlers for monitoring backup resources
	bgc.registerEventHandlers()

	// Start the informer and wait for the cache to sync
	return bgc.startAndWaitForInformer(stopCh)
}

// setupIndexers configures the custom indexes used for efficient backup lookups
func (bgc *BackupGarbageCollector) setupIndexers() error {
	indexers := map[string]cache.IndexFunc{
		BackupSnapshotIndex:         BackupSnapshotIndexFunc,
		BackupPreviousSnapshotIndex: BackupPreviousSnapshotIndexFunc,
	}

	return bgc.zfsBackupInformer.AddIndexers(indexers)
}

// registerEventHandlers adds handlers for Add, Update, and Delete events of ZFSBackup resources
func (bgc *BackupGarbageCollector) registerEventHandlers() {
	bgc.zfsBackupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bgc.handleBackupCreation,
		UpdateFunc: bgc.handleBackupUpdate,
		DeleteFunc: bgc.handleBackupDeletion,
	})
}

// startAndWaitForInformer runs the informer and ensures the cache is synced before proceeding
func (bgc *BackupGarbageCollector) startAndWaitForInformer(stopCh <-chan struct{}) error {
	go bgc.zfsBackupInformer.Run(stopCh)

	klog.Info("Waiting for ZFS backup informer cache to be synced")
	if !cache.WaitForCacheSync(stopCh, bgc.zfsBackupInformer.HasSynced) {
		return errors.New("timed out waiting for ZFS backup informer cache to sync")
	}

	klog.Info("ZFS backup informer cache synced successfully")
	return nil
}

// handleBackupCreation processes new ZFSBackup resources when they are created.
// It validates whether the previous snapshot references are valid and initiates cleanup
// of invalid references to maintain backup chain integrity.
func (bgc *BackupGarbageCollector) handleBackupCreation(obj interface{}) {
	backup, ok := obj.(*zfsapi.ZFSBackup)
	if !ok {
		klog.Errorf("Expected ZFSBackup, got %T", obj)
		return
	}

	klog.InfoS("Processing backup creation", "backupName", backup.Name)

	// Check if the backup references a prevSnapName that doesn't exist
	if backup.Spec.PrevSnapName != "" {
		go bgc.deleteOrphanedIncBackup(backup)
	}
}

// handleBackupUpdate processes changes to existing ZFSBackup resources.
// It specifically handles changes to the previous snapshot reference to ensure
// the backup chain remains valid after updates.
func (bgc *BackupGarbageCollector) handleBackupUpdate(oldObj, newObj interface{}) {
	oldBackup, ok := oldObj.(*zfsapi.ZFSBackup)
	if !ok {
		klog.Errorf("Expected ZFSBackup, got %T", oldObj)
		return
	}

	newBackup, ok := newObj.(*zfsapi.ZFSBackup)
	if !ok {
		klog.Errorf("Expected ZFSBackup, got %T", newObj)
		return
	}

	klog.InfoS("Processing backup update",
		"backupName", newBackup.Name,
		"prevSnapshot", newBackup.Spec.PrevSnapName)

	// If prevSnapName was added or changed, validate it
	if oldBackup.Spec.PrevSnapName != newBackup.Spec.PrevSnapName && newBackup.Spec.PrevSnapName != "" {
		go bgc.deleteOrphanedIncBackup(newBackup)
	}
}

// handleBackupDeletion processes ZFSBackup deletion events and cleans up any
// child backups that reference the deleted backup as their previous snapshot.
// This ensures the backup chain integrity by removing dependent backups when
// a parent backup is removed.
func (bgc *BackupGarbageCollector) handleBackupDeletion(obj interface{}) {
	backup, ok := obj.(*zfsapi.ZFSBackup)
	if !ok {
		// In case of delete event, we might get a DeletedFinalStateUnknown instead of the object
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		backup, ok = tombstone.Obj.(*zfsapi.ZFSBackup)
		if !ok {
			klog.Errorf("Tombstone contained object that is not a ZFSBackup: %#v", obj)
			return
		}
	}

	klog.InfoS("Processing backup deletion",
		"backupName", backup.Name,
		"volumeName", backup.Spec.VolumeName,
		"snapshot", backup.Spec.SnapName)

	// When a backup is deleted, find and delete any backups that reference it as prevSnapName
	go bgc.deleteBackupsReferencingDeletedBackup(backup)
}

// deleteOrphanedIncBackup checks if the previous snapshot referenced by a backup exists.
// If the reference doesn't exist, the backup is deleted to maintain chain integrity.
func (bgc *BackupGarbageCollector) deleteOrphanedIncBackup(backup *zfsapi.ZFSBackup) {
	if backup.Spec.PrevSnapName == "" {
		return
	}

	parentBackupKey := bgc.getBackupKeyBySpec(
		backup.Spec.VolumeName,
		backup.Spec.OwnerNodeID,
		backup.Spec.PrevSnapName)

	parentItem, err := bgc.zfsBackupInformer.GetIndexer().ByIndex(BackupSnapshotIndex, parentBackupKey)
	if err != nil {
		klog.ErrorS(err, "Error validating previous snapshot reference for backup",
			"backupName", backup.Name,
			"prevSnapshotName", backup.Spec.PrevSnapName)
		return
	}
	if len(parentItem) == 0 {
		// PrevSnapName doesn't exist, delete the backup
		klog.InfoS("Deleting backup with invalid previous snapshot reference",
			"backupName", backup.Name,
			"prevSnapshotName", backup.Spec.PrevSnapName)

		deleteErr := bkpbuilder.NewKubeclient().WithNamespace(zfs.OpenEBSNamespace).Delete(backup.Name)

		if deleteErr != nil {
			klog.ErrorS(deleteErr, "Failed to delete backup with invalid previous snapshot reference",
				"backupName", backup.Name)
		}
	}
}

// getBackupKeyBySpec creates a unique key from volume name, node ID, and snapshot name
// This utility helps maintain consistent key generation across different functions
func (bgc *BackupGarbageCollector) getBackupKeyBySpec(volumeName, ownerNodeID, snapName string) string {
	return volumeName + "." + ownerNodeID + "." + snapName
}

// deleteBackupsReferencingDeletedBackup finds and deletes any ZFSBackups that reference
// the deleted backup as their prevSnapName to maintain backup chain integrity
func (bgc *BackupGarbageCollector) deleteBackupsReferencingDeletedBackup(backup *zfsapi.ZFSBackup) {
	// Create key to look up child backups that reference this backup as their previous snapshot
	childBackupKey := bgc.getBackupKeyBySpec(
		backup.Spec.VolumeName,
		backup.Spec.OwnerNodeID,
		backup.Spec.SnapName)

	possibleChildBackups, err := bgc.zfsBackupInformer.GetIndexer().ByIndex(
		BackupPreviousSnapshotIndex,
		childBackupKey)

	if err != nil {
		klog.ErrorS(err, "Error looking up dependent backups",
			"volumeName", backup.Spec.VolumeName,
			"ownerNodeID", backup.Spec.OwnerNodeID,
			"snapshot", backup.Spec.SnapName)
		return
	}

	// No dependent backups found
	if len(possibleChildBackups) == 0 {
		klog.InfoS("No dependent backups found that reference deleted backup",
			"backupName", backup.Name,
			"snapshot", backup.Spec.SnapName)
		return
	}

	// Transform the possible child backups to ZFSBackup objects
	childBackups, err := bgc.parseBackupsFromIndexResults(possibleChildBackups)
	if err != nil {
		klog.ErrorS(err, "Error parsing possible child backups to ZFSBackup objects")
		return
	}

	// Delete each dependent backup
	for _, childBackup := range childBackups {
		klog.InfoS("Deleting dependent backup with broken reference chain",
			"backupName", childBackup.Name,
			"prevSnapshotName", childBackup.Spec.PrevSnapName)

		deleteErr := bkpbuilder.NewKubeclient().
			WithNamespace(zfs.OpenEBSNamespace).
			Delete(childBackup.Name)

		if deleteErr != nil {
			klog.ErrorS(deleteErr, "Failed to delete dependent backup",
				"backupName", childBackup.Name)
		}
	}
}

// parseBackupsFromIndexResults extracts typed ZFSBackup objects from generic interface slice
// returned by informer indexers. Fails immediately on encountering any invalid type.
func (bgc *BackupGarbageCollector) parseBackupsFromIndexResults(objects []interface{}) ([]*zfsapi.ZFSBackup, error) {
	// Pre-allocate with capacity to avoid reallocations
	backups := make([]*zfsapi.ZFSBackup, 0, len(objects))

	// Process all objects, failing on first error
	for i, obj := range objects {
		backup, ok := obj.(*zfsapi.ZFSBackup)
		if !ok {
			return nil, fmt.Errorf("item[%d]: expected *zfsapi.ZFSBackup, got %T", i, obj)
		}
		backups = append(backups, backup)
	}

	return backups, nil
}
