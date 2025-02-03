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
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("[zfspv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with zfs driver", func() {
		It("Running zfs volume Creation Test", volumeCreationTest)
		It("Running zfs volume Creation Test with custom node id", Label("custom-node-id"), volumeCreationTest)
	})
})

func fsVolCreationTest() {
	storageClass := getStoragClassParams()
	for _, params := range storageClass {
		exhaustiveVolumeTests(params)
	}
}

// Test to cater create, snapshot, clone and delete resources
func exhaustiveVolumeTests(parameters map[string]string) {
	fstype := parameters["fstype"]
	create(parameters)
	snapshotAndCloneCreate()
	// btrfs does not support online resize
	if fstype != "btrfs" {
		By("Resizing the PVC", func() { resizeAndVerifyPVC(pvcNameFS) })
	}
	snapshotAndCloneCleanUp()
	cleanUp()
}

// Creates the resources
func create(parameters map[string]string) {
	By("####### Creating the storage class : " + parameters["fstype"] + " #######")
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
	createClone(clonePvcNameFS, snapNameFS, scObj.Name)
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameFS, clonePvcNameFS) })
}

// Removes the snapshot/clone resources
func snapshotAndCloneCleanUp() {
	deleteAppDeployment(cloneAppNameFS)
	deletePVC(clonePvcNameFS)
	deleteSnapshot(pvcNameFS, snapNameFS)
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
	createClone(clonePvcNameBlock, snapNameBlock, scObj.Name)
	By("Creating and deploying clone app pod", func() { createDeployVerifyCloneApp(cloneAppNameBlock, clonePvcNameBlock) })

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

func volumeCreationTest() {
	By("Running volume creation test", fsVolCreationTest)
	By("Running block volume creation test", blockVolCreationTest)
	By("Running block volume creation test with retain reclaim policy", blockVolCreationWithReclaimRetainTest)

}
