### 1. What is ZFS-LocalPV

ZFS-LocalPV is a CSI driver for dynamically provisioning a volume in ZFS storage. It also takes care of tearing down the volume from the ZFS storage once volume is deprovisioned.

### 2. How to install ZFS-LocalPV

Make sure that all the nodes have zfsutils-linux installed. We should go to the each node of the cluster and install zfs utils

```
$ apt-get install zfsutils-linux
```
Go to each node and create the ZFS Pool, which will be used for provisioning the volumes. You can create the Pool of your choice, it can be striped, mirrored or raidz pool.

Once ZFS POOL is created we can install OpenEBS ZFS driver by running the following command.

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

### 3. How to upgrade the driver to newer version

Follow the instructions here https://github.com/openebs/zfs-localpv/tree/master/upgrade.

### 4. ZFS Pools are there on certain nodes only, how can I create the storage class.

If ZFS pool is available on certain nodes only, then make use of topology to tell the list of nodes where we have the ZFS pool available.
As shown in the below storage class, we can use allowedTopologies to describe ZFS pool availability on nodes.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
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


### 5. How to install the provisioner in HA

To have HA for the provisioner(controller), we can update the replica count to 2(or more as per need) and deploy the yaml. Once yaml is deployed, you can see 2(or more) controller pod running. At a time only one will be active and once it is down, the other will take over. They will use lease mechanism to decide who is active/master. Please note that it has anti affinity rules, so on one node only one pod will be running, that means, if you are using 2 replicas on a single node cluster, the other pod will be in pending state because of the anti-affinity rule. So, before changing the replica count, please make sure you have sufficient nodes.

here is the yaml snippet to do that :-

```yaml
kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: openebs-zfs-controller
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: openebs-zfs-controller
      role: openebs-zfs
  serviceName: "openebs-zfs"
  replicas: 2
---
```

### 6. How to add custom topology key

To add custom topology key, we can label all the nodes with the required key and value :-

```sh
$ kubectl label node pawan-node-1 openebs.io/rack=rack1
node/pawan-node-1 labeled

$ kubectl get nodes pawan-node-1 --show-labels
NAME           STATUS   ROLES    AGE   VERSION   LABELS
pawan-node-1   Ready    worker   16d   v1.17.4   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64,kubernetes.io/hostname=pawan-node-1,kubernetes.io/os=linux,node-role.kubernetes.io/worker=true,openebs.io/rack=rack1

```
It is recommended is to label all the nodes with the same key, they can have different values for the given keys, but all keys should be present on all the worker node.

Once we have labeled the node, we can install the zfs driver. The driver will pick the node labels and add that as the supported topology key. If the driver is already installed and you want to add a new topology information, you can label the node with the topology information and then restart of the ZFSPV CSI driver daemon sets (openebs-zfs-node) are required so that the driver can pick the labels and add them as supported topology keys. We should restart the pod in kube-system namespace with the name as openebs-zfs-node-[xxxxx] which is the node agent pod for the ZFS-LocalPV Driver.

Note that restart of ZFSPV CSI driver daemon sets are must in case, if we are going to use WaitForFirstConsumer as volumeBindingMode in storage class. In case of immediate volume binding mode, restart of daemon set is not a must requirement, irrespective of sequence of labeling the node either prior to install zfs driver or after install. However it is recommended to restart the daemon set if we are labeling the nodes after the installation.

```sh
$ kubectl get pods -n kube-system -l role=openebs-zfs

NAME                       READY   STATUS    RESTARTS   AGE
openebs-zfs-controller-0   4/4     Running   0          5h28m
openebs-zfs-node-4d94n     2/2     Running   0          5h28m
openebs-zfs-node-gssh8     2/2     Running   0          5h28m
openebs-zfs-node-twmx8     2/2     Running   0          5h28m
```

We can verify that key has been registered successfully with the ZFSPV CSI Driver by checking the CSI node object yaml :-

```yaml
$ kubectl get csinodes pawan-node-1 -oyaml
apiVersion: storage.k8s.io/v1
kind: CSINode
metadata:
  creationTimestamp: "2020-04-13T14:49:59Z"
  name: pawan-node-1
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: pawan-node-1
    uid: fe268f4b-d9a9-490a-a999-8cde20c4dadb
  resourceVersion: "4586341"
  selfLink: /apis/storage.k8s.io/v1/csinodes/pawan-node-1
  uid: 522c2110-9d75-4bca-9879-098eb8b44e5d
spec:
  drivers:
  - name: zfs.csi.openebs.io
    nodeID: pawan-node-1
    topologyKeys:
    - beta.kubernetes.io/arch
    - beta.kubernetes.io/os
    - kubernetes.io/arch
    - kubernetes.io/hostname
    - kubernetes.io/os
    - node-role.kubernetes.io/worker
    - openebs.io/rack
```

We can see that "openebs.io/rack" is listed as topology key. Now we can create a storageclass with the topology key created :

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
allowedTopologies:
- matchLabelExpressions:
  - key: openebs.io/rack
    values:
      - rack1
```

The ZFSPV CSI driver will schedule the PV to the nodes where label "openebs.io/rack" is set to "rack1". If there are multiple nodes qualifying this prerequisite, then it will pick the node which has less number of volumes provisioned for the given ZFS Pool.

Note that if storageclass is using Immediate binding mode and topology key is not mentioned then all the nodes should be labeled using same key, that means, same key should be present on all nodes, nodes can have different values for those keys. If nodes are labeled with different keys i.e. some nodes are having different keys, then ZFSPV's default scheduler can not effictively do the volume count based scheduling. Here, in this case the CSI provisioner will pick keys from any random node and then prepare the preferred topology list using the nodes which has those keys defined and ZFSPV scheduler will schedule the PV among those nodes only.

### 7. Why the ZFS volume size is different than the reqeusted size in PVC

Here, we have to note that the size will be rounded off to the nearest Mi or Gi unit. Please note that M/G notation uses 1000 base and Mi/Gi notation uses 1024 base, so 1M will be 1000 * 1000 byte and 1Mi will be 1024 * 1024.

The driver uses below logic to roundoff the capacity:

1. if PVC size is > Gi (1024 * 1024 * 1024), then it will find the size in the nearest Gi unit and allocate that.

allocated = ((size + 1Gi - 1) / Gi) * Gi

For example if the PVC is requesting 4G storage space :-

```
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
      storage: 4G
```

Then driver will find the nearest size in Gi, the size allocated will be ((4G + 1Gi - 1) / Gi) * Gi, which will be 4Gi.

2. if PVC size is < Gi (1024 * 1024 * 1024), then it will find the size in the nearest Mi unit and allocate that.

allocated = ((size + 1Mi - 1) / Mi) * Mi

For example if the PVC is requesting 1G (1000 * 1000 * 1000) storage space which is less than 1Gi (1024 * 1024 * 1024):-

```
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
      storage: 1G
```

Then driver will find the nearest size in Mi, the size allocated will be ((1G + 1Mi - 1) / Mi) * Mi, which will be 954Mi.

PVC size as zero in not a valid capacity. The minimum allocatable size for the ZFS-LocalPV driver is 1Mi, which means that if we are requesting 1 byte of storage space then 1Mi will be allocated for the volume.

