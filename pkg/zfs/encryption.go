/*
Copyright 2024 The OpenEBS Authors

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

package zfs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	apis "github.com/openebs/zfs-localpv/v2/pkg/apis/openebs.io/zfs/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// ZFSLoadKeyArg is the zfs sub-command used to load an encryption key.
const ZFSLoadKeyArg = "load-key"

// PVCEncryptionSecretAnnotation is the PVC annotation naming the Secret that holds
// the volume key.
const PVCEncryptionSecretAnnotation = "encryption.openebs.io/secret-name"

var (
	kubeClientMu sync.Mutex
	kubeClient   kubernetes.Interface
)

// getKubeClient returns a cached Kubernetes clientset. A failed attempt is not
// cached, so it is retried on the next call.
func getKubeClient() (kubernetes.Interface, error) {
	kubeClientMu.Lock()
	defer kubeClientMu.Unlock()
	if kubeClient != nil {
		return kubeClient, nil
	}
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return nil, fmt.Errorf("zfs: failed to build kubeconfig: %w", err)
	}
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("zfs: failed to build kube client: %w", err)
	}
	kubeClient = c
	return kubeClient, nil
}

// kubeAPITimeout bounds the API calls of the node paths that have no request
// context, such as the ZFSVolume workers.
const kubeAPITimeout = 30 * time.Second

// apiContext returns a context bounded by kubeAPITimeout.
func apiContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), kubeAPITimeout)
}

// GetPVC returns the PersistentVolumeClaim identified by namespace/name.
func GetPVC(ctx context.Context, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	client, err := getKubeClient()
	if err != nil {
		return nil, err
	}
	return client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UsesManagedKey reports whether vol's key comes from a Secret rather than a
// keylocation file.
func UsesManagedKey(vol *apis.ZFSVolume) bool {
	return vol.Spec.EncryptionKeyRef != nil
}

// ValidateEncryptionSecret checks that the Secret exists and holds a valid key.
func ValidateEncryptionSecret(ctx context.Context, name, namespace string) error {
	_, err := fetchKeyFromSecret(ctx, namespace, name)
	return err
}

// FetchEncryptionKey returns vol's key, read from its Secret.
func FetchEncryptionKey(ctx context.Context, vol *apis.ZFSVolume) (string, error) {
	ref := vol.Spec.EncryptionKeyRef
	if ref == nil {
		return "", fmt.Errorf("zfs: no encryption key secret referenced")
	}
	return fetchKeyFromSecret(ctx, ref.Namespace, ref.Name)
}

// getDatasetProperty returns a zfs property of any dataset.
func getDatasetProperty(dataset, prop string) (string, error) {
	cmd := exec.Command(ZFSVolCmd, ZFSGetArg, "-pH", "-o", "value", prop, dataset)
	out, err := runCmd(cmd, dataset)
	if err != nil {
		return "", NewZFSError(fmt.Sprintf("zfs get %s", prop), dataset, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// loadKeyForDataset loads vol's key on the encryption root of dataset, unless it
// is already loaded or the dataset is not encrypted.
func loadKeyForDataset(ctx context.Context, vol *apis.ZFSVolume, dataset string) error {
	keystatus, err := getDatasetProperty(dataset, "keystatus")
	if err != nil {
		return err
	}
	switch keystatus {
	case "", "none":
		// not an encrypted dataset
		return nil
	case "available":
		// key already loaded
		return nil
	}

	// keystatus == "unavailable": load the key on the encryption root.
	root, err := getDatasetProperty(dataset, "encryptionroot")
	if err != nil {
		return err
	}
	if root == "" || root == "-" {
		return nil
	}

	key, err := FetchEncryptionKey(ctx, vol)
	if err != nil {
		return err
	}

	// zfs load-key reads the key from stdin for keylocation=prompt datasets.
	cmd := exec.Command(ZFSVolCmd, ZFSLoadKeyArg, root)
	cmd.Stdin = strings.NewReader(key + "\n")
	out, err := runCmd(cmd, root)
	if err != nil {
		// a concurrent load-key (clones, mounts after a reboot) may have won the race
		if status, serr := getDatasetProperty(root, "keystatus"); serr == nil && status == "available" {
			klog.Infof("zfs: encryption key already loaded for %s", root)
			return nil
		}
		ref := vol.Spec.EncryptionKeyRef
		zerr := NewZFSError("zfs load-key", root, err, out)
		klog.Errorf("zfs: could not load the encryption key of %s from Secret %s/%s: %s", root, ref.Namespace, ref.Name, zerr)
		return fmt.Errorf("zfs: could not load the encryption key of %s from Secret %s/%s: %w", root, ref.Namespace, ref.Name, zerr)
	}
	klog.Infof("zfs: loaded encryption key for %s", root)
	return nil
}

// checkManagedKeySpec rejects a volume that references a key Secret without
// enabling encryption, which the controller never creates (e.g. a hand-edited CR).
func checkManagedKeySpec(vol *apis.ZFSVolume) error {
	if UsesManagedKey(vol) && !EncryptionEnabled(vol.Spec.Encryption) {
		ref := vol.Spec.EncryptionKeyRef
		return fmt.Errorf("zfs: volume %s references the encryption key Secret %s/%s but its encryption is %q",
			vol.Name, ref.Namespace, ref.Name, vol.Spec.Encryption)
	}
	return nil
}

// EnsureKeyLoaded loads vol's key before it is mounted: the keys of
// keylocation=prompt datasets are not loaded on pool import, e.g. after a reboot.
func EnsureKeyLoaded(ctx context.Context, vol *apis.ZFSVolume) error {
	if !UsesManagedKey(vol) {
		return nil
	}
	if err := loadKeyForDataset(ctx, vol, vol.Spec.PoolName+"/"+vol.Name); err != nil {
		return err
	}
	// udev creates the zvol device asynchronously after load-key
	if vol.Spec.VolumeType != VolTypeDataset {
		return waitForZvolDevice(ctx, vol)
	}
	return nil
}

const (
	// zvolDeviceWaitDefaultTimeout bounds the wait for an encrypted zvol's device.
	zvolDeviceWaitDefaultTimeout = 10 * time.Second
	// zvolDeviceWaitInterval is the poll interval while waiting for the device.
	zvolDeviceWaitInterval = 100 * time.Millisecond
	// ZvolDeviceWaitTimeoutEnv overrides the wait timeout with a Go duration.
	ZvolDeviceWaitTimeoutEnv = "ZFS_ZVOL_DEVICE_WAIT_TIMEOUT"
)

// zvolDeviceWaitTimeout returns the wait timeout, set by ZFS_ZVOL_DEVICE_WAIT_TIMEOUT.
func zvolDeviceWaitTimeout() time.Duration {
	if v := os.Getenv(ZvolDeviceWaitTimeoutEnv); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		klog.Warningf("zfs: ignoring invalid %s=%q, using default %s", ZvolDeviceWaitTimeoutEnv, v, zvolDeviceWaitDefaultTimeout)
	}
	return zvolDeviceWaitDefaultTimeout
}

// waitForZvolDevice waits for the device node of an encrypted zvol.
func waitForZvolDevice(ctx context.Context, vol *apis.ZFSVolume) error {
	return waitForDevice(ctx, ZFSDevPath+vol.Spec.PoolName+"/"+vol.Name, zvolDeviceWaitTimeout())
}

// waitForDevice polls for device until it exists, the timeout elapses or ctx is done.
func waitForDevice(ctx context.Context, device string, timeout time.Duration) error {
	if _, err := os.Stat(device); err == nil {
		return nil
	}
	ticker := time.NewTicker(zvolDeviceWaitInterval)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("zfs: interrupted waiting for encrypted zvol device %s: %w", device, ctx.Err())
		case <-timer.C:
			return fmt.Errorf("zfs: encrypted zvol device %s did not appear within %s after load-key", device, timeout)
		case <-ticker.C:
			if _, err := os.Stat(device); err == nil {
				klog.Infof("zfs: encrypted zvol device %s ready", device)
				return nil
			}
		}
	}
}
