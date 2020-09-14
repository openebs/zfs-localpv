# Backup and Restore for ZFS-LocalPV Volumes

## Prerequisites

We should have installed the ZFS-LocalPV 1.0.0 or later version for the Backup and Restore, see [readme](../README.md) for the steps to install the ZFS-LocalPV driver.

## Setup

### 1. Install Velero CLI

follow the steps mentioned [here](https://velero.io/docs/v1.4/basic-install/) to install velero CLI

### 2. Deploy Velero

1. setup the credential file

```
$ cat /home/pawan/velero/credentials-minio
[default]

aws_access_key_id = minio

aws_secret_access_key = minio123

```
2. Install Velero

```
velero install --provider aws --bucket velero --secret-file /home/pawan/velero/credentials-minio --plugins velero/velero-plugin-for-aws:v1.0.0-beta.1 --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio.velero.svc:9000 --use-volume-snapshots=true --use-restic
```

### 3. Deploy MinIO

Deploy the minio for storing the backup :-

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/master/deploy/sample/minio.yaml
```

The above minio uses tmp directory inside the pod to store the data, so when restart happens, the backed up data will be gone. We can change the above yaml to use persistence storage to store the data so that we can persist the data after restart.

Check the Velero Pods are UP and Running

```
$ kubectl get po -n velero
NAME                      READY   STATUS      RESTARTS   AGE
minio-d787f4bf7-xqmq5     1/1     Running     0          8s
minio-setup-prln8         0/1     Completed   0          8s
restic-4kx8l              1/1     Running     0          69s
restic-g5zq9              1/1     Running     0          69s
restic-k7k4s              1/1     Running     0          69s
velero-7d9c448bc5-j424s   1/1     Running     3          69s
```

### 4. Setup ZFS-LocalPV Plugin

1. Install the Velero Plugin for ZFS-LocalPV

```
velero plugin add openebs/velero-plugin:2.1.0
```

We have to install the velero-plugin 2.1.0 or later version which has the support for ZFS-LocalPV.

2. Setup the snapshot location to store the data

Create the volume snapshot location which has the information about where the snapshot should be stored

```yaml
apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: zfspv
  namespace: velero
spec:
  provider: openebs.io/zfspv-blockstore
  config:
    bucket: velero
    prefix: zfs
    namespace: openebs # this is namespace where ZFS-LocalPV creates all the CRs, passed as OPENEBS_NAMESPACE env in the ZFS-LocalPV deployment
    provider: aws
    region: minio
    s3ForcePathStyle: "true"
    s3Url: http://minio.velero.svc:9000
```

if you have deployed the ZFS-LocalPV to use different namespace then please use that namespace in the above yaml. To find what namespace ZFS-LocalPV driver is using, we can check what is the value of OPENEBS_NAMESPACE env passed to the ZFS-LocalPV Pods.

Check the volumesnapshot location

```
kubectl get volumesnapshotlocations.velero.io -n velero
```

### 5. Create the Backup

1. Create the backup using the below velero command, add all the namespaces you want to backed up in comma separated format in --include-namespaces parameter.

```
velero backup create my-backup --snapshot-volumes --include-namespaces=<backup-namespaces> --volume-snapshot-locations=zfspv --storage-location=default
```

2. Check the backup status

```
$ velero backup get
NAME        STATUS       CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
my-backup   InProgress   2020-09-14 21:09:06 +0530 IST   29d       default            <none>
```

Once Status is Complete, the backup has been completed successfully.

### 6. Do the Restore

1. We can restore the backup using below command, we can provide the namespace mapping if we want to restore in different namespace. If namespace mapping is not provided, then it will restore in the source namespace in which the backup was present.

```
velero restore create --from-backup my-backup --restore-volumes=true --namespace-mappings <source-ns>:<dest-ns>
```
2. Check the restore status

```
$ velero restore get
NAME                       BACKUP      STATUS       WARNINGS   ERRORS   CREATED                         SELECTOR
my-backup-20200914211331   my-backup   InProgress   0          0        2020-09-14 21:13:31 +0530 IST   <none>
```

Once the Status is Completed we can check the pods in the destination namespace and verify that everything is up and running. We can also verify the data has been restored.

### 6. Uninstall Velero

We can delete the velero installation by using this command

```
$ kubectl delete namespace/velero clusterrolebinding/velero
$ kubectl delete crds -l component=velero
```

## Reference

Check the [velero doc](https://velero.io/docs/) to find all the supported commands and options for the backup and restore.
