# OpenEBS ZFS CSI Driver

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

1. create a Storage class

```
$ cat sc.yaml

```yaml
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
provisioner: openebs.io/zfs
```

The storage class contains the volume paramaters like blocksize, compression, dedup and thinprovision. You can select what are all
parameters you want. The above yaml shows the default values in case paramenters are not provided or wrong value has been provided.
The *poolname* is the must argument. There must be a ZPOOL running on the node with the name given in this storage class.

Here we have to give the provisioner as "openebs.io/zfs" which is the provisioner name of the ZFS driver.

2. create a PVC

```
$ cat pvc.yaml

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

Create a PVC using the storage class created with the openebs.io/zfs provisioner.

3. Check the kubernetes resource is created for the corresponding zfs volume

```
$ kubectl get zv -n openebs
NAME                                       NODE   SIZE
pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9          4294967296
```

Here note that NODE field will be empty as application POD has not yet deployed.
When application will be deployed, as a part of deploying the application the ZFS
driver will create the zfs volume of name pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
in the pool mentioned in the storage class.

4. Deploy the application using this PVC

```
$ cat fio.yaml

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: fio
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - k8s-virtual-machine
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

Here in alpha version of the ZFS driver we have to make use of node selector or node affinity
to make the application pod stick to the node as the application pod should not move to the
other node because the data will be there on one node only.

After the deployment of the application we can go to the node and see that a zfs volume has been
created in the pool mentioned in the storage class and application is using that volume for writting
the data. This is in effect working like waitforFirstConsumer so the actual ZFS volume will be create
when application is deployed to the node.

```
$ zfs list
NAME                                                  USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                           4.25G  92.1G    96K  /zfspv-pool
zfspv-pool/pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9  4.25G  96.4G  5.69M  -
```
Also we can check the kubernetes resource for the corresponding zfs volume

```
$ kubectl get zv -n openebs
NAME                                       NODE                  SIZE
pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9   k8s-virtual-machine   4294967296

$ kubectl describe zv pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9 -n openebs

```yaml
Name:         pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
Namespace:    openebs
Labels:       kubernetes.io/nodename=k8s-virtual-machine
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
  Owner Node ID:  k8s-virtual-machine
  Pool Name:      zfspv-pool
  Thin Provison:  no
Events:           <none>
```

5. ZFS Volume Property Change like compression on/off can be done by just simply
   editing the kubernetes resource for the corresponding zfs volume by using below command :

```
kubectl edit zv pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9 -n openebs
```

You can edit the relevant property like make compression on or make dedup on and save it.
This property will be applied to the corresponding volume and can be verified using
below command on the node:

```
zfs get all zfspv-pool/pvc-37b07ad6-db68-11e9-bbb6-000c296e38d9
```

6. for deprovisioning the volume we can delete the application which is using
   the volume and then we can go ahead and delete the pv, as part of deletion of
   pv this volume will also be deleted from the ZFS pool and data will be freed.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-zfspv" deleted
```

