# OpenEBS ZFS CSI Driver
[![Build Status](https://travis-ci.org/openebs/zfs-localpv.svg?branch=master)](https://travis-ci.org/openebs/zfs-localpv)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_shield)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3523/badge)](https://bestpractices.coreinfrastructure.org/en/projects/3523)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://openebsslacksignup.herokuapp.com/)


CSI driver for provisioning Local PVs backed by ZFS and more.

## Project Status

This project is under active development and considered to be in Alpha state.
The current implementation only supports provisioning and de-provisioning of ZFS Volumes. Also, few properties like compression, dedup and recordsize can be provided while provisioning the volumes and can also be changed after provisioning is done.

## Project Tracker

See [roadmap](https://github.com/orgs/openebs/projects/10) and [e2e-test](https://github.com/openebs/e2e-tests/projects/7).

## Usage

### Prerequisites

Before installing ZFS driver please make sure your Kubernetes Cluster
must meet the following prerequisites:

1. all the nodes must have zfs utils installed
2. ZPOOL has been setup for provisioning the volume
3. You have access to install RBAC components into kube-system namespace.
   The OpenEBS ZFS driver components are installed in kube-system namespace
   to allow them to be flagged as system critical components.

### Supported System

K8S : 1.14+

OS : ubuntu 18.04

ZFS : 0.7, 0.8

### Setup

All the node should have zfsutils-linux installed. We should go to the
each node of the cluster and install zfs utils
```
$ apt-get install zfsutils-linux
```

Go to each node and create the ZFS Pool, which will be used for provisioning the volumes. You can create the Pool of your choice, it can be striped, mirrored or raidz pool.

Configure the custom topology keys (if needed). This can be used for many purposes like if we want to create the PV on nodes in a particuler zone or building. We can label the nodes accordingly and use that key in the storageclass for taking the scheduling decesion:

https://github.com/openebs/zfs-localpv/blob/master/docs/faq.md#6-how-to-add-custom-topology-key

### Installation

OpenEBS ZFS driver components can be installed by running the
following command.

```
kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/master/deploy/zfs-operator.yaml
```

Verify that the ZFS driver Components are installed and running using below command :


```
$ kubectl get pods -n kube-system -l role=openebs-zfs
```

Depending on number of nodes, you will see one zfs-controller pod and zfs-node daemonset running
on the nodes.

```
NAME                       READY   STATUS    RESTARTS   AGE
openebs-zfs-controller-0   4/4     Running   0          5h28m
openebs-zfs-node-4d94n     2/2     Running   0          5h28m
openebs-zfs-node-gssh8     2/2     Running   0          5h28m
openebs-zfs-node-twmx8     2/2     Running   0          5h28m

```

Once ZFS driver is installed we can provision a volume.


### Deployment

#### 1. Create a Storage class

```
$ cat sc.yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  recordsize: "4k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

The storage class contains the volume parameters like recordsize, compression, dedup and fstype. You can select what are all
parameters you want. In case, zfs properties paramenters are not provided, the volume will inherit the properties from the ZFS Pool.
Also currently supported fs types are ext2/3/4, xfs and zfs only. The *poolname* is the must argument.
Also there must be a ZPOOL running on *all the nodes* with the name given in the storage class.

##### ext2/3/4 or xfs as FsType

If we provide fstype as ext2/3/4 or xfs, the driver will create a ZVOL, which is a blockdevice carved out of ZFS Pool.
This blockdevice will again formatted as corresponding filesystem(ext2/3/4 or xfs). In this way applications will get desired filesystem.
Here, in this case there will be a filesystem layer on top of ZFS filesystem, and applications may not get the optimal performance.
The sample storage class for ext4 fstype is provided below :-

```
$ cat sc.yaml

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

Here please note that we are providing `volblocksize` instead of `recordsize` since we will create a ZVOL, for which we can choose the blocksize with which we want to create the block device.

##### zfs as FsType

In case if we provide "zfs" as the fstype, the zfs driver will create ZFS DATASET in the ZFS Pool, which is the zfs filesystem.
Here, there will not be any extra layer between application and storage, and applications can get the optimal performance.
The sample storage class for zfs fstype is provided below :-

```
$ cat sc.yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
parameters:
  recordsize: "4k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

Here please note that we are providing `recordsize` which will be used to create the ZFS datasets, which specifies the maximum block size for files in the zfs file system.

##### ZPOOL Availability

If ZFS pool is available on certain nodes only, then make use of topology to tell the list of nodes where we have the ZFS pool available. 
As shown in the below storage class, we can use allowedTopologies to describe ZFS pool availability on nodes.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  recordsize: "4k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: kubernetes.io/hostname
    values:
      - zfspv-node1
      - zfspv-node2
```

The above storage class tells that ZFS pool "zfspv-pool" is available on nodes zfspv-node1 and zfspv-node2 only. The ZFS driver will create volumes on those nodes only.

Please note that the provisioner name for ZFS driver is "zfs.csi.openebs.io", we have to use this while creating the storage class so that the volume provisioning/deprovisioning request can come to ZFS driver.

#### 2. Create a PVC

```
$ cat pvc.yaml

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

Create a PVC using the storage class created for the ZFS driver.

#### 3. Check the kubernetes resource is created for the corresponding zfs volume

```
$ kubectl get zv -n openebs
NAME                                       ZPOOL        NODE          SIZE         VOLBLOCKSIZE   RECORDSIZE   FILESYSTEM
pvc-34133838-0d0d-11ea-96e3-42010a800114   zfspv-pool   zfspv-node1   4294967296                  4k           zfs

```

The ZFS driver will create a ZFS dataset(zvol) on the node zfspv-node1 for the mentioned ZFS pool and the dataset name will same as PV name.
Go to the node zfspv-node1 and check the volume :-

```
$ zfs list
NAME                                                  USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                            444K   362G    96K  /zfspv-pool
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114    96K  4.00G    96K  none

```

#### 4. Scheduler
 
The ZFS driver has a scheduler which will try to distribute the PV across the nodes so that one node should not be loaded with all the volumes. Currently the driver has
VolumeWeighted scheduling algorithm, in which it will try to find a ZFS pool which has less number of volumes provisioned in it from all the nodes where the ZFS pools are available.
Once it is able to find the node, it will create a PV for that node and also create a ZFSVolume custom resource for the volume with the NODE information. The watcher for this ZFSVolume
CR will get all the information for this object and creates a ZFS dataset(zvol) with the given ZFS property on the mentioned node.

The scheduling algorithm currently only accounts for the number of ZFS volumes and does not account for other factors like available cpu or memory while making scheduling decisions.
So if you want to use node selector/affinity rules on the application pod, or have cpu/memory constraints, kubernetes scheduler should be used.
To make use of kubernetes scheduler, you can set the `volumeBindingMode` as `WaitForFirstConsumer` in the storage class.
This will cause a delayed binding, i.e kubernetes scheduler will schedule the application pod first and then it will ask the ZFS driver to create the PV. 
The driver will then create the PV on the node where the pod is scheduled.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  recordsize: "4k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```
Please note that once a PV is created for a node, application using that PV will always get scheduled to that particular node only, as PV will be sticky to that node.
The scheduling algorithm by ZFS driver or kubernetes will come into picture only during the deployment time. Once the PV is created, 
the application can not move anywhere as the data is there on the node where the PV is.

#### 5. Deploy the application using this PVC

```
$ cat fio.yaml

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
    args: ["-c", "while true ;do sleep 50; done"]
    volumeMounts:
       - mountPath: /datadir
         name: fio-vol
    tty: true
  volumes:
  - name: fio-vol
    persistentVolumeClaim:
      claimName: csi-zfspv
```

After the deployment of the application, we can go to the node and see that the zfs volume is being used
by the application for reading/writting the data and space is consumed from the ZFS pool.

Also we can check the kubernetes resource for the corresponding zfs volume

```
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

#### 6. ZFS Property Change
ZFS Volume Property can be changed like compression on/off can be done by just simply editing the kubernetes resource for the corresponding zfs volume by using below command :

```
kubectl edit zv pvc-34133838-0d0d-11ea-96e3-42010a800114 -n openebs
```

You can edit the relevant property like make compression on or make dedup on and save it.
This property will be applied to the corresponding volume and can be verified using
below command on the node:

```
zfs get all zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114
```

#### 7. Snapshot

We can create a snapshot of a volume which can be used further for creating a clone and for taking a backup. To create a snapshot, we have to first create a snapshotclass just like a storage class.

```yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1beta1
metadata:
  name: zfspv-snapclass
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
driver: zfs.csi.openebs.io
deletionPolicy: Delete
```

Then create the snapshot using the above snapshotclass :

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: zfspv-snap
spec:
  volumeSnapshotClassName: zfspv-snapclass
  source:
    persistentVolumeClaimName: csi-zfspv
```
Plese note that, you have to create the snapshot in the same namespace where the pvc is created. Check the created snapshot resource, make sure readyToUse field is true, before using this snapshot for any purpose.

```
$ kubectl get volumesnapshot.snapshot
NAME         AGE
zfspv-snap   2m8s

$ kubectl get volumesnapshot.snapshot zfspv-snap -o yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"snapshot.storage.k8s.io/v1beta1","kind":"VolumeSnapshot","metadata":{"annotations":{},"name":"zfspv-snap","namespace":"default"},"spec":{"source":{"persistentVolumeClaimName":"csi-zfspv"},"volumeSnapshotClassName":"zfspv-snapclass"}}
  creationTimestamp: "2020-02-25T08:25:51Z"
  finalizers:
  - snapshot.storage.kubernetes.io/volumesnapshot-as-source-protection
  - snapshot.storage.kubernetes.io/volumesnapshot-bound-protection
  generation: 1
  name: zfspv-snap
  namespace: default
  resourceVersion: "447494"
  selfLink: /apis/snapshot.storage.k8s.io/v1beta1/namespaces/default/volumesnapshots/zfspv-snap
  uid: 3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
spec:
  source:
    persistentVolumeClaimName: csi-zfspv
  volumeSnapshotClassName: zfspv-snapclass
status:
  boundVolumeSnapshotContentName: snapcontent-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  creationTime: "2020-02-25T08:25:51Z"
  readyToUse: true
  restoreSize: "0"
```

Check the OpenEBS resource for the created snapshot. Check, status should be Ready.

```
$ kubectl get zfssnap -n openebs
NAME                                            AGE
snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd   3m32s

$ kubectl get zfssnap snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd -n openebs -oyaml
apiVersion: openebs.io/v1alpha1
kind: ZFSSnapshot
metadata:
  creationTimestamp: "2020-02-25T08:25:51Z"
  finalizers:
  - zfs.openebs.io/finalizer
  generation: 2
  labels:
    kubernetes.io/nodename: zfspv-node1
    openebs.io/persistent-volume: pvc-34133838-0d0d-11ea-96e3-42010a800114
  name: snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  namespace: openebs
  resourceVersion: "447328"
  selfLink: /apis/openebs.io/v1alpha1/namespaces/openebs/zfssnapshots/snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  uid: 6142492c-3785-498f-aa4a-569ec6c0e2b8
spec:
  capacity: "4294967296"
  fsType: zfs
  ownerNodeID: zfspv-node1
  poolName: zfspv-pool
  volumeType: DATASET
status:
  state: Ready
```

we can go to the node and confirm that snapshot has been created :-

```
$ zfs list -t all
NAME                                                                                                USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                                                                          468K  96.4G    96K  /zfspv-pool
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114                                                  96K  4.00G    96K  none
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd     0B      -    96K  -
```

#### 8. Clone

We can create a clone volume from a snapshot and use that volume for some application. We can create a pvc yaml and mention the snapshot name in the datasource. Please note that for kubernetes version less than 1.17, `VolumeSnapshotDataSource` feature gate needs to be enabled at kubelet and kube-apiserver

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: zfspv-clone
spec:
  storageClassName: openebs-zfspv
  dataSource:
    name: zfspv-snap
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
```
The above yaml says that create a volume from the snapshot zfspv-snap. Applying the above yaml will create a clone volume on the same node where the original volume is present. The newly created clone PV will also be there on the same node where the original PV is there.

Note that the clone PVC should also be of the same size as that of the original volume as right now resize is not supported. Also note that the poolname should also be same, as across the ZPOOL clone is not supported. So, if you are using a separate storageclass for the clone PVC, please make sure it refers to the same ZPOOL.

```
$ kubectl get pvc
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-zfspv     Bound    pvc-34133838-0d0d-11ea-96e3-42010a800114   4Gi        RWO            openebs-zfspv   3h42m
zfspv-clone   Bound    pvc-e1230d2c-b32a-48f7-8b76-ca335b253dcd   4Gi        RWO            openebs-zfspv   78s
```

We can see in the above output that zfspv-clone claim has been created and it is bound also. Also, we can check the zfs list on node and verify that clone volume is created.

```
$ zfs list -t all
NAME                                                                                                USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                                                                          444K  96.4G    96K  /zfspv-pool
zfspv-pool/pvc-e1230d2c-b32a-48f7-8b76-ca335b253dcd                                                   0B     4G    96K  none
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114                                                  96K  4.00G    96K  none
zfspv-pool/pvc-34133838-0d0d-11ea-96e3-42010a800114@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd     0B      -    96K  -
```

The clone volume will have properties same as snapshot properties which are the properties when that snapshot has been created. The ZFSVolume object for the clone volume will be something like below :-

```
$ kubectl describe zv pvc-e1230d2c-b32a-48f7-8b76-ca335b253dcd -n openebs
Name:         pvc-e1230d2c-b32a-48f7-8b76-ca335b253dcd
Namespace:    openebs
Labels:       kubernetes.io/nodename=zfspv-node1
Annotations:  <none>
API Version:  zfs.openebs.io/v1alpha1
Kind:         ZFSVolume
Metadata:
  Creation Timestamp:  2019-11-22T09:49:29Z
  Finalizers:
    zfs.openebs.io/finalizer
  Generation:        1
  Resource Version:  2881
  Self Link:         /apis/openebs.io/v1alpha1/namespaces/openebs/zfsvolumes/pvc-e1230d2c-b32a-48f7-8b76-ca335b253dcd
  UID:               60bc4df2-0d0d-11ea-96e3-42010a800114
Spec:
  Capacity:       4294967296
  Fs Type:        zfs
  Owner Node ID:  zfspv-node1
  Pool Name:      zfspv-pool
  Snap Name:      pvc-34133838-0d0d-11ea-96e3-42010a800114@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  Volume Type:    DATASET
Status:
  State: Ready
Events:           <none>

Here you can note that this resource has Snapname field which tells that this volume is created from that snapshot.

```

#### 9. Volume Resize

check [resize doc](docs/resize.md).

#### 10. Raw Block Volume

check [raw block volume](docs/raw-block-volume.md).

#### 11. Deprovisioning
for deprovisioning the volume we can delete the application which is using the volume and then we can go ahead and delete the pv, as part of deletion of pv this volume will also be deleted from the ZFS pool and data will be freed.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-zfspv" deleted
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_large)
