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

3. Deploy the application using this PVC

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
            - gke-user-zfspv-default-pool-fb71317f-rgcm
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
zfspv-pool/pvc-f52058b7-da1c-11e9-80e0-42010a800fcd  4.25G  96.4G  5.69M  -
```

4. for deprovisioning the volume we can delete the application which is using
   the volume and then we can go ahead and delete the pv, as part of deletion of
   pv this volume will also be deleted from the ZFS pool and data will be free.



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fzfs-localpv?ref=badge_large)