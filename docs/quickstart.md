# Usage and Deployment

## Prerequisites

> **Important:** Before installing the ZFS LocalPV driver, ensure the following prerequisites are met:
- All nodes must have the ZFS utilities package installed.
- A ZFS Pool (ZPool) must be configured for provisioning volumes.

## Setup

All nodes must have the same version of `zfsutils-linux` installed: Please check here for [version](../README.md#supported-system) details. For example for installing on Ubuntu disto use:

```bash
apt-get install zfsutils-linux
```

### Creating a ZFS Pool

On each node, create the ZFS pool that will be used for provisioning volumes. You can create different types of pools (striped, mirrored, RAID-Z) as needed.

For a striped pool on a disk (e.g., `/dev/sdb`):

```bash
zpool create zfspv-pool /dev/sdb
```

For more details on creating mirrored or RAID-Z pools, refer to the [OpenZFS documentation](https://github.com/openzfs/zfs).

#### Creating a Loopback ZFS Pool (For Testing Only)

If no physical disk is available, create a ZFS pool on a loopback device backed by a sparse file:

```bash
truncate -s 100G /tmp/disk.img
zpool create zfspv-pool $(losetup -f /tmp/disk.img --show)
```

### Verifying the ZFS Pool

Run the following command to verify the ZFS pool:

```bash
zpool status
```

Expected output:

```
  pool: zfspv-pool
 state: ONLINE
  scan: none requested
config:

    NAME        STATE     READ WRITE CKSUM
    zfspv-pool  ONLINE       0     0     0
      sdb       ONLINE       0     0     0

errors: No known data errors
```

### Configuring Custom Topology Keys

For advanced scheduling, configure custom topology keys to define volume placement based on zones, racks, or other node-specific attributes. More details are available in the [OpenEBS ZFS FAQ](./faq.md#6-how-to-add-custom-topology-key).

---

## Installation

### Node Labeling

To support future data migration, label each node with a unique `openebs.io/nodeid` value:

```bash
kubectl label node <node-name> openebs.io/nodeid=<unique-id>
```

Refer to the [migration guide](docs/faq.md#8-how-to-migrate-pvs-to-the-new-node-in-case-old-node-is-not-accessible) for more details.

### Installing the OpenEBS ZFS Driver

Installation using operator YAMLs is no longer supported. Instead, use Helm:

```bash
helm repo add openebs https://openebs.github.io/openebs
helm repo update
helm install openebs --namespace openebs openebs/openebs --create-namespace
```

**Note:** If using a custom kubelet directory, specify it during installation:

```bash
--set zfs-localpv.zfsNode.kubeletDir=<your-directory-path>
```

Examples:
- **MicroK8s:** `/var/snap/microk8s/common/var/lib/kubelet/`
- **K0s:** `/var/lib/k0s/kubelet`
- **RancherOS:** `/opt/rke/var/lib/kubelet`

### Verifying Installation

After installation, ensure that the ZFS LocalPV CSI driver components are running:

```bash
kubectl get pods -n openebs -l role=openebs-zfs
```

Expected output (depending on node count):

```
NAME                                              READY   STATUS    RESTARTS   AGE
openebs-zfs-localpv-controller-f78f7467c-blr7q    5/5     Running   0          11m
openebs-zfs-localpv-node-h46m5                    2/2     Running   0          11m
openebs-zfs-localpv-node-svfgq                    2/2     Running   0          11m
openebs-zfs-localpv-node-wm9ks                    2/2     Running   0          11m
```

Once the ZFS driver is installed and running, you can provision volumes.

---

## Deployment

### 1. Creating a StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  recordsize: "128k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```
### Storage Class Configuration

The storage class contains volume parameters such as `recordsize` (which must be a power of 2), `compression`, `dedup`, and `fstype`. You can select the parameters you wish to configure. If ZFS property parameters are not provided, the volume will inherit the properties from the ZFS pool or the defaults.

### Pool Name Requirement

The `poolname` parameter is mandatory. It is important to note that `poolname` can either be the root dataset or a child dataset. For example:

```yaml
poolname: "zfspv-pool"
poolname: "zfspv-pool/child"
```

Additionally, the dataset specified under `poolname` must exist on *all nodes* with the given name in the storage class. Refer to the [Storage Classes documentation](./storageclasses.md) for a complete list of supported parameters for LocalPV-ZFS.

### Supported Filesystem Types

#### ext2/3/4, XFS, or Btrfs as `fstype`

If `fstype` is set to `ext2`, `ext3`, `ext4`, `xfs`, or `btrfs`, the driver will create a ZVOL, which is a block device carved out of the ZFS pool. This block device will be formatted with the specified filesystem before being used by the driver.

> **Note:**
> Since there is a filesystem layer on top of the ZFS volume, applications may not achieve optimal performance.

##### Example: Storage Class for `ext4`

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  volblocksize: "4k"
  compression: "off"
  dedup: "off"
  fstype: "ext4"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

Here, `volblocksize` is specified instead of `recordsize` since a ZVOL is created, and we can define the block size for the block device. Note that for ZFS, `volblocksize` must be a power of 2.

#### ZFS as `fstype`

If `fstype` is set to `zfs`, the ZFS driver will create a ZFS dataset within the ZFS pool, acting as a native ZFS filesystem. In this case, no extra layer exists between the application and storage, allowing for optimal performance.

##### Example: Storage Class for `zfs`

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  recordsize: "128k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

In this case, `recordsize` is specified for ZFS datasets. This defines the maximum block size for files in the ZFS filesystem. The `recordsize` must be a power of 2.

### ZFS Pool Availability

If the ZFS pool is available only on specific nodes, `allowedTopologies` can be used to specify where the pool exists. The following example demonstrates this configuration:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  recordsize: "128k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: kubernetes.io/hostname
    values:
      - k8s-node-1
      - k8s-node2
```

This storage class specifies that the ZFS pool `zfspv-pool` is only available on `k8s-node-1` and `k8s-node-1`, ensuring that volumes are created only on these nodes.

> **Note:** The provisioner name for the ZFS driver is `zfs.csi.openebs.io`. This must be used when creating the storage class to direct volume provisioning and deprovisioning requests to the ZFS driver.

> **Note:** The ZFS driver includes its own scheduler, designed to distribute PVs across nodes to prevent overloading a single node. The driver supports two scheduling algorithms. Check [this](./scheduler.md) to read in detail.

#### 2. Creating a PersistentVolumeClaim (PVC)

To create a PVC using the storage class configured for the ZFS driver, use the following YAML definition:

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-zfspv
spec:
  storageClassName: openebs-zfspv
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
```

The allocated volume size will be rounded off to the nearest MiB or GiB notation. Refer to the [FAQ](./faq.md#7-why-the-zfs-volume-size-is-different-than-the-reqeusted-size-in-pvc) section for more details.

If the storage class uses immediate binding, you can check the corresponding Kubernetes resource for the ZFS volume immediately. However, in the case of late binding, this information will be available only after the pod has been scheduled.

To check the created ZFS volume:

```bash
$ kubectl get zv -n openebs
NAME                                       ZPOOL        NODE           SIZE         STATUS   FILESYSTEM   AGE
pvc-34133838-0d0d-11ea-96e3-42010a800114   zfspv-pool   zfspv-node1    4294967296   Ready    zfs          4s
```

To get detailed information about the ZFS volume:

```bash
$ kubectl describe zv pvc-34133838-0d0d-11ea-96e3-42010a800114 -n openebs
Name:         pvc-34133838-0d0d-11ea-96e3-42010a800114
Namespace:    openebs
Labels:       kubernetes.io/nodename=zfspv-node1
Annotations:  <none>
API Version:  openebs.io/v1alpha1
Kind:         ZFSVolume
Metadata:
  Creation Timestamp:  2019-11-22T09:49:29Z
  Finalizers:
    zfs.openebs.io/finalizer
  Generation:        1
  Resource Version:  2881
  Self Link:         /apis/openebs.io/v1alpha1/namespaces/openebs/zfsvolumes/pvc-34133838-0d0d-11ea-96e3-42010a800114
  UID:               60bc4df2-0d0d-11ea-96e3-42010a800114
Spec:
  Capacity:       4294967296
  Compression:    off
  Dedup:          off
  Fs Type:        zfs
  Owner Node ID:  zfspv-node1
  Pool Name:      zfspv-pool
  Recordsize:     4k
  Volume Type:    DATASET
Status:
  State: Ready
Events:           <none>
```

The ZFS driver will create a ZFS dataset (or a zvol, depending on the `fstype` defined in the storage class) on node `zfspv-node1` within the specified ZFS pool. The dataset name will be the same as the PV name.

To verify the volume on the node `zfspv-node1`, run the following command:

```bash
$ zfs list
NAME                                                  USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                            444K   362G    96K  /zfspv-pool
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114    96K  4.00G    96K  legacy
```

### 3. Deploying an Application

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: fio
spec:
  restartPolicy: Never
  containers:
  - name: perfrunner
    image: openebs/tests-fio
    command: ["/bin/bash"]
    args: ["-c", "while true; do sleep 50; done"]
    volumeMounts:
       - mountPath: /datadir
         name: fio-vol
    tty: true
  volumes:
  - name: fio-vol
    persistentVolumeClaim:
      claimName: csi-zfspv
```

Once the application is deployed, you can verify that the ZFS volume is being utilized by the application for read/write operations. The allocated space will be consumed from the ZFS pool accordingly.

### 4. Modifying ZFS Properties

ZFS volume properties, such as enabling or disabling compression, can be modified by editing the corresponding Kubernetes resource using the following command:

```bash
kubectl edit zv pvc-34133838-0d0d-11ea-96e3-42010a800114 -n openebs
```

Modify the desired properties (e.g., enable compression or deduplication) and save the changes. To verify that the updated properties have been applied to the volume, run the following command on the node:

```bash
zfs get all zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114
```

### 5. Deprovisioning a Volume

To deprovision a volume, first delete the application using the volume. Then, delete the PersistentVolumeClaim (PVC). Once the PVC is deleted, the corresponding volume will be removed from the ZFS pool, freeing up the associated storage.

```bash
kubectl delete -f fio.yaml
# Output: pod "fio" deleted

kubectl delete -f pvc.yaml
# Output: persistentvolumeclaim "csi-zfspv" deleted
```
