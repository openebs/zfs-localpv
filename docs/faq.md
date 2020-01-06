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

In the [operator file](../deploy/zfs-operator.yaml), change the zfs-driver image to the required tag which you want (like for tag v0.2 use `quay.io/openebs/zfs-driver:v0.2`), and then apply the yaml, there are two places where we need to change the image, one for the controller and once for the node agent. By default, the operator uses the `ci` tag which always points to development image not the release tag, so if you want to test the development image you can use ci tag. Please note that the default ImagePullPolicy is IfNotPresent, that means if `ci` image is already there on the node, it will not be pulled again.

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


### 3. How to install the provisioner in HA

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
