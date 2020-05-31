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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	"github.com/openebs/zfs-localpv/tests/container"
	"github.com/openebs/zfs-localpv/tests/deploy"
	"github.com/openebs/zfs-localpv/tests/k8svolume"
	"github.com/openebs/zfs-localpv/tests/pod"
	"github.com/openebs/zfs-localpv/tests/pts"
	"github.com/openebs/zfs-localpv/tests/pvc"
	"github.com/openebs/zfs-localpv/tests/sc"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func IsPVCBoundEventually(pvcName string) bool {
	return Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 5).
		Should(BeTrue())
}

// IsPVCResizedEventually checks if the pvc is bound or not eventually
func IsPVCResizedEventually(pvcName string, newCapacity string) bool {
	newStorage, err := resource.ParseQuantity(newCapacity)
	if err != nil {
		return false
	}
	return Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		pvcStorage := volume.Status.Capacity[corev1.ResourceName(corev1.ResourceStorage)]
		return pvcStorage == newStorage
	},
		120, 5).
		Should(BeTrue())
}

// IsPodRunningEventually return true if the pod comes to running state
func IsPodRunningEventually(namespace, podName string) bool {
	return Eventually(func() bool {
		p, err := PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		60, 5).
		Should(BeTrue())
}

// IsPropUpdatedEventually checks if the property is updated or not eventually
func IsPropUpdatedEventually(vol *apis.ZFSVolume, prop string, val string) bool {
	return Eventually(func() bool {

		newVal, err := zfs.GetVolumeProperty(vol, prop)
		Expect(err).ShouldNot(HaveOccurred())
		return (newVal == val)
	},
		60, 5).
		Should(BeTrue())
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func IsPVCDeletedEventually(pvcName string) bool {
	return Eventually(func() bool {
		_, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(BeTrue())
}

func createExt4StorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname": POOLNAME,
		"fstype":   "ext4",
	}

	By("building a ext4 storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(ZFSProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(),
		"while building ext4 storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating a ext4 storageclass {%s}", scName)
}

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname": POOLNAME,
	}

	By("building a default storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(ZFSProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating a default storageclass {%s}", scName)
}

func createZfsStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname": POOLNAME,
		"fstype":   "zfs",
	}

	By("building a zfs storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithVolumeExpansion(true).
		WithProvisioner(ZFSProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(),
		"while building zfs storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating a zfs storageclass {%s}", scName)
}

func VerifyZFSVolume() {
	By("fetching zfs volume")
	vol, err := ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", pvcObj.Spec.VolumeName)

	volType := zfs.VOLTYPE_ZVOL
	if scObj.Parameters["fstype"] == zfs.FSTYPE_ZFS {
		volType = zfs.VOLTYPE_DATASET
	}

	By("verifying zfs volume")
	Expect(vol.Spec.PoolName).To(Equal(scObj.Parameters["poolname"]),
		"while checking poolname of zfs volume", pvcObj.Spec.VolumeName)
	Expect(vol.Spec.FsType).To(Equal(scObj.Parameters["fstype"]),
		"while checking fstype of zfs volume", pvcObj.Spec.VolumeName)
	Expect(vol.Spec.VolumeType).To(Equal(volType),
		"while checking Volume type as dataset", pvcObj.Spec.VolumeName)
	Expect(vol.Spec.Capacity).To(Equal(capacity),
		"while checking capacity of zfs volume", pvcObj.Spec.VolumeName)

	// it might fail if we are checking finializer before event is processed by node agent
	Expect(vol.Finalizers[0]).To(Equal(zfs.ZFSFinalizer), "while checking finializer to be set {%s}", pvcObj.Spec.VolumeName)
}

