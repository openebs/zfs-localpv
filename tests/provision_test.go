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
	. "github.com/onsi/ginkgo"
)

var _ = Describe("[zfspv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with zfs driver", func() {
		It("Running zfs volume Creation Test", volumeCreationTest)
	})
})

func fsVolCreationTest() {
	///
	fstypes := []string{"zfs", "ext4", "xfs", "btrfs"}
	for _, fstype := range fstypes {
		By("####### Creating the storage class : " + fstype + " #######")
		parameters := getStoragClassParams(fstype, "", "")
		exhaustiveVolumeTests(fstype, parameters)

		// verify the different compression type
		compression := []string{"on", "lzjb", "zstd-19", "gzip-9", "zle", "lz4", "off"}
		for _, compressionValue := range compression {
			By("####### Creating the storage class : " + fstype + " and compression as " + compressionValue + " #######")
			parameters = getStoragClassParams(fstype, "compression", compressionValue)
			minimalVolumeTest("compression", compressionValue, parameters)
		}

		// verify the ddedup functionality
		dedup := []string{"on", "off"}
		for _, dedupValue := range dedup {
			By("####### Creating the storage class : " + fstype + " and dedup as " + dedupValue + " #######")
			parameters = getStoragClassParams(fstype, "dedup", dedupValue)
			minimalVolumeTest("dedup", dedupValue, parameters)
		}
	}
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

// Test to cater create and delete resources
func minimalVolumeTest(property, property_value string, parameters map[string]string) {
	create(parameters)
	VerifyStorageClassParams(property, property_value)
	cleanUp()
}

// Returns a map to be consumed by storage class
func getStoragClassParams(fstype, key, value string) map[string]string {
	parameters := map[string]string{
		"poolname": POOLNAME,
		"fstype":   fstype,
	}
	if len(key) != 0 && len(value) != 0 {
		parameters[key] = value
	}
	return parameters
}

// Test to cater create, snapshot, clone and delete resources
func exhaustiveVolumeTests(fstype string, parameters map[string]string) {
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
	createFstypeStorageClass(parameters)
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Creating and deploying app pod", createDeployVerifyApp)
	By("verifying ZFSVolume object", VerifyZFSVolume)
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
