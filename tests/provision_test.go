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
		By("Resizing the PVC", resizeAndVerifyPVC)
	}
	snapshotAndCloneCleanUp()
	cleanUp()
}

// Creates the resources
func create(parameters map[string]string) {
	By("####### Creating the storage class : " + parameters["fstype"] + " #######")
	createFstypeStorageClass(parameters)
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Creating and deploying app pod", createDeployVerifyApp)
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying storage class parameters")
	VerifyStorageClassParams(parameters)
}

// Creates the snapshot/clone resources
func snapshotAndCloneCreate() {
	createSnapshot(pvcName, snapName)
	verifySnapshotCreated(snapName)
	createClone(clonePvcName, snapName, scObj.Name)
	By("Creating and deploying clone app pod", createDeployVerifyCloneApp)
}

// Removes the snapshot/clone resources
func snapshotAndCloneCleanUp() {
	deleteAppDeployment(cloneAppName)
	deletePVC(clonePvcName)
	deleteSnapshot(pvcName, snapName)
}

// Removes the resources
func cleanUp() {
	deleteAppDeployment(appName)
	deletePVC(pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func blockVolCreationTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status", createAndVerifyBlockPVC)

	By("Creating and deploying app pod", createDeployVerifyBlockApp)
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying ZFSVolume property change", VerifyZFSVolumePropEdit)
	By("Deleting application deployment")

	createSnapshot(pvcName, snapName)
	verifySnapshotCreated(snapName)
	createClone(clonePvcName, snapName, scObj.Name)
	By("Creating and deploying clone app pod", createDeployVerifyCloneApp)

	By("Deleting clone and main application deployment")
	deleteAppDeployment(cloneAppName)
	deleteAppDeployment(appName)

	By("Deleting snapshot, main pvc and clone pvc")
	deletePVC(clonePvcName)
	deleteSnapshot(pvcName, snapName)
	deletePVC(pvcName)

	By("Deleting storage class", deleteStorageClass)
}

func volumeCreationTest() {
	By("Running volume creation test", fsVolCreationTest)
	By("Running block volume creation test", blockVolCreationTest)
}