func VerifyZFSVolumePropEdit() {
	By("verifying compression property update")

	By("fetching zfs volume for setting compression=on")
	vol, err := ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val := "on"
	vol.Spec.Compression = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status := IsPropUpdatedEventually(vol, "compression", val)
	Expect(status).To(Equal(true), "while updating compression=on {%s}", vol.Name)

	By("fetching zfs volume for setting compression=off")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "off"
	vol.Spec.Compression = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "compression", val)
	Expect(status).To(Equal(true), "while updating compression=off {%s}", vol.Name)

	By("verifying dedup property update")

	By("fetching zfs volume for setting dedup=on")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "on"
	vol.Spec.Dedup = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "dedup", val)
	Expect(status).To(Equal(true), "while updating dedup=on {%s}", vol.Name)

	By("fetching zfs volume for setting dedup=off")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "off"
	vol.Spec.Dedup = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "dedup", val)
	Expect(status).To(Equal(true), "while updating dedup=off {%s}", vol.Name)

	if vol.Spec.VolumeType == zfs.VOLTYPE_DATASET {
		By("verifying recordsize property update")

		By("fetching zfs volume for setting the recordsize")
		vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

		val = "4096" // 4k
		vol.Spec.RecordSize = val
		vol.Spec.VolBlockSize = "8192"
		_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
		Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

		status = IsPropUpdatedEventually(vol, "recordsize", val)
		Expect(status).To(Equal(true), "while updating redordsize {%s}", vol.Name)
	} else {

		Expect(vol.Spec.VolumeType).To(Equal(zfs.VOLTYPE_ZVOL), "voltype should be zvol {%s}", vol.Name)

		By("verifying blocksize property update")

		By("fetching zfs volume for setting the blocksize")
		vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		Expect(err).To(BeNil(), "while fetching the zfs volume {%s}", vol.Name)

		val, err = zfs.GetVolumeProperty(vol, "volblocksize")
		Expect(err).ShouldNot(HaveOccurred())

		nval := "8192" // 8k
		vol.Spec.VolBlockSize = nval
		vol.Spec.RecordSize = "16384"
		_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
		Expect(err).To(BeNil(), "while updating the zfs volume {%s}", vol.Name)

		status = IsPropUpdatedEventually(vol, "volblocksize", val)
		Expect(status).To(Equal(true), "while updating volblocksize {%s}", vol.Name)
	}
}

func deleteStorageClass() {
	err := SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(),
		"while deleting zfs storageclass {%s}", scObj.Name)
}

func createAndVerifyPVC() {
	var (
		err     error
		pvcName = "zfspv-pvc"
	)
	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	Expect(status).To(Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func createAndVerifyBlockPVC() {
	var (
		err     error
		pvcName = "zfspv-pvc"
	)

	volmode := corev1.PersistentVolumeBlock

	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithVolumeMode(&volmode).
		WithCapacity(capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	Expect(status).To(Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func resizeAndVerifyPVC() {
	var (
		err     error
		pvcName = "zfspv-pvc"
	)
	By("updating the pvc with new size")
	pvcObj, err = pvc.BuildFrom(pvcObj).
		WithCapacity(NewCapacity).Build()
	Expect(err).To(
		BeNil(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Update(pvcObj)
	Expect(err).To(
		BeNil(),
		"while updating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	By("verifying pvc size to be updated")

	status := IsPVCResizedEventually(pvcName, NewCapacity)
	Expect(status).To(Equal(true),
		"while checking pvc resize")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(
		BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}
func createDeployVerifyApp() {
	By("creating and deploying app pod", createAndDeployAppPod)
	time.Sleep(30 * time.Second)
	By("verifying app pod is running", verifyAppPodRunning)
}

func createAndDeployAppPod() {
	var err error
	By("building a busybox app pod deployment using above zfs volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/datadir",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	Expect(err).ShouldNot(HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createAndDeployBlockAppPod() {
	var err error
	By("building a busybox app pod deployment using above zfs volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeDevicesNew(
							[]corev1.VolumeDevice{
								corev1.VolumeDevice{
									Name:       "datavol1",
									DevicePath: "/dev/xvda",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	Expect(err).ShouldNot(HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createDeployVerifyBlockApp() {
	By("creating and deploying app pod", createAndDeployBlockAppPod)
	time.Sleep(30 * time.Second)
	By("verifying app pod is running", verifyAppPodRunning)
}

func verifyAppPodRunning() {
	var err error
	appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	Expect(err).ShouldNot(HaveOccurred(), "while verifying application pod")

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	Expect(status).To(Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func deleteAppDeployment() {
	err := DeployClient.WithNamespace(OpenEBSNamespace).
		Delete(deployObj.Name, &metav1.DeleteOptions{})
	Expect(err).ShouldNot(HaveOccurred(), "while deleting application pod")
}

func deletePVC() {
	err := PVCClient.WithNamespace(OpenEBSNamespace).Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
	By("verifying deleted pvc")
	status := IsPVCDeletedEventually(pvcName)
	Expect(status).To(Equal(true), "while trying to get deleted pvc")

}
