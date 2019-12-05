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

func datasetCreationTest() {
	By("Creating zfs storage class", createZfsStorageClass)
	By("creating and verifying PVC bound status", createAndVerifyPVC)
	By("Creating and deploying app pod", createDeployVerifyApp)
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying ZFSVolume property change", VerifyZFSVolumePropEdit)
	By("Deleting application deployment", deleteAppDeployment)
	By("Deleting pvc", deletePVC)
	By("Deleting storage class", deleteStorageClass)
}

func zvolCreationTest() {
	By("Creating ext4 storage class", createExt4StorageClass)
	By("creating and verifying PVC bound status", createAndVerifyPVC)

	/*
	 * commenting app deployment as provisioning is taking time
	 * since we are creating a zfs pool on a sparse file and mkfs
	 * is taking forever for zvol.
	 * Should create the zfs pool on the disk. Need to check if travis
	 * has that functionality.
	 */
	//By("Creating and deploying app pod", createDeployVerifyApp)
	By("verifying ZFSVolume object", VerifyZFSVolume)
	By("verifying ZFSVolume property change", VerifyZFSVolumePropEdit)
	//By("Deleting application deployment", deleteAppDeployment)
	By("Deleting pvc", deletePVC)
	By("Deleting storage class", deleteStorageClass)
}

func volumeCreationTest() {
	By("Running dataset creation test", datasetCreationTest)
	By("Running zvol creation test", zvolCreationTest)
}
