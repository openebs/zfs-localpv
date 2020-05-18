## Import Existing Volumes to ZFS-LocalPV


### Introduction

There can be cases where node can not be recovered from the failure condition, in that case application can not come into running state as k8s will try to schedule it to the same node as the data is there on that node only. Also, there can be cases where people want to attach the existing volume to the ZFS-LocalPV driver. 

In case of node failure, we can move the disks to the other node and import the pool there. All the volumes and their data will be intact as disks are still in good shape. Now we can just attach those volumes to the ZFS-LocalPV driver and everything will work seamlessly.

Here, I will walk through the steps to attach the existing volumes to the ZFS-LocalPV CSI driver.

### Prerequisites

- We should have ZFS-LocalPV Driver(version 0.6 or later) installed.
- volumes ready to be imported

### Setup

#### Storageclass

Setup a storageclaass which will be used for importing the volumes to ZFS-LocalPV

```
$ cat sc.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  poolname: "zfspv-pool"
  fstype: "zfs"
provisioner: zfs.csi.openebs.io

$ kubectl apply -f sc.yaml
storageclass.storage.k8s.io/openebs-zfspv configured
```

Make sure ZFS-LocalPV components are install and in running state before proceeding further

```
$ kubectl get pods -n kube-system -l role=openebs-zfs
NAME                       READY   STATUS    RESTARTS   AGE
openebs-zfs-controller-0   5/5     Running   0          5m50s
openebs-zfs-node-b42ft     2/2     Running   0          5m50s
openebs-zfs-node-txd59     2/2     Running   0          5m50s
```

### Import Steps


#### Step 1 : Create The Persistence Volume

Get the node name and ZPOOL name where volume is present

```sh
$ zfs list
NAME                 USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool          11.6M   246G    20K  /zfspv-pool
zfspv-pool/fio-vol    24K  4.00G    24K  /zfspv-pool/fio-vol
```

Here in the above ZPOOL, a dataset of name "fio-vol" is present and we want to import that to the ZFS-LocalPV CSI driver. First if volume is mounted then we have to unmount it so that it can be mounted by ZFS-LocalPV driver. For ZFS dataset use `zfs umount` command and for zvol we use `umount` command to unmount the volume

Get the volume size :

For ZFS dataset we can get the volume size as below :-

```
$ zfs get quota zfspv-pool/fio-vol
NAME                PROPERTY  VALUE  SOURCE
zfspv-pool/fio-vol  quota     4G     local
```

For ZVOL we can get the volume size as below :-

```
$ zfs get volsize zfspv-pool/fio-vol
NAME                PROPERTY  VALUE    SOURCE
zfspv-pool/fio-vol  volsize   4G       local
```

Here the volume size is 4Gi(zfs reports Gi as G), we will use this size for PV and PVC.

If volume is a ZFS dataset then create the PV with the fstype as "zfs" :

```
$ cat pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: fio-vol-pv # some unique name
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 4Gi # size of the volume
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: fio-vol-pvc # name of pvc which can claim this PV
    namespace: default # namespace for the pvc
  csi:
    driver: zfs.csi.openebs.io
    fsType: zfs
    volumeAttributes:
      openebs.io/poolname: zfspv-pool # change the pool name accordingly
    volumeHandle: fio-vol # This should be same as the zfs volume name
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - pawan-3 # change the hostname where volume is present
  persistentVolumeReclaimPolicy: Delete
  storageClassName: openebs-zfspv
  volumeMode: Filesystem

$ kubectl apply -f pv.yaml
persistentvolume/fio-vol-pv created
```

The volumeHandle field should be the name of the ZFS volume and Storage field should be same as ZFS volume size. If volume is a ZVOL, then zvol may be formatted with a filesystem, use the same fstype with which zvol was formatted to create the PV. Please note that in claimRef we have mentioned name of the pvc as "fio-vol-pvc", this means that this PV can be claimed by a pvc of name fio-vol-pvc only. We have to make sure pvc claim of the name "fio-vol-pvc" matches with all the details mentioed in the PV like capacity, if it does not match then it will create a new PV for that claim and applicaiton will get a new volume not the desired volume.

Verify the PV has been created with status as "Available"

```
$ kubectl get pv
NAME         CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                 STORAGECLASS    REASON   AGE
fio-vol-pv   4Gi        RWO            Retain           Available   default/fio-vol-pvc   openebs-zfspv            1s
```

#### Step 2 : Attach The Volume With ZFS-LocalPV

If the volume is a zfs dataset, then create ZFS-LocalPV resource with volumeType as "DATASET", otherwise it should be "ZVOL".

```
$ cat zfspvcr.yaml
apiVersion: zfs.openebs.io/v1alpha1
kind: ZFSVolume
metadata:
  finalizers:
  - zfs.openebs.io/finalizer
  name: fio-vol  # should be same as zfs volume name
  namespace: openebs
spec:
  capacity: "4294967296" # size of the volume in bytes
  fsType: zfs
  ownerNodeID: pawan-3 # should be the nodename where ZPOOL is running
  poolName: zfspv-pool # poolname where the volume is present
  volumeType: DATASET # whether it is a DATASET or ZVOL
Status:
  State: Ready
```

Modify the parameters :-
- name should be same as zfs volume name
- capacity to the size of the zfs volumes in bytes
- fstype, the fstype of the volume (zfs, ext4, xfs etc)
- ownerNodeID which is node where the pool is present.
- volumeType should be DATASET if fstype is "zfs" otherwise it should be "ZVOL"

Now volume has been imported to ZFS-LocalPV CSI driver.

### Deploy Application

#### Create PVC

Create the persistence volume claim with the same name as the name given in the PV's claimRef section and in the same namespace

```
$ cat pvc.yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: fio-vol-pvc
spec:
  storageClassName: openebs-zfspv
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi

$ kubectl apply -f pvc.yaml
persistentvolumeclaim/fio-vol-pvc created
```

We can see the this claim has been bound to the PV that we have created

```
$ kubectl get pvc
NAME          STATUS   VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS    AGE
fio-vol-pvc   Bound    fio-vol-pv   4Gi        RWO            openebs-zfspv   2m35s
```

#### Deploy Application

Deploy the application using the persistence volume claim created

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fio
  labels:
    name: fio
spec:
  replicas: 1
  selector:
    matchLabels:
      name: fio
  template:
    metadata:
      labels:
        name: fio
    spec:
      containers:
        - resources:
          name: perfrunner
          image: ljishen/fio
          imagePullPolicy: IfNotPresent
          command: ["/bin/sh"]
          args: ["-c", "while true ;do sleep 50; done"]
          volumeMounts:
            - mountPath: /datadir
              name: fio-vol
      volumes:
        - name: fio-vol
          persistentVolumeClaim:
            claimName: fio-vol-pvc

$ kubectl apply -f app.yaml
deployment.apps/fio created
```
We can see application is running and data is also visible to it :-

```
$ kubectl get po
NAME                   READY   STATUS    RESTARTS   AGE
fio-569965649c-g8dtx   1/1     Running   0          2m53s

$ kubectl exec -it fio-569965649c-g8dtx sh
/ # ls /datadir
a.txt
```
