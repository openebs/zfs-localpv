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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/zfs-localpv/pkg/builder/volbuilder"
	"github.com/openebs/zfs-localpv/tests/deploy"
	"github.com/openebs/zfs-localpv/tests/pod"
	"github.com/openebs/zfs-localpv/tests/pvc"
	"github.com/openebs/zfs-localpv/tests/sc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// zfs pool name where volume provisioning will happen
	POOLNAME = "zfspv-pool"
)

var (
	ZFSClient      *volbuilder.Kubeclient
	SCClient       *sc.Kubeclient
	PVCClient      *pvc.Kubeclient
	DeployClient   *deploy.Kubeclient
	PodClient      *pod.KubeClient
	nsName         = "zfspv-provision"
	scName         = "zfspv-sc"
	ZFSProvisioner = "zfs.csi.openebs.io"
	pvcName        = "zfspv-pvc"
	snapName       = "zfspv-snap"
	appName        = "busybox-zfspv"
	clonePvcName   = "zfspv-pvc-clone"
	cloneAppName   = "busybox-zfspv-clone"

	scObj            *storagev1.StorageClass
	deployObj        *appsv1.Deployment
	pvcObj           *corev1.PersistentVolumeClaim
	appPod           *corev1.PodList
	accessModes      = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity         = "5368709120" // 5Gi
	NewCapacity      = "8Gi"        // 8Gi, for testing resize
	KubeConfigPath   string
	OpenEBSNamespace string
)

func init() {
	KubeConfigPath = os.Getenv("KUBECONFIG")

	OpenEBSNamespace = os.Getenv("OPENEBS_NAMESPACE")
	if OpenEBSNamespace == "" {
		klog.Fatalf("OPENEBS_NAMESPACE environment variable not set")
	}
	SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(KubeConfigPath))
	PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(KubeConfigPath))
	DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(KubeConfigPath))
	PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(KubeConfigPath))
	ZFSClient = volbuilder.NewKubeclient(volbuilder.WithKubeConfigPath(KubeConfigPath))
}

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test ZFSPV volume provisioning")
}
