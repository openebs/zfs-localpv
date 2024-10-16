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

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

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
	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPVCResizedEventually checks if the pvc is bound or not eventually
func IsPVCResizedEventually(pvcName string, newCapacity string) bool {
	newStorage, err := resource.ParseQuantity(newCapacity)
	if err != nil {
		return false
	}
	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		pvcStorage := volume.Status.Capacity[corev1.ResourceName(corev1.ResourceStorage)]
		return pvcStorage == newStorage
	},
		120, 5).
		Should(gomega.BeTrue())
}

// IsPodRunningEventually return true if the pod comes to running state
func IsPodRunningEventually(namespace, podName string) bool {
	return gomega.Eventually(func() bool {
		p, err := PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPropUpdatedEventually checks if the property is updated or not eventually
func IsPropUpdatedEventually(vol *apis.ZFSVolume, prop string, val string) bool {
	return gomega.Eventually(func() bool {
		newVal, err := zfs.GetVolumeProperty(vol, prop)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return (newVal == val)
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func IsPVCDeletedEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		_, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(gomega.BeTrue())
}

// VerifyStorageClassParams verifies the volume properties set at creation time
func VerifyStorageClassParams(property map[string]string) {
	vol, err := ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)
	// Check for file system type
	if property["fstype"] == "zfs" {
		property["type"] = "filesystem"
	} else {
		property["type"] = "volume"
	}
	generateThinProvisionParams(property)
	delete(property, "fstype")

	for propertyKey, propertyVal := range property {
		status := IsPropUpdatedEventually(vol, propertyKey, propertyVal)
		gomega.Expect(status).To(gomega.Equal(true), "while updating {%s%}={%s%} {%s}", propertyKey, propertyVal, vol.Name)
	}

}

// It populates the map for thing provisioning params
// Refer https://github.com/openebs/zfs-localpv/issues/560#issuecomment-2232535073
func generateThinProvisionParams(property map[string]string) {
	if property["fstype"] == "zfs" {
		property["quota"] = capacity
		property["reservation"] = defaultReservation
		if property["thinprovision"] == "no" {
			property["reservation"] = capacity
		}
	} else {
		property["quota"] = "-"
		property["reservation"] = defaultReservation
	}
	delete(property, "thinprovision")
}

func createFstypeStorageClass(addons map[string]string) {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname": POOLNAME,
	}

	// Update params with addons
	for key, value := range addons {
		parameters[key] = value
	}

	ginkgo.By("building a " + addons["ftype"] + " storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(ZFSProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building ext4 storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a ext4 storageclass {%s}", scName)
}

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname": POOLNAME,
	}

	ginkgo.By("building a default storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(ZFSProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
}

// VerifyZFSVolume verify the properties of a zfs-volume
func VerifyZFSVolume() {
	ginkgo.By("fetching zfs volume")
	vol, err := ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", pvcObj.Spec.VolumeName)

	volType := zfs.VolTypeZVol
	if scObj.Parameters["fstype"] == zfs.FSTypeZFS {
		volType = zfs.VolTypeDataset
	}

	ginkgo.By("verifying zfs volume")
	gomega.Expect(vol.Spec.PoolName).To(gomega.Equal(scObj.Parameters["poolname"]),
		"while checking poolname of zfs volume", pvcObj.Spec.VolumeName)
	gomega.Expect(vol.Spec.FsType).To(gomega.Equal(scObj.Parameters["fstype"]),
		"while checking fstype of zfs volume", pvcObj.Spec.VolumeName)
	gomega.Expect(vol.Spec.VolumeType).To(gomega.Equal(volType),
		"while checking Volume type as dataset", pvcObj.Spec.VolumeName)
	gomega.Expect(vol.Spec.Capacity).To(gomega.Equal(capacity),
		"while checking capacity of zfs volume", pvcObj.Spec.VolumeName)

	// it might fail if we are checking finializer before event is processed by node agent
	gomega.Expect(vol.Finalizers[0]).To(gomega.Equal(zfs.ZFSFinalizer), "while checking finializer to be set {%s}", pvcObj.Spec.VolumeName)
}

// VerifyZFSVolumePropEdit verifies the volume properties
func VerifyZFSVolumePropEdit() {
	ginkgo.By("verifying compression property update")

	ginkgo.By("fetching zfs volume for setting compression=on")
	vol, err := ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val := "on"
	vol.Spec.Compression = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status := IsPropUpdatedEventually(vol, "compression", val)
	gomega.Expect(status).To(gomega.Equal(true), "while updating compression=on {%s}", vol.Name)

	ginkgo.By("fetching zfs volume for setting compression=off")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "off"
	vol.Spec.Compression = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "compression", val)
	gomega.Expect(status).To(gomega.Equal(true), "while updating compression=off {%s}", vol.Name)

	ginkgo.By("verifying dedup property update")

	ginkgo.By("fetching zfs volume for setting dedup=on")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "on"
	vol.Spec.Dedup = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "dedup", val)
	gomega.Expect(status).To(gomega.Equal(true), "while updating dedup=on {%s}", vol.Name)

	ginkgo.By("fetching zfs volume for setting dedup=off")
	vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

	val = "off"
	vol.Spec.Dedup = val
	_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
	gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

	status = IsPropUpdatedEventually(vol, "dedup", val)
	gomega.Expect(status).To(gomega.Equal(true), "while updating dedup=off {%s}", vol.Name)

	if vol.Spec.VolumeType == zfs.VolTypeDataset {
		ginkgo.By("verifying recordsize property update")

		ginkgo.By("fetching zfs volume for setting the recordsize")
		vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

		val = "4096" // 4k
		vol.Spec.RecordSize = val
		vol.Spec.VolBlockSize = "8192"
		_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
		gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

		status = IsPropUpdatedEventually(vol, "recordsize", val)
		gomega.Expect(status).To(gomega.Equal(true), "while updating redordsize {%s}", vol.Name)
	} else {

		gomega.Expect(vol.Spec.VolumeType).To(gomega.Equal(zfs.VolTypeZVol), "voltype should be zvol {%s}", vol.Name)

		ginkgo.By("verifying blocksize property update")

		ginkgo.By("fetching zfs volume for setting the blocksize")
		vol, err = ZFSClient.WithNamespace(OpenEBSNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		gomega.Expect(err).To(gomega.BeNil(), "while fetching the zfs volume {%s}", vol.Name)

		val, err = zfs.GetVolumeProperty(vol, "volblocksize")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		nval := "8192" // 8k
		vol.Spec.VolBlockSize = nval
		vol.Spec.RecordSize = "16384"
		_, err = ZFSClient.WithNamespace(OpenEBSNamespace).Update(vol)
		gomega.Expect(err).To(gomega.BeNil(), "while updating the zfs volume {%s}", vol.Name)

		status = IsPropUpdatedEventually(vol, "volblocksize", val)
		gomega.Expect(status).To(gomega.Equal(true), "while updating volblocksize {%s}", vol.Name)
	}
}

func deleteStorageClass() {
	err := SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	gomega.Expect(err).To(gomega.BeNil(),
		"while deleting zfs storageclass {%s}", scObj.Name)
}

func createAndVerifyPVC() {
	var (
		err     error
		pvcName = "zfspv-pvc"
	)
	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
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

	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithVolumeMode(&volmode).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
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
	ginkgo.By("updating the pvc with new size")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	pvcObj, err = pvc.BuildFrom(pvcObj).
		WithCapacity(NewCapacity).Build()
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Update(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while updating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc size to be updated")

	status := IsPVCResizedEventually(pvcName, NewCapacity)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking pvc resize")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}
func createDeployVerifyApp() {
	ginkgo.By("creating and deploying app pod")
	createAndDeployAppPod(appName)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func createDeployVerifyCloneApp() {
	ginkgo.By("creating and deploying app pod")
	createAndDeployAppPod(cloneAppName)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func createAndDeployAppPod(appname string) {
	var err error
	ginkgo.By("building a busybox app pod deployment using above zfs volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appname).
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

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createAndDeployBlockAppPod() {
	var err error
	ginkgo.By("building a busybox app pod deployment using above zfs volume")
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

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createDeployVerifyBlockApp() {
	ginkgo.By("creating and deploying app pod", createAndDeployBlockAppPod)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func verifyAppPodRunning() {
	var err error
	appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func deleteAppDeployment(appname string) {
	err := DeployClient.WithNamespace(OpenEBSNamespace).
		Delete(appname, &metav1.DeleteOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while deleting application pod")
}

func deletePVC(pvcname string) {
	err := PVCClient.WithNamespace(OpenEBSNamespace).Delete(pvcname, &metav1.DeleteOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcname,
		OpenEBSNamespace,
	)
	ginkgo.By("verifying deleted pvc")
	status := IsPVCDeletedEventually(pvcname)
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted pvc")
}

func getStoragClassParams() []map[string]string {
	return []map[string]string{
		{
			"fstype":      "zfs",
			"compression": "lz4",
		},
		{
			"fstype":      "zfs",
			"compression": "lz4",
			"dedup":       "on",
		},
		{
			"fstype":        "zfs",
			"compression":   "zstd-fast",
			"dedup":         "on",
			"thinprovision": "yes",
		},
		{
			"fstype": "zfs",
			"dedup":  "on",
		},
		{
			"fstype":        "zfs",
			"compression":   "gzip",
			"thinprovision": "yes",
		},
		{
			"fstype":        "zfs",
			"compression":   "gzip",
			"dedup":         "on",
			"thinprovision": "yes",
		},
		{
			"fstype":      "xfs",
			"compression": "zstd-6",
		},
		{
			"fstype":        "xfs",
			"compression":   "zstd-6",
			"dedup":         "on",
			"thinprovision": "yes",
		},
		{
			"fstype": "ext4",
			"dedup":  "on",
		},
		{
			"fstype":      "btrfs",
			"compression": "zstd-6",
			"dedup":       "on",
		},
		{
			"fstype":      "xfs",
			"compression": "zstd-fast",
		},
	}
}
