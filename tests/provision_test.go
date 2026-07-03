/*
Copyright 2019 The OpenEBS Authors

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

package tests

import (
	"fmt"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("[zfspv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with zfs driver", func() {
		It("Running zfs volume Creation Test", volumeCreationTest)
		It("Running zfs volume Creation Test with custom node id", Label("custom-node-id"), volumeCreationTest)
		It("Running encrypted volume creation test", encryptedVolCreationTest)
		It("Running shared volume tests", fsVolCreationWithSharedParameterTest)
	})
})

func fsVolCreationTest() {
	storageClass := getStoragClassParams()
	for _, params := range storageClass {
		exhaustiveVolumeTests(params)
	}
}

func fsVolCreationWithSharedParameterTest() {
	storageClass := getSharedStorageClassParameters()
	for _, params := range storageClass {
		SharedVolumeTests(params)
	}
}

func SharedVolumeTests(parameters map[string]string) {
	sharedParam := parameters["shared"]
	By(fmt.Sprintf("####### Creating the storage class with shared parameter : %s #######", formatParams(parameters)))
	By("Creating Storage Class", func() { createFstypeStorageClass(parameters) })
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameFS) })
	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameFS, pvcNameFS) })
	if sharedParam == "yes" {
		By("Creating and deploying app pod sharing the same PVC", func() { createDeployVerifyApp(appNameFSShared, pvcNameFS) })
	} else {
		By("Creating a second app pod expected to fail mounting the non-shared PVC", func() { createAndDeployAppPod(appNameFSShared, pvcNameFS) })
		By("verifying second app pod never becomes running", func() { verifyAppPodNotRunning(appNameFSShared) })
	}
	By("Deleting main application deployment", func() { deleteAppDeployment(appNameFS) })
	By("Deleting application deployment using shared volume", func() { deleteAppDeployment(appNameFSShared) })
	By("Deleting shared pvc", func() { deletePVC(pvcNameFS) })
	By("Deleting storage class", deleteStorageClass)
}

// Test to cater create, snapshot, clone and delete resources
func exhaustiveVolumeTests(parameters map[string]string) {
	fstype := parameters["fstype"]
	create(parameters)
	snapshotAndCloneCreate()
	volumeCloneCreate()
	// btrfs does not support online resize
	if fstype != "btrfs" {
		By("Resizing the PVC", func() { resizeAndVerifyPVC(pvcNameFS) })
	}
	snapshotAndCloneCleanUp()
	volumeCloneCleanUp()
	cleanUp()
}

func formatParams(parameters map[string]string) string {
	keys := make([]string, 0, len(parameters))
	for k := range parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(parameters))
	for _, k := range keys {
		parts = append(parts, k+"="+parameters[k])
	}
	return strings.Join(parts, ", ")
}

// Creates the resources
func create(parameters map[string]string) {
	By(fmt.Sprintf("####### Creating the storage class : %s #######", formatParams(parameters)))
	createFstypeStorageClass(parameters)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameFS) })
	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameFS, pvcNameFS) })
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying storage class parameters")
	VerifyStorageClassParams(parameters)
}

// Creates the snapshot/clone resources
func snapshotAndCloneCreate() {
	createSnapshot(pvcNameFS, snapNameFS)
	verifySnapshotCreated(snapNameFS)
	createClone(clonePvcNameFS, snapNameFS, scObj.Name, "Filesystem")
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameFS, clonePvcNameFS) })
}

// Removes the snapshot/clone resources
func snapshotAndCloneCleanUp() {
	deleteAppDeployment(cloneAppNameFS)
	deletePVC(clonePvcNameFS)
	deleteSnapshot(pvcNameFS, snapNameFS)
}

// Creates the volume clone resources
func volumeCloneCreate() {
	createVolumeClone(volumeClonePvcNameFS, pvcNameFS, scObj.Name, "Filesystem")
	By("Creating and deploying volume clone app pod", func() {
		createDeployVerifyCloneApp(volumeCloneAppNameFS, volumeClonePvcNameFS)
	})
}

// Removes the volume clone resources
func volumeCloneCleanUp() {
	deleteAppDeployment(volumeCloneAppNameFS)
	deletePVC(volumeClonePvcNameFS)
}

// Removes the resources
func cleanUp() {
	deleteAppDeployment(appNameFS)
	deletePVC(pvcNameFS)
	By("Deleting storage class", deleteStorageClass)
}

func blockVolCreationTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameBlock) })

	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameBlock, pvcNameBlock) })
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying ZFSVolume property change", VerifyZFSVolumePropEdit)

	createSnapshot(pvcNameBlock, snapNameBlock)
	verifySnapshotCreated(snapNameBlock)

	By("creating volume clone from block pvc", func() {
		createVolumeClone(volumeClonePvcNameBlock, pvcNameBlock, scObj.Name, "Block")
	})
	By("Creating and deploying volume clone app pod", func() {
		createDeployVerifyCloneApp(volumeCloneAppNameBlock, volumeClonePvcNameBlock)
	})

	createClone(clonePvcNameBlock, snapNameBlock, scObj.Name, "Block")
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameBlock, clonePvcNameBlock) })

	By("Deleting volume clone application deployment")
	deleteAppDeployment(volumeCloneAppNameBlock)
	By("Deleting volume clone pvc")
	deletePVC(volumeClonePvcNameBlock)

	By("Deleting main application deployment")
	deleteAppDeployment(appNameBlock)

	zvName := getZVName(pvcNameBlock)
	By("Deleting main pvc")
	deletePVC(pvcNameBlock)

	By("Verifying ZFSVolume object after pvc deletion when snapshot is present", VerifyZFSVolume)

	By("Deleting clone application deployment")
	deleteAppDeployment(cloneAppNameBlock)

	By("Deleting snapshot and clone pvc")

	deletePVC(clonePvcNameBlock)
	By("Verifying that ZV is present after pvc deletion ", func() { IsZVPresentConsistently(zvName) })
	deleteSnapshot(pvcNameBlock, snapNameBlock)
	By("Verifying that ZV is deleted after snapshot deletion ", func() { IsZVDeletedEventually(zvName) })

	By("Deleting storage class", deleteStorageClass)
}

func blockVolCreationWithReclaimRetainTest() {
	By("Creating storage class retain reclaim policy", createStorageClassWithReclaimPolicy)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameBlock) })

	By("verifying ZFSVolume object", VerifyZFSVolume)

	createSnapshot(pvcNameBlock, snapNameBlock)
	verifySnapshotCreated(snapNameBlock)

	By("creating volume clone from block pvc", func() {
		createVolumeClone(volumeClonePvcNameBlock, pvcNameBlock, scObj.Name, "Block")
	})
	By("Creating and deploying volume clone app pod", func() {
		createDeployVerifyCloneApp(volumeCloneAppNameBlock, volumeClonePvcNameBlock)
	})
	By("Deleting volume clone application deployment")
	deleteAppDeployment(volumeCloneAppNameBlock)
	By("Deleting volume clone pvc")
	volumeCloneZVName := getZVName(volumeClonePvcNameBlock)
	deletePVC(volumeClonePvcNameBlock)
	// The volume clone uses the Retain reclaim policy SC, so the external-provisioner
	// never calls CSI DeleteVolume. The ZFSVolume CR and ZFS clone data remain on disk.
	// Explicitly delete the ZV so the ZFS clone is destroyed before we try to destroy
	// the parent volume (which would fail with "volume has dependent clones").
	By("Deleting volume clone ZV for cleanup", func() { DeleteZV(volumeCloneZVName) })

	zvName := getZVName(pvcNameBlock)
	By("Deleting main pvc", func() { deletePVC(pvcNameBlock) })

	By("Verifying ZFSVolume object after PVC deletion when snapshot is present", VerifyZFSVolume)

	By("Verifying that ZV is present after PVC deletion ", func() { IsZVPresentConsistently(zvName) })
	By("Deleting snapshot", func() { deleteSnapshot(pvcNameBlock, snapNameBlock) })
	By("Verifying that ZV is present after snapshot deletion ", func() { IsZVPresentConsistently(zvName) })
	deletePV(zvName)
	By("Verifying that ZV is present after PV deletion ", func() { IsZVPresentConsistently(zvName) })
	By("Create and Verifying PV from the retained ZV ", func() {
		createAndVerifyPVFromRetainedZV(pvFromRetainZV, zvName)
	})
	By("creating and verifying PVC bound status from retained ZV", func() { createAndVerifyPVC(pvcFromRetainZV) })

	zvNewName := getZVName(pvcFromRetainZV)
	deletePVC(pvcFromRetainZV)
	deletePV(zvNewName)
	By("Deleting the ZV for cleanup ", func() { DeleteZV(zvName) })
	By("Deleting storage class", deleteStorageClass)

}

func encryptedVolCreationTest() {
	By("Creating encrypted storage class", createEncryptedStorageClass)
	By("creating and verifying PVC bound status", func() { createAndVerifyPVC(pvcNameFS) })

	By("Creating and deploying app pod", func() { createDeployVerifyApp(appNameFS, pvcNameFS) })
	By("verifying ZFSVolume object", VerifyZFSVolume)

	createSnapshot(pvcNameFS, snapNameFS)
	verifySnapshotCreated(snapNameFS)

	createClone(clonePvcNameFS, snapNameFS, scObj.Name, "Filesystem")
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameFS, clonePvcNameFS) })

	By("Deleting clone application deployment")
	deleteAppDeployment(cloneAppNameFS)
	By("Deleting clone pvc")
	deletePVC(clonePvcNameFS)

	By("Deleting snapshot")
	deleteSnapshot(pvcNameFS, snapNameFS)

	By("creating volume clone from encrypted pvc", func() {
		createVolumeClone(volumeClonePvcNameFS, pvcNameFS, scObj.Name, "Filesystem")
	})
	By("Creating and deploying volume clone app pod", func() {
		createDeployVerifyCloneApp(volumeCloneAppNameFS, volumeClonePvcNameFS)
	})
	By("Deleting volume clone application deployment")
	deleteAppDeployment(volumeCloneAppNameFS)
	By("Deleting volume clone pvc")
	deletePVC(volumeClonePvcNameFS)

	By("Deleting main application deployment")
	deleteAppDeployment(appNameFS)
	By("Deleting main pvc")
	deletePVC(pvcNameFS)

	By("Deleting storage class", deleteStorageClass)
}

func volumeCreationTest() {
	By("Running volume creation test", fsVolCreationTest)
	By("Running block volume creation test", blockVolCreationTest)
	By("Running block volume creation test with retain reclaim policy", blockVolCreationWithReclaimRetainTest)

}
