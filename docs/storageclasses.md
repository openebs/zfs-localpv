## Parameters

### poolname (*must* parameter)

poolname specifies the name of the pool where the volume has been created. The *poolname* is the must argument. It should be noted that *poolname* can either be the root dataset or a child dataset e.g.
```
poolname: "zfspv-pool"
poolname: "zfspv-pool/child"
```
Also the dataset provided under `poolname` must exist on *all the nodes* with the name given in the storage class.

### fstype (*optional* parameter)

FsType specifies filesystem type for the zfs volume/dataset. If FsType is provided as "zfs", then the driver will create a ZFS dataset, formatting is
not required as underlying filesystem is ZFS anyway. If FsType is ext2, ext3, ext4 btrfs or xfs, then the driver will create a ZVOL and format the volume
accordingly. FsType can not be modified once volume has been provisioned. If fstype is not provided, k8s takes ext4 as the default fstype.

allowed values: "zfs", "ext2", "ext3", "ext4", "xfs", "btrfs"

### recordsize (*optional* parameter)

This parameter is applicable if fstype provided is "zfs" otherwise it will be ignored. It specifies a suggested block size for files in the file system.

allowed values: Any power of 2 from 512 bytes to 128 Kbytes

### volblocksize (*optional* parameter)

This parameter is applicable if fstype is anything but "zfs" where we create a ZVOL a raw block device carved out of ZFS Pool. It specifies the block size to use for the zvol. The volume size can only be set to a multiple of volblocksize, and cannot be zero.

allowed values: Any power of 2 from 512 bytes to 128 Kbytes

### compression (*optional* parameter)

Compression specifies the block-level compression algorithm to be applied to the ZFS Volume and datasets. The value "on" indicates ZFS to use the default compression algorithm.

allowed values: "on", "off", "lzjb", "zstd", "zstd-1", "zstd-2", "zstd-3", "zstd-4", "zstd-5", "zstd-6", "zstd-7", "zstd-8", "zstd-9", "zstd-10", "zstd-11", "zstd-12", "zstd-13", "zstd-14", "zstd-15", "zstd-16", "zstd-17", "zstd-18", "zstd-19", "gzip", "gzip-1", "gzip-2", "gzip-3", "gzip-4", "gzip-5", "gzip-6", "gzip-7", "gzip-8", "gzip-9", "zle", "lz4"

### dedup (*optional* parameter)

Deduplication is the process for removing redundant data at the block level, reducing the total amount of data stored.

allowed values: "on", "off"

### thinprovision (*optional* parameter)

ThinProvision describes whether space reservation for the source volume is required or not. The value "yes" indicates that volume should be thin provisioned and "no" means thick provisioning of the volume. If thinProvision is set to "yes" then volume can be provisioned even if the ZPOOL does not have the enough capacity. If thinProvision is set to "no" then volume can be provisioned only if the ZPOOL has enough capacity and capacity required by volume can be reserved.
Omitting this parameter lets ZFS default behavior prevail: thin provisioning for filesystems and thick provisioning (through `refreservation`) for volumes.

allowed values: "yes", "no"

### shared (*optional* parameter)

Shared specifies whether the volume can be shared among multiple pods. If it is not set to "yes", then the LocalPV-ZFS Driver will not allow the volumes to be mounted by more than one pods. The default value is "no" if shared is not provided in the storageclass.

allowed values: "yes", "no"

## Usage

Let us look at few storageclasses.

### StorageClass Backed by ZFS Dataset

We can create a StorageClass with the fstype as “zfs”. Here, the LocalPV-ZFS driver will create a ZFS dataset for the persistence storage. The application will get a dataset for the storage operation. We can also provide recordsize, compression, or dedup property in the StorageClass. The dataset will be created with all the properties mentioned in the StorageClass:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: openebs-zfspv
allowVolumeExpansion: true
parameters:
 recordsize: "128k"
 thinprovision: "no"
 fstype: "zfs"
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

We have the thinprovision option as “no” in the StorageClass, which means that do reserve the space for all the volumes provisioned using this StorageClass. We can set it to “yes” if we don’t want to reserve the space for the provisioned volumes.

