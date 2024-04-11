## Prerequisite

For clone, we need to have `VolumeSnapshotDataSource` support, which is in beta in Kubernetes 1.17. If you are using the Kubernetes version less than 1.17, you have to enable the `VolumeSnapshotDataSource` feature gate at kubelet and kube-apiserver.

## Create Clone From Snapshot

We can create a clone volume from a snapshot and use that volume for some application. We can create a PVC YAML and mention the snapshot name in the datasource.

```
$ cat clone.yaml
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

The above yaml says that create a volume from the snapshot zfspv-snap. Applying the above yaml will create a clone volume on the same node where the original volume is present. The newly created clone PV will also be there on the same node where the original PV is there. Apply the clone yaml

```
$ kubectl apply -f clone.yaml 
persistentvolumeclaim/zfspv-clone created
```

Note that the clone PVC should also be of the same size as that of the original volume. Currently resize is not supported. Also, note that the poolname should also be same, as across the ZPOOL clone is not supported. So, if you are using a separate storageclass for the clone PVC, please make sure it refers to the same ZPOOL.

```
$ kubectl get pvc
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-zfspv     Bound    pvc-73402f6e-d054-4ec2-95a4-eb8452724afb   4Gi        RWO            openebs-zfspv   13m
zfspv-clone   Bound    pvc-c095aa52-8d09-4bbe-ac3c-bb88a0e7be19   4Gi        RWO            openebs-zfspv   34s
```

We can see in the above output that zfspv-clone claim has been created and it is bound. Also, we can check the zfs list on node and verify that clone volume is created.

```
$ zfs list -t all
NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT
test-pool                                                                                          834K  9.63G    24K  /test-pool
test-pool/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb                                                  24K  4.00G    24K  /var/lib/kubelet/pods/3862895a-8a67-446e-80f7-f3c18881e391/volumes/kubernetes.io~csi/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb/mount
test-pool/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd     0B      -    24K  -
test-pool/pvc-c095aa52-8d09-4bbe-ac3c-bb88a0e7be19                                                   0B  9.63G    24K  none
```

The clone volume will have properties same as snapshot properties which are the properties when that snapshot has been created. The ZFSVolume object for the clone volume will be something like below:

```
$ kubectl describe zv pvc-c095aa52-8d09-4bbe-ac3c-bb88a0e7be19 -n openebs
Name:         pvc-c095aa52-8d09-4bbe-ac3c-bb88a0e7be19
Namespace:    openebs
Labels:       kubernetes.io/nodename=e2e1-node2
Annotations:  none
API Version:  openebs.io/v1alpha1
Kind:         ZFSVolume
Metadata:
  Creation Timestamp:  2020-02-25T08:34:25Z
  Finalizers:
    zfs.openebs.io/finalizer
  Generation:        1
  Resource Version:  448930
  Self Link:         /apis/openebs.io/v1alpha1/namespaces/openebs/zfsvolumes/pvc-c095aa52-8d09-4bbe-ac3c-bb88a0e7be19
  UID:               e38a9f9a-fb76-466b-a6f9-8d070e0bec6f
Spec:
  Capacity:       4294967296
  Fs Type:        zfs
  Owner Node ID:  e2e1-node2
  Pool Name:      test-pool
  Snapname:       pvc-73402f6e-d054-4ec2-95a4-eb8452724afb@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  Volume Type:    DATASET
Events:           none
```

Here you can note that this resource has Snapname field which tells that this volume is created from that snapshot.

## Create Clone From Volume

We can create a clone volume from an existing volume and use that volume for some application. We can create a PVC YAML and mention the source volume name from where we want to create the clone in the datasource.

```
$ cat clone.yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: zfspv-clone
spec:
  storageClassName: openebs-zfspv
  dataSource:
    name: zfspv-pvc
    kind: PersistentVolumeClaim
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
```

The above yaml says that create a clone from the pvc zfspv-pvc as source. Applying the above yaml will create a clone volume on the same node where the original volume is present. The newly created clone PV will also be there on the same node where the original PV is there. Apply the clone yaml

```
$ kubectl apply -f clone.yaml 
persistentvolumeclaim/zfspv-clone created
```

Note that the clone PVC should also be of the same size as that of the original volume. Also, note that the poolname should also be same, as across the ZPOOL clone is not supported. So, if you are using a separate storageclass for the clone PVC, please make sure it refers to the same ZPOOL.

```
$ kubectl get pvc
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
zfspv-clone   Bound    pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7   4Gi        RWO            openebs-zfspv   52s
zfspv-pvc     Bound    pvc-9df1e7ba-bcb1-414a-b318-5084f4f6edeb   4Gi        RWO            openebs-zfspv   92s
```

We can see in the above output that zfspv-clone claim has been created and it is bound. Also, we can check the zfs list on node and verify that clone volume is created.

```
$ zfs list -t all
NAME                                                                                           USED  AVAIL  REFER  MOUNTPOINT
zfspv-pool                                                                                    4.26G   497G    24K  /zfspv-pool
zfspv-pool/pvc-9df1e7ba-bcb1-414a-b318-5084f4f6edeb                                           4.25G   502G   130M  -
zfspv-pool/pvc-9df1e7ba-bcb1-414a-b318-5084f4f6edeb@pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7     0B      -   130M  -
zfspv-pool/pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7                                             67K   497G   130M  -
```

The clone volume will have properties same as source volume properties at the time of creating the clone. The ZFSVolume object for the clone volume will be something like below:

```
$ kubectl describe zv pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7 -n openebs
Name:         pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7
Namespace:    openebs
Labels:       kubernetes.io/nodename=pawan-node-1
              openebs.io/source-volume=pvc-9df1e7ba-bcb1-414a-b318-5084f4f6edeb
Annotations:  <none>
API Version:  zfs.openebs.io/v1
Kind:         ZFSVolume
Metadata:
  Creation Timestamp:  2020-12-10T05:00:54Z
  Finalizers:
    zfs.openebs.io/finalizer
  Generation:        2
  Resource Version:  53615100
  Self Link:         /apis/zfs.openebs.io/v1/namespaces/openebs/zfsvolumes/pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7
  UID:               b67ea231-0f5c-4d15-918c-425160706953
Spec:
  Capacity:       4294967296
  Owner Node ID:  pawan-node-1
  Pool Name:      zfspv-pool
  Snapname:       pvc-9df1e7ba-bcb1-414a-b318-5084f4f6edeb@pvc-b757fbca-f008-49c6-954e-7ea3e1c1bbc7
  Volume Type:    ZVOL
Status:
  State:  Ready
Events:   <none>
```

The LocalPV-ZFS driver creates an internal snapshot on the source volume with the name same as clone volume name and then creates the clone from that snapshot. Here you can note that this resource has Snapname field which tells that this volume is created from that internal snapshot.
