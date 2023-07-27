## Snapshot

We can create a snapshot of a volume that can be used further for creating a clone and for taking a backup. To create a snapshot, we have to first create a SnapshotClass just like a storage class where you can provide deletionPolicy as Retain or Delete.

```yaml
$ cat snapshotclass.yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1
metadata:
  name: zfspv-snapclass
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
driver: zfs.csi.openebs.io
deletionPolicy: Delete
```

Apply the snapshotclass YAML:

```
$ kubectl apply -f snapshotclass.yaml
volumesnapshotclass.snapshot.storage.k8s.io/zfspv-snapclass created
```

Find a PVC for which snapshot has to be created

```
$ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-zfspv   Bound    pvc-73402f6e-d054-4ec2-95a4-eb8452724afb   4Gi        RWO            openebs-zfspv   2m35s
```

Create the snapshot using the created SnapshotClass for the selected PVC

```
$ cat snapshot.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: zfspv-snap
spec:
  volumeSnapshotClassName: zfspv-snapclass
  source:
    persistentVolumeClaimName: csi-zfspv
```

Apply the snapshot.yaml

```
$ kubectl apply -f snapshot.yaml
volumesnapshot.snapshot.storage.k8s.io/zfspv-snap created
```

Please note that you have to create the snapshot in the same namespace where the PVC is created. Check the created snapshot resource, make sure readyToUsefield is true, before using this snapshot for any purpose.

```
$ kubectl get volumesnapshot.snapshot
NAME         AGE
zfspv-snap   2m8s
```
```
$ kubectl get volumesnapshot.snapshot zfspv-snap -o yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"snapshot.storage.k8s.io/v1","kind":
      "VolumeSnapshot","metadata":{"annotations":{},"name":
      "zfspv-snap","namespace":"default"},"spec":{"source":{"persistentVolumeClaimName":"csi-zfspv"},"volumeSnapshotClassName":"zfspv-snapclass"}}
  creationTimestamp: "2020-02-25T08:25:51Z"
  finalizers:
  - snapshot.storage.kubernetes.io/volumesnapshot-as-source-protection
  - snapshot.storage.kubernetes.io/volumesnapshot-bound-protection
  generation: 1
  name: zfspv-snap
  namespace: default
  resourceVersion: "447494"
  selfLink: /apis/snapshot.storage.k8s.io/v1/namespaces/default/volumesnapshots/zfspv-snap
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
```
```
$ kubectl get zfssnap snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd -n openebs -oyaml
apiVersion: openebs.io/v1alpha1
kind: ZFSSnapshot
metadata:
  creationTimestamp: "2020-02-25T08:25:51Z"
  finalizers:
  - zfs.openebs.io/finalizer
  generation: 2
  labels:
    kubernetes.io/nodename: e2e1-node2
    openebs.io/persistent-volume: pvc-73402f6e-d054-4ec2-95a4-eb8452724afb
  name: snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  namespace: openebs
  resourceVersion: "447328"
  selfLink: /apis/openebs.io/v1alpha1/namespaces/openebs/zfssnapshots/snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd
  uid: 6142492c-3785-498f-aa4a-569ec6c0e2b8
spec:
  capacity: "4294967296"
  fsType: zfs
  ownerNodeID: e2e1-node2
  poolName: test-pool
  volumeType: DATASET
status:
  state: Ready
```

We can go to the node and confirm that snapshot has been created:

```
# zfs list -t all
NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT
test-pool                                                                                          818K  9.63G    24K  /test-pool
test-pool/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb                                                  24K  4.00G    24K  /var/lib/kubelet/pods/3862895a-8a67-446e-80f7-f3c18881e391/volumes/kubernetes.io~csi/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb/mount
test-pool/pvc-73402f6e-d054-4ec2-95a4-eb8452724afb@snapshot-3cbd5e59-4c6f-4bd6-95ba-7f72c9f12fcd     0B      -    24K  -
```
