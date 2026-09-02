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
	"os/exec"

	"k8s.io/klog/v2"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/openebs/zfs-localpv/pkg/zfs"
	"github.com/openebs/zfs-localpv/tests/container"
	"github.com/openebs/zfs-localpv/tests/deploy"
	"github.com/openebs/zfs-localpv/tests/k8svolume"
	"github.com/openebs/zfs-localpv/tests/pod"
	"github.com/openebs/zfs-localpv/tests/pts"
	"github.com/openebs/zfs-localpv/tests/pv"
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

// IsPVAvailableEventually checks if the pv is bound or not eventually
func IsPVAvailableEventually(pvName string) bool {
	return gomega.Eventually(func() bool {
		volume, err := PVClient.
			Get(pvName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pv.NewForAPIObject(volume).IsAvailable()
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
		newVal, err := GetVolumeProperty(vol, prop)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return (newVal == val)
	},
		60, 5).
		Should(gomega.BeTrue())
}

// GetVolumeProperty gets zfs properties for the volume
func GetVolumeProperty(vol *apis.ZFSVolume, prop string) (string, error) {
	var ZFSVolArg []string
	volume := vol.Spec.PoolName + "/" + vol.Name

	ZFSVolArg = append(ZFSVolArg, zfs.ZFSVolCmd, zfs.ZFSGetArg, "-pH", "-o", "value", prop, volume)
	cmd := exec.Command("sudo", ZFSVolArg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("zfs: could not get %s on dataset %v cmd %v error: %s",
			prop, volume, ZFSVolArg, string(out))
		return "", fmt.Errorf("zfs get %s failed, %s", prop, string(out))
	}
	val := out[:len(out)-1]
	return string(val), nil
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func IsPVCDeletedEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		_, err := PVCClient.WithNamespace(OpenEBSNamespace).
			Get(pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(gomega.BeTrue())
}

// IsPVDeletedEventually tries to get the deleted pv
// and returns true if pv is not found
// else returns false
func IsPVDeletedEventually(pvName string) bool {
	return gomega.Eventually(func() bool {
		_, err := PVClient.
			Get(pvName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(gomega.BeTrue())
}

// IsZVDeletedEventually tries to get the deleted zv
// and returns true if zv is not found
// else returns false
func IsZVDeletedEventually(zvName string) bool {
	return gomega.Eventually(func() bool {
		_, err := ZFSClient.WithNamespace(OpenEBSNamespace).
			Get(zvName, metav1.GetOptions{})
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
		if property["quotatype"] == "quota" {
			property["quota"] = string(capacity)
			if property["thinprovision"] == "no" {
				property["reservation"] = string(capacity)
			}
		}
		if property["quotatype"] == "refquota" {
			property["refquota"] = string(capacity)
			if property["thinprovision"] == "no" {
				property["refreservation"] = string(capacity)
			}
		}
		delete(property, "quotatype")
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

func createStorageClassWithReclaimPolicy() {
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
		WithProvisioner(ZFSProvisioner).
		WithReclaimPolicy(&RetainReclaimPolicy).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
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

func createEncryptedStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"poolname":    POOLNAME,
		"encryption":  "aes-256-gcm",
		"keyformat":   "raw",
		"keylocation": "file:///etc/zfs/keys/zfspool0.key",
		"fstype":      "zfs",
		"compression": "lz4",
	}

	ginkgo.By("building an encrypted storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(ZFSProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building encrypted storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating an encrypted storageclass {%s}", scName)
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

func createAndVerifyPVC(pvcName string) {
	var (
		err error
	)
	ginkgo.By("building a pvc " + pvcName)
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()

	if pvcName == "zfspv-pvc-block" {
		volmode := corev1.PersistentVolumeBlock
		pvcObj.Spec.VolumeMode = &volmode
	}

	if pvcName == "pvc-from-retain-zv" {
		volmode := corev1.PersistentVolumeBlock
		pvcObj.Spec.VolumeMode = &volmode
		pvcObj.Spec.VolumeName = "pv-from-retain-zv"
	}

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

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
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

func resizeAndVerifyPVC(pvcName string) {
	var (
		err error
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
func createDeployVerifyApp(appName, pvcName string) {
	ginkgo.By("creating and deploying app pod")
	if pvcName == "zfspv-pvc-block" || pvcName == "pvc-name-for-del-test" {
		createAndDeployBlockAppPod(appName, pvcName)
	} else {
		createAndDeployAppPod(appName, pvcName)
	}
	ginkgo.By("verifying app pod is running", func() { verifyAppPodRunning(appName) })
}

func createDeployVerifyCloneApp(cloneAppName, clonePvcName string) {
	ginkgo.By("creating and deploying app pod")
	if clonePvcName == "zfspv-pvc-clone-block" || clonePvcName == "zfspv-pvc-vol-clone-block" {
		createAndDeployBlockAppPod(cloneAppName, clonePvcName)
	} else {
		createAndDeployAppPod(cloneAppName, clonePvcName)
	}
	ginkgo.By("verifying app pod is running", func() { verifyAppPodRunning(cloneAppName) })
}

func createAndDeployAppPod(appname string, pvcname string) {
	var err error
	labels := map[string]string{
		"app":     "busybox",
		"role":    "test",
		"appName": appname,
	}
	ginkgo.By("building a busybox app pod deployment using above zfs volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appname).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(labels).
		WithSelectorMatchLabelsNew(labels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labels).
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
						WithPVCSource(pvcname),
				).
				WithTerminationGracePeriodSeconds(5),
		).
		Build()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appname)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appname,
		OpenEBSNamespace,
	)
}

/*
 * verifyFormatOptions runs a pod that compares the block size of the mounted
 * filesystem with the one asked for with formatOptions in the storage class.
 * The pod exits 0 on a match and 1 otherwise and is not restarted, so its phase
 * carries the result: a wrong block size means the options never reached mkfs.
 */
func verifyFormatOptions(podName, pvcName, blockSize string) {
	check := fmt.Sprintf(
		"got=$(stat -fc %%s /mnt/datadir); "+
			"echo \"block size $got, expected %s\"; "+
			"test \"$got\" = \"%s\"",
		blockSize, blockSize,
	)

	checkerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: OpenEBSNamespace,
			Labels:    map[string]string{"role": "test", "appName": podName},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            "busybox",
					Image:           "busybox",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"sh", "-c", check},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "datavol1", MountPath: "/mnt/datadir"},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "datavol1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
		},
	}

	_, err := PodClient.WithNamespace(OpenEBSNamespace).Create(checkerPod)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		podName,
		OpenEBSNamespace,
	)

	var phase corev1.PodPhase
	gomega.Eventually(func() corev1.PodPhase {
		p, err := PodClient.WithNamespace(OpenEBSNamespace).Get(podName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while getting pod {%s}", podName)

		phase = p.Status.Phase

		return phase
	},
		300, 5).
		Should(gomega.Or(gomega.Equal(corev1.PodSucceeded), gomega.Equal(corev1.PodFailed)),
			"while waiting for pod {%s} to finish", podName)

	gomega.Expect(phase).To(gomega.Equal(corev1.PodSucceeded),
		"while checking the block size of the volume of pvc {%s}, the expected format options did not reach mkfs",
		pvcName)
}

// deleteFormatOptionsCheckerPod removes the pod of verifyFormatOptions
func deleteFormatOptionsCheckerPod(podName string) {
	err := PodClient.WithNamespace(OpenEBSNamespace).Delete(podName, &metav1.DeleteOptions{})
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while deleting pod {%s} in namespace {%s}",
		podName,
		OpenEBSNamespace,
	)
}

func createAndDeployBlockAppPod(appName, pvcName string) {
	var err error
	labels := map[string]string{
		"app":     "busybox",
		"role":    "test",
		"appName": appName,
	}
	ginkgo.By("building a busybox app pod deployment using above zfs volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(labels).
		WithSelectorMatchLabelsNew(labels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labels).
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
						WithPVCSource(pvcName),
				).
				WithTerminationGracePeriodSeconds(5),
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

func verifyAppPodRunning(appname string) {
	var err error
	gomega.Eventually(func() bool {
		appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
			List(metav1.ListOptions{
				LabelSelector: "role=test,appName=" + appname,
			})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")
		return len(appPod.Items) == 1
	},
		60, 5).
		Should(gomega.BeTrue())

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func verifyAppPodNotRunning(appname string) {
	var err error
	gomega.Eventually(func() bool {
		appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
			List(metav1.ListOptions{LabelSelector: "role=test,appName=" + appname})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return len(appPod.Items) == 1
	}, 60, 5).Should(gomega.BeTrue())

	status := IsPodNotRunningConsistently(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true),
		"pod %s should never become running when shared=no", appPod.Items[0].Name)
}

// IsPodNotRunningConsistently returns true only if the pod stays out of Running for the window
func IsPodNotRunningConsistently(namespace, podName string) bool {
	return gomega.Consistently(func() bool {
		p, err := PodClient.WithNamespace(namespace).Get(podName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pod.NewForAPIObject(p).IsRunning()
	}, 30, 5).Should(gomega.BeFalse())
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

func deletePV(pvName string) {
	err := PVClient.Delete(pvName, &metav1.DeleteOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting pv {%s} ",
		pvName,
	)
	ginkgo.By("verifying deleted pv")
	status := IsPVDeletedEventually(pvName)
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted pv")
}

// DeleteZV deletes the zv
func DeleteZV(zvName string) {
	err := ZFSClient.WithNamespace(OpenEBSNamespace).Delete(zvName)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting zv {%s} in namespace {%s}",
		zvName,
		OpenEBSNamespace,
	)
	ginkgo.By("verifying deleted zv")
	status := IsZVDeletedEventually(zvName)
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted zv")
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
		{
			"fstype":    "zfs",
			"quotatype": "quota",
		},
		{
			"fstype":    "zfs",
			"quotatype": "refquota",
		},
		{
			"fstype":        "zfs",
			"thinprovision": "no",
			"quotatype":     "refquota",
		},
		{
			"fstype":      "zfs",
			"encryption":  "aes-256-gcm",
			"keyformat":   "raw",
			"keylocation": "file:///etc/zfs/keys/zfspool0.key",
			"compression": "lz4",
		},
	}
}

func getSharedStorageClassParameters() []map[string]string {
	return []map[string]string{
		{
			"fstype": "zfs",
			"shared": "yes",
		},
		{
			"fstype": "ext4",
			"shared": "yes",
		},
		{
			"fstype": "xfs",
			"shared": "yes",
		},
		{
			"fstype": "btrfs",
			"shared": "yes",
		},
		{
			"fstype": "zfs",
			"shared": "no",
		},
	}

}

// IsZVPresentConsistently checks if the zfs volume is present or not consistently
func IsZVPresentConsistently(zvName string) bool {
	return gomega.Consistently(func() bool {
		volume, err := ZFSClient.WithNamespace(OpenEBSNamespace).Get(zvName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return volume.Name == zvName
	},
		30, 5). // Check consistency for 60 seconds, polling every 5 seconds
		Should(gomega.BeTrue())
}

// getZVName return the zv name
func getZVName(pvcName string) string {
	pvcObj, _ = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcName, metav1.GetOptions{})
	return pvcObj.Spec.VolumeName

}

func createAndVerifyPVFromRetainedZV(pvName, volumeHandle string) {
	var (
		err error
	)

	ginkgo.By("building a pv from retained zv")
	pvObj, err := pv.NewBuilder().
		WithName(pvName).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()

	source := corev1.CSIPersistentVolumeSource{Driver: ZFSProvisioner, VolumeHandle: volumeHandle}
	volmode := corev1.PersistentVolumeBlock
	pvObj.Spec.VolumeMode = &volmode
	pvObj.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimRetain
	pvObj.Spec.CSI = &source

	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pv {%s} ",
		pvName,
	)

	ginkgo.By("creating above pv")
	pvObj, err = PVClient.Create(pvObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pv {%s} ",
		pvName,
	)

	ginkgo.By("verifying pvc status as bound ")

	pvObj, err = PVClient.Get(pvName, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s}",
		pvName,
	)

	status := IsPVAvailableEventually(pvName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to available")

	pvObj, err = PVClient.Get(pvName, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s}",
		pvName,
	)
}
