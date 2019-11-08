# OpenEBS ZFS CSI Driver
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_shield)


CSI driver for provisioning Local PVs backed by ZFS and more.

## Project Status

This project is under active development and considered to be in Alpha state.
The current implementation only supports provisioning and de-provisioning of ZFS Volumes.

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
  blocksize: "4k"
  compression: "off"
  dedup: "off"
  thinprovision: "no"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
```

The storage class contains the volume paramaters like blocksize, compression, dedup and thinprovision. You can select what are all
parameters you want. In case paramenters are not provided, the volume will inherit the properties from the ZFS pool.
The *poolname* is the must argument. Also there must be a ZPOOL running on *all the nodes* with the name given in the storage class.

If ZFS pool is available on certain nodes only, then make use of topology to tell the list of nodes where we have the ZFS pool available. 
As shown in the below storage class, we can use allowedTopologies to describe ZFS pool availability on nodes.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  blocksize: "4k"
  compression: "off"
  dedup: "off"
  thinprovision: "no"
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
NAME                                       ZPOOL        NODE          SIZE
pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9   zfspv-pool   zfspv-node1   4294967296
```

The ZFS driver will create a ZFS dataset(zvol) on the node zfspv-node1 for the mentioned ZFS pool and the dataset name will same as PV name.
Go to the node zfspv-node1 and check the volume :-

```
$ zfs list
NAME                                                  USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                           4.25G  92.1G    96K  /zfspv-pool
zfspv-pool/pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9  4.25G  96.4G  5.69M  -
```

#### 4. Scheduler
 
The ZFS driver has a scheduler which will try to distribute the PV across the nodes so that one node should not be loaded with the volumes. Currently the driver has
VolumeWeighted scheduling algorithm, in which it will try to find a ZFS pool which has less number of volumes provisioned in it from all the nodes where the ZFS pools are available.
Once it is able to find the node, it will create a PV for that node and also create a ZFSVolume custom resource for the volume with the NODE information. The watcher for this ZFSVolume
CR will get all the information for this object and creates a ZFS dataset(zvol) with the given ZFS property on the mentioned node.

As the scheduler takes into account the count of ZFS volumes only for scheduling decisions, it does not account available cpu or memory or anything while scheduling, so if you want to use node selector/affinity rules on the application pod or have cpu, memory constraints, you should use kubernetes scheduler for that, you can put volumeBindingMode as WaitForFirstConsumer in the storage class for delayed binding, which will make kubernetes scheduler to schedule the POD first then it will ask the ZFS driver to create the PV, the driver, then, will create the PV on the node where the POD is scheduled.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  blocksize: "4k"
  compression: "on"
  dedup: "on"
  thinprovision: "no"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```
Please note that once PV is created for a node, application using that PV will always come to that node only as PV will be stick to that node. The scheduling algorithm by ZFS driver or kubernetes will come into picture only at the deployment time, once PV is created, the application can not move anywhere as the data is there on the node where PV is.

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
by the application for reading/writting the data and space is consumed form the ZFS pool.

Also we can check the kubernetes resource for the corresponding zfs volume

```
$ kubectl describe zv pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9 -n openebs

Name:         pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
Namespace:    openebs
Labels:       kubernetes.io/nodename=zfspv-node1
Annotations:  <none>
API Version:  openebs.io/v1alpha1
Kind:         ZFSVolume
Metadata:
  Creation Timestamp:  2019-09-20T05:33:52Z
  Finalizers:
    zfs.openebs.io/finalizer
  Generation:        2
  Resource Version:  20029636
  Self Link:         /apis/openebs.io/v1alpha1/namespaces/openebs/zfsvolumes/pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
  UID:               3b20990a-db68-11e9-bbb6-000c296e38d9
Spec:
  Blocksize:      4k
  Capacity:       4294967296
  Compression:    off
  Dedup:          off
  Encryption:     
  Keyformat:      
  Keylocation: 
  Owner Node ID:  zfspv-node1
  Pool Name:      zfspv-pool
  Thin Provison:  no
Events:           <none>
```

#### 6. ZFS Property Change
ZFS Volume Property can be changed like compression on/off can be done by just simply editing the kubernetes resource for the corresponding zfs volume by using below command :

```
kubectl edit zv pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9 -n openebs
```

You can edit the relevant property like make compression on or make dedup on and save it.
This property will be applied to the corresponding volume and can be verified using
below command on the node:

```
zfs get all zfspv-pool/pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
```

#### 7. Deprovisioning
for deprovisioning the volume we can delete the application which is using the volume and then we can go ahead and delete the pv, as part of deletion of pv this volume will also be deleted from the ZFS pool and data will be freed.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-zfspv" deleted
```



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_large)