The allowVolumeExpansion is needed if we want to resize the volumes provisioned by the StorageClass. LocalPV-ZFS supports online volume resize, which means we don’t need to scale down the application. The new size will be visible to the application automatically.

Once the storageClass is created, we can go ahead and create the PVC and deploy a pod using that PVC.

### StorageClass Backed by ZFS Volume

There are a few applications that need to have different filesystems to work optimally. For example, Concourse performs best using the “btrfs” filesystem (https://github.com/openebs/zfs-localpv/issues/169). Here we can create a StorageClass with the desired fstype we want. The LocalPV-ZFS driver will create a ZVOL, which is a raw block device carved out from the mentioned ZPOOL and format it to the desired filesystem for the applications to use as persistence storage backed by ZFS Storage Pool:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: opeenbs-zfspv
parameters:
 volblocksize: "4k"
 thinprovision: "yes"
 fstype: "btrfs"
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

Here, we can mention any fstype we want. As of 0.9 release, the driver supports ext2/3/4, xfs, and btrfs fstypes for which it will create a ZFS Volume. Please note here, if fstype is not provided in the StorageClass, the k8s takes “ext4" as the default fstype. Here also we can provide volblocksize, compression, and dedup properties to create the volume, and the driver will create the volume with all the properties provided in the StorageClass.

We have the thinprovision option as “yes” in the StorageClass, which means that it does not reserve the space for all the volumes provisioned using this StorageClass. We can set it to “no” if we want to reserve the space for the provisioned volumes.

### StorageClass for Sharing the Persistence Volumes

By default, the LocalPV-ZFS driver does not allow Volumes to be mounted by more than one pod. Even if we try to do that, only one Pod will come into the running state, and the other Pod will be in ContainerCreating state, and it will be failing on the mount.

If we want to share a volume among multiple pods, we can create a StorageClass with the “shared” option as “yes”. For this, we can create a StorageClass backed by ZFS dataset as below :

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: openebs-zfspv
allowVolumeExpansion: true
parameters:
 fstype: "zfs"
 shared: "yes"
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

Or, we can create the StorageClass backed by ZFS Volume for sharing it among multiple pods as below :

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: openebs-zfspv
allowVolumeExpansion: true
parameters:
 fstype: "ext4"
 shared: "yes"
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

Here, we have to note that all the Pods using that volume will come to the same node as the data is available on that particular node only. Also, applications need to be aware that the volume is shared by multiple pods and should synchronize with the other Pods to access the data from the volume.

### StorageClass With k8s Scheduler

The LocalPV-ZFS Driver has two types of its own scheduling logic, VolumeWeighted and CapacityWeighted (Supported from zfs-driver:1.3.0+). To choose any one of the scheduler add scheduler parameter in storage class and give its value accordingly.
```
parameters:
 scheduler: "VolumeWeighted"
 fstype: "zfs"
 poolname: "zfspv-pool"
```
CapacityWeighted is the default scheduler in zfs-localpv driver, so even if we don't use scheduler parameter in storage-class, driver will pick the node where total provisioned volumes have occupied less capacity from the given pool. On the other hand for using VolumeWeighted scheduler, we have to specify it under scheduler parameter in storage-class. Then driver will pick the node to create volume where ZFS Pool is less loaded with the volumes. Here, it just checks the volume count and creates the volume where less volume is configured in a given ZFS Pool. It does not account for other factors like available CPU or memory while making scheduling decisions.

In case where you want to use node selector/affinity rules on the application pod or have CPU/Memory constraints, the Kubernetes scheduler should be used. To make use of Kubernetes scheduler, we can set the volumeBindingMode as WaitForFirstConsumer in the storage class:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: openebs-zfspv
allowVolumeExpansion: true
parameters:
 fstype: "zfs"
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```

Here, in this case, the Kubernetes scheduler will select a node for the POD and then ask the LocalPV-ZFS driver to create the volume on the selected node. The driver will create the volume where the POD has been scheduled.

From zfs-driver version 1.6.0+, pvc will not be bound till the provisioner succesfully creates the volume on node. Previously, pvc gets bound even if zfs volume creation on nodes keeps failing because scheduler used to return only a single node and provisioner keeps trying to provision the volume on that node only. Now onwards scheduler will return the list of nodes that satisfies the provided topology constraints. Then csi controller will continuosly attempt the volume creation on all these nodes and till volume is created on any of the node or volume creation gets failed on all the nodes. PVC will be bound to a PV only if volume creation succeeds on any one of the nodes.

### StorageClass With Custom Node Labels

There can be a use case where we have certain kinds of ZFS Pool present on certain nodes only, and we want a particular type of application to use that ZFS Pool. We can create a storage class with `allowedTopologies` and mention all the nodes there where that pool is present:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: nvme-zfspv
allowVolumeExpansion: true
parameters:
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
 - key: openebs.io/nodename
   values:
     - node-1
     - node-2
```

At the same time, you must set env variables in the LocalPV-ZFS CSI driver daemon sets (openebs-zfs-node) so that it can pick the node label as the supported topology. It adds "openebs.io/nodename" as default topology key. If the key doesn't exist in the node labels when the CSI ZFS driver register, the key will not add to the topologyKeys. Set more than one keys separated by commas.

```yaml
env:
  - name: OPENEBS_NODE_NAME
    valueFrom:
      fieldRef:
        fieldPath: spec.nodeName
  - name: OPENEBS_CSI_ENDPOINT
    value: unix:///plugin/csi.sock
  - name: OPENEBS_NODE_DRIVER
    value: agent
  - name: OPENEBS_NAMESPACE
    value: openebs
  - name: ALLOWED_TOPOLOGIES
    value: "test1,test2"
```

We can verify that key has been registered successfully with the ZFS LocalPV CSI Driver by checking the CSI node object yaml :-

```yaml
$ kubectl get csinodes pawan-node-1 -oyaml
apiVersion: storage.k8s.io/v1
kind: CSINode
metadata:
  creationTimestamp: "2020-04-13T14:49:59Z"
  name: k8s-node-1
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: k8s-node-1
    uid: fe268f4b-d9a9-490a-a999-8cde20c4dadb
  resourceVersion: "4586341"
  selfLink: /apis/storage.k8s.io/v1/csinodes/k8s-node-1
  uid: 522c2110-9d75-4bca-9879-098eb8b44e5d
spec:
  drivers:
  - name: zfs.csi.openebs.io
    nodeID: k8s-node-1
    topologyKeys:
    - openebs.io/nodename
    - test1
    - test2
```

If you want to change topology keys, just set new env(ALLOWED_TOPOLOGIES) .Check [faq](./faq.md#6-how-to-add-custom-topology-key) for more details.

```
$ kubectl edit ds -n openebs openebs-zfs-node
```

Here we can have ZFS Pool of name “zfspv-pool” created on the nvme disks and want to use this high performing ZFS Pool for the applications that need higher IOPS. We can use the above SorageClass to create the PVC and deploy the application using that.

The LocalPV-ZFS driver will create the Volume in the Pool “zfspv-pool” present on the node  which will be seleted based on scheduler we chose in storage-class. In the above StorageClass, if total capacity of provisioned volumes on node-1 is less, it will create the volume on node-1 only. Alternatively, we can use `volumeBindingMode: WaitForFirstConsumer` to let the k8s select the node where the volume should be provisioned.

The problem with the above StorageClass is that it works fine if the number of nodes is less, but if the number of nodes is huge, it is cumbersome to list all the nodes like this. In that case, what we can do is, we can label all the similar nodes using the same key value and use that label to create the StorageClass.

```
pawan@pawan-master:~/pawan$ kubectl label node pawan-node-2 openebs.io/zpool=nvme
node/pawan-node-2 labeled
pawan@pawan-master:~/pawan$ kubectl label node pawan-node-1 openebs.io/zpool=nvme
node/pawan-node-1 labeled
```

Add "openebs.io/zpool" to the LocalPV-ZFS CSI driver daemon sets env(ALLOWED_TOPOLOGIES). Now, we can create the StorageClass like this:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: nvme-zfspv
allowVolumeExpansion: true
parameters:
 poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
 - key: openebs.io/zpool
   values:
     - nvme
```

Here, the volumes will be provisioned on the nodes which has label “openebs.io/zpool” set as “nvme”.

## Conclusion :

We can set up different kinds of StorageClasses as per our need, and then we can proceed with PVC and POD creation. The driver will take the care of honoring the requests put in the PVC and the StorageClass.
