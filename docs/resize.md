## ZFS-LocalPV Volume Resize

We can resize the volume by updating the PVC yaml to the desired size and apply it. The ZFS Driver will take care of updating the quota in case of dataset. If we are using a Zvol and have mounted it as ext2/3/4 or xfs file system, the driver will take care of expanding the volume via reize2fs/xfs_growfs binaries.

For resize, storageclass that provisions the pvc must support resize. We should have allowVolumeExpansion as true in storageclass

```
$ cat sc.yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io


$ kubectl apply -f sc.yaml
storageclass.storage.k8s.io/openebs-zfspv created
```

Create the PVC using the above storage class

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


$ kubectl apply -f pvc.yaml
persistentvolumeclaim/csi-zfspv created
```

OpenEBS ZFS driver supports Online Volume expansion, which means that we can expand the volume even if volume is being used by the application and we also don't need to restart the application to use the expanded volume, the ZFS Driver will take care of making the space availbale to it. Please note that file system expansion does not happen until a Application pod references the resized volume, so if no pods referencing the volume are running, file system expansion will not happen.

Deploy the application using the PVC. Here is sample yaml for the application :

```
$ cat fio.yaml
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
            claimName: csi-zfspv


$ kubectl apply -f fio.yaml
deployment.apps/fio created

$ kubectl get po
NAME                   READY   STATUS    RESTARTS   AGE
fio-5b7884bc7b-4mssk   1/1     Running   0          40s

```

Check the current PVC status

```
$ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-zfspv   Bound    pvc-966b0749-5dea-442f-a584-013cf5d25201   4Gi        RWO            openebs-zfspv   85s

```
Exec into the application pod and check the size

```
# df -h /datadir/
Filesystem                Size      Used Available Use% Mounted on
/dev/zd0                  3.9G     16.0M      3.8G   0% /datadir
```

Deploy the application using the PVC which supports volume expansion. Once the application pod is deployed, we will expand the PVC to 5Gi from 4Gi. Just edit the PVC yaml and update the size to 5Gi and apply it :-

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
      storage: 5Gi
```

Apply the above yaml which will resize the volume

```
$ kubectl apply -f pvc.yaml
persistentvolumeclaim/csi-zfspv configured

```

Check the PVC yaml

```yaml
$ kubectl get pvc csi-zfspv -oyaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"PersistentVolumeClaim","metadata":{"annotations":{},"name":"csi-zfspv","namespace":"default"},"spec":{"accessModes":["ReadWriteOnce"],"r
esources":{"requests":{"storage":"5Gi"}},"storageClassName":"openebs-zfspv"}}
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: zfs.csi.openebs.io
  creationTimestamp: "2020-03-06T06:40:08Z"
  finalizers:
  - kubernetes.io/pvc-protection
  name: csi-zfspv
  namespace: default
  resourceVersion: "2547405"
  selfLink: /api/v1/namespaces/default/persistentvolumeclaims/csi-zfspv
  uid: 966b0749-5dea-442f-a584-013cf5d25201
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: openebs-zfspv
  volumeMode: Filesystem
  volumeName: pvc-966b0749-5dea-442f-a584-013cf5d25201
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 4Gi
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2020-03-06T06:41:22Z"
    message: Waiting for user to (re-)start a pod to finish file system resize of
      volume on node.
    status: "True"
    type: FileSystemResizePending
  phase: Bound

```

Here you see in the message that it is waiting on FileSystemResizePending. The resize request will go to the node where appliccation pod is running. The ZFS driver node agent will resize the filesytem for the application. Keep checking the PVC yaml for FileSystemResizePending to go away, once PVC is resized, the yaml will look like this :-

```yaml
$ kubectl get pvc csi-zfspv -oyaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"PersistentVolumeClaim","metadata":{"annotations":{},"name":"csi-zfspv","namespace":"default"},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"5Gi"}},"storageClassName":"openebs-zfspv"}}
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: zfs.csi.openebs.io
  creationTimestamp: "2020-03-06T06:40:08Z"
  finalizers:
  - kubernetes.io/pvc-protection
  name: csi-zfspv
  namespace: default
  resourceVersion: "2547449"
  selfLink: /api/v1/namespaces/default/persistentvolumeclaims/csi-zfspv
  uid: 966b0749-5dea-442f-a584-013cf5d25201
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: openebs-zfspv
  volumeMode: Filesystem
  volumeName: pvc-966b0749-5dea-442f-a584-013cf5d25201
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 5Gi
  phase: Bound
```

```
$ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-zfspv   Bound    pvc-675bf643-c744-4a30-984c-5b2c53c51f14   5Gi        RWO            openebs-zfspv   28m
```

Also, we can exec into the application pod and verify the same :-

```
# df -h /datadir/
Filesystem                Size      Used Available Use% Mounted on
/dev/zd0                  4.9G     16.0M      4.8G   0% /datadir
```
As we can see the volume mount point /datadir is showing that it has been resized.
