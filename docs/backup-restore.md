# Backup and Restore for LocalPV-ZFS Volumes

## Prerequisites

We should have installed the LocalPV-ZFS 1.0.0 or later version for the Backup and Restore, see [readme](../README.md) for the steps to install the LocalPV-ZFS driver.

| Project | Minimum Version |
| :--- | :--- |
| LocalPV-ZFS | 1.0.0+ |
| Velero | 1.5+ |
| Velero-Plugin | 3.6.0+ |

### Note

- To work with velero-plugin version 2.7.0 (adding support for restore on encrypted zpools) and above we have to update zfs-localpv driver version to at least 1.5.0
- With velero version v1.5.2 and v1.5.3 there is an issue [see here](https://github.com/vmware-tanzu/velero/issues/3470) where PV's are not getting cleaned up for restored volume.

## Setup

### a. Install Velero Binary

follow the steps mentioned [here](https://velero.io/docs/v1.5/basic-install/) to install velero CLI

### b. Install Velero

setup the credential file

```
$ cat /home/pawan/velero/credentials-minio
[default]

aws_access_key_id = minio

aws_secret_access_key = minio123

```
We can install Velero by using below command

```
velero install --provider aws --bucket velero --secret-file /home/pawan/velero/credentials-minio --plugins velero/velero-plugin-for-aws:v1.10.1 --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio.velero.svc:9000 --use-volume-snapshots=true --use-node-agent
```

If you would like to use cloud storage like AWS-S3 buckets for storing backups, you could use a command like the following: 

```
velero install --provider aws --bucket <bucket_name> --secret-file <./aws-iam-creds> --plugins velero/velero-plugin-for-aws:v1.10.1 --backup-location-config region=<bucket_region>,s3ForcePathStyle="true" --use-volume-snapshots=true --use-node-agent
```

We have to install the velero 1.5 or later version for LocalPV-ZFS.

### c. Deploy MinIO

Deploy the minio for storing the backup :-

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/zfs-localpv/develop/deploy/sample/minio.yaml
```

The above minio uses tmp directory inside the pod to store the data, so when restart happens, the backed up data will be gone. We can change the above yaml to use persistence storage to store the data so that we can persist the data after restart.

Check the Velero Pods are UP and Running

```
$ kubectl get po -n velero
NAME                      READY   STATUS      RESTARTS   AGE
minio-d787f4bf7-xqmq5     1/1     Running     0          8s
minio-setup-prln8         0/1     Completed   0          8s
node-agent-lltf2          1/1     Running     0          69s
velero-7d9c448bc5-j424s   1/1     Running     3          69s
```

### d. Setup LocalPV-ZFS Plugin

We can Install the Velero Plugin for LocalPV-ZFS using below command

```
velero plugin add openebs/velero-plugin:3.6.0
```

We have to install the velero-plugin 3.6.0 or later version which has the support for LocalPV-ZFS. Once setup is done, we can go ahead and create the backup/restore.

## Create Backup

We can create 3 kind of backups for LocalPV-ZFS. Let us go through them one by one:

### 1. Create the *Full* Backup

To take the full backup, we can create the Volume Snapshot Location as below :

```yaml
apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: zfspv-full
  namespace: velero
spec:
  provider: openebs.io/zfspv-blockstore
  config:
    bucket: velero
    prefix: zfs
    namespace: openebs # this is the namespace where LocalPV-ZFS creates all the CRs, passed as OPENEBS_NAMESPACE env in the LocalPV-ZFS deployment
    provider: aws
    region: minio
    s3ForcePathStyle: "true"
    s3Url: http://minio.velero.svc:9000
```

The volume snapshot location has the information about where the snapshot should be stored. Here we have to provide the namespace which we have used as OPENEBS_NAMESPACE env while deploying the LocalPV-ZFS. The LocalPV-ZFS Operator yamls uses "openebs" as default value for OPENEBS_NAMESPACE env. Verify the volumesnapshot location:

```
kubectl get volumesnapshotlocations.velero.io -n velero
```

Now, we can execute velero backup command using the above VolumeSnapshotLocation and the LocalPV-ZFS plugin will take the full backup. We can use the below velero command to create the full backup, we can add all the namespaces we want to be backed up in a comma separated format in --include-namespaces parameter.

```
velero backup create my-backup --snapshot-volumes --include-namespaces=<backup-namespace> --volume-snapshot-locations=zfspv-full --storage-location=default
```

We can check the backup status using `velero backup get` command:

```
$ velero backup get
NAME        STATUS       CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
my-backup   InProgress   2020-09-14 21:09:06 +0530 IST   29d       default            <none>
```

Once Status is `Completed`, the backup has been taken successfully.

### 2. Create the scheduled *Full* Backup

To create the scheduled full backup, we can create the Volume Snapshot Location same as above to create the full backup:

```yaml
apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: zfspv-full
  namespace: velero
spec:
  provider: openebs.io/zfspv-blockstore
  config:
    bucket: velero
    prefix: zfs
    namespace: openebs # this is the namespace where LocalPV-ZFS creates all the CRs, passed as OPENEBS_NAMESPACE env in the LocalPV-ZFS deployment
    provider: aws
    region: minio
    s3ForcePathStyle: "true"
    s3Url: http://minio.velero.svc:9000
```

Update the above VolumeSnapshotLocation with namespace and other fields accordingly. Verify the volumesnapshot location:

```
kubectl get volumesnapshotlocations.velero.io -n velero
```

Now, we can create a backup schedule using the above VolumeSnapshotLocation and the LocalPV-ZFS plugin will take the full backup of the resources periodically. For example, to take the full backup at every 5 min, we can create the below schedule :

```
velero create schedule schd --schedule="*/5 * * * *" --snapshot-volumes --include-namespaces=<backup-namespace1>,<backup-namespace2> --volume-snapshot-locations=zfspv-full --storage-location=default
```

The velero will start taking the backup at every 5 minute of the namespaces mentioned in --include-namespaces.

We can check the backup status using `velero backup get` command:

```
$ velero backup get
NAME                   STATUS       CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
schd-20201012122706    InProgress   2020-10-12 17:57:06 +0530 IST   29d       default            <none>

```

The scheduled backup will have `<schedule name>-<timestamp>` format. Once Status is `Completed`, the backup has been taken successfully and then velero will take the next backup after 5 min and periodically keep doing that.

### 3. Create the scheduled *Incremental* Backup

Incremental backup works for scheduled backup only. We can create the VolumeSnapshotLocation as below to create the incremental backup schedule :-

```yaml
apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: zfspv-incr
  namespace: velero
spec:
  provider: openebs.io/zfspv-blockstore
  config:
    bucket: velero
    prefix: zfs
    incrBackupCount: "3" # number of incremental backups we want to have
    namespace: openebs # this is the namespace where LocalPV-ZFS creates all the CRs, passed as OPENEBS_NAMESPACE env in the LocalPV-ZFS deployment
    provider: aws
    region: minio
    s3ForcePathStyle: "true"
    s3Url: http://minio.velero.svc:9000
```

Update the above VolumeSnapshotLocation with namespace and other fields accordingly. Verify the volumesnapshot location:

```
kubectl get volumesnapshotlocations.velero.io -n velero
```

If we have created a backup schedule using the above VolumeSnapshotLocation, the LocalPV-ZFS plugin will start taking the incremental backups. Here, we have to provide `incrBackupCount` parameter which indicates that how many incremental backups we should keep before taking the next full backup. So, in the above case the LocalPV-ZFS plugin will create full backup first and then it will create three incremental backups and after that it will again create a full backup followed by three incremental backups and so on.

For Restore, we need to have the full backup and all the in between the incremental backups available. All the incremental backups are linked to its previous backup, so this link should not be broken otherwise restore will fail.

One thing to note here is `incrBackupCount` parameter defines how many incremental backups we want, it does not include the first full backup. While doing the restore, we just need to give the backup name which we want to restore. The plugin is capable of identifying the incremental backup group and will restore from the full backup and keep restoring the incremental backup till the backup name provided in the restore command.

Now we can create a backup schedule using the above VolumeSnapshotLocation and the LocalPV-ZFS plugin will take care of taking the backup of the resources periodically. For example, to take the incremental backup at every 5 min, we can create the below schedule :

```
velero create schedule schd --schedule="*/5 * * * *" --snapshot-volumes --include-namespaces=<backup-namespace1>,<backup-namespace2> --volume-snapshot-locations=zfspv-incr --storage-location=default --ttl 60m
```

Velero natively does not support the incremental backup, so while taking the incremental backup we have to set the appropriate ttl for the backups so that we have full incremental backup group available for restore. For example, in the above case we creating a schedule to take the backup at every 5 min and VolumeSnapshotLocation says we should keep 3 incremental backups then ttl should be set to 5 min * (3 incr + 1 full) = 20 min or more. So that the full backup and all the incremental backups are available for the restore. If we don't set the ttl correctly and full backup gets deleted, we won't be able use that backup, so we should make sure that correct ttl is set for the incremental backups schedule.

We can check the backup status using `velero backup get` command:

```
$ velero backup get
NAME                  STATUS      CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
schd-20201012134510   Completed   2020-10-12 19:15:10 +0530 IST   29d       default            <none>
schd-20201012134010   Completed   2020-10-12 19:10:10 +0530 IST   29d       default            <none>
schd-20201012133510   Completed   2020-10-12 19:05:10 +0530 IST   29d       default            <none>
schd-20201012133010   Completed   2020-10-12 19:00:10 +0530 IST   29d       default            <none>
schd-20201012132516   Completed   2020-10-12 18:55:18 +0530 IST   29d       default            <none>
schd-20201012132115   Completed   2020-10-12 18:51:15 +0530 IST   29d       default            <none>
```

#### Explanation:

Since we have used incrBackupCount as 3 in the volume snapshot location and created the backup. So first backup will be full backup and next 3 backup will be incremental

```
schd-20201012134510 <============== incr backup, 6th backup
schd-20201012134010 <============== full backup, 5th backup
schd-20201012133510 <============== incr backup, 4th backup
schd-20201012133010 <============== incr backup, 3rd backup
schd-20201012132516 <============== incr backup, 2nd backup
schd-20201012132115 <============== full backup, 1st backup
```

We do not need to know which is the full backup or incremental backup. We can pick any backup in the list and the plugin will find the corresponding full backup and start the restore from there to all the way upto the backup name provided in the restore command. For example, if we want to restore schd-20201012133010, the plugin will restore in the below order

```
1. schd-20201012132115 <============== 1st restore
2. schd-20201012132516 <============== 2nd restore
3. schd-20201012133010 <============== 3rd restore
```
It will stop at 3rd as we want to restore till schd-20201012133010. For us, it will be like we have restored the backup schd-20201012132115 and we don't need to bother about incremenal or full backup.

Suppose we want to restore schd-20201012134010(5th backup), the plugin will restore schd-20201012134010 only as it is full backup and we want to restore till that point only.

## Restore

We can restore the backup using below command, we can provide the namespace mapping if we want to restore in different namespace. If namespace mapping is not provided, then it will restore in the source namespace in which the backup was present.

```
velero restore create --from-backup my-backup --restore-volumes=true --namespace-mappings <source-ns>:<dest-ns>
```
Now we can check the restore status:

```
$ velero restore get
NAME                       BACKUP      STATUS       WARNINGS   ERRORS   CREATED                         SELECTOR
my-backup-20200914211331   my-backup   InProgress   0          0        2020-09-14 21:13:31 +0530 IST   <none>
```

Once the Status is `Completed` we can check the pods in the destination namespace and verify that everything is up and running. We can also verify the data has been restored.

### Restore on a different node

We have the node affinity set on the PV and the ZFSVolume object has the original node name as the owner of the Volume. While doing the restore if original node is not present, the Pod will not come into running state.
We can use velero [RestoreItemAction](https://velero.io/docs/v1.5/restore-reference/#changing-pvc-selected-node) for this and create a config map which will have the node mapping like below:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  # any name can be used; Velero uses the labels (below)
  # to identify it rather than the name
  name: change-pvc-node-selector-config
  # must be in the velero namespace
  namespace: velero
  # the below labels should be used verbatim in your
  # ConfigMap.
  labels:
    # this value-less label identifies the ConfigMap as
    # config for a plugin (i.e. the built-in restore item action plugin)
    velero.io/plugin-config: ""
    # this label identifies the name and kind of plugin
    # that this ConfigMap is for.
    velero.io/change-pvc-node-selector: RestoreItemAction
data:
  # add 1+ key-value pairs here, where the key is the old
  # node name and the value is the new node name.
  pawan-old-node1: pawan-new-node1
  pawan-old-node2: pawan-new-node2
```

While doing the restore the LocalPV-ZFS plugin will set the affinity on the PV as per the node mapping provided in the config map. Here in the above case the PV created on nodes `pawan-old-node1` and `pawan-old-node2` will be moved to `pawan-new-node1` and `pawan-new-node2` respectively.

## Things to Consider:

- Once VolumeSnapshotLocation has been created, we should never modify it, we should always create a new VolumeSnapshotLocation and use that. If we want to modify it, we should cleanup old backups/schedule first and then modify it and then create the backup/schedule. Also we should not switch the volumesnapshot location for the given scheduled backup, we should always create a new schedule if backups for the old schedule is present.

- For the incremental backup, the higher the value of `incrBackupCount` the more time it will take to restore the volumes. So, we should not have very high number of incremental backup.

## UnInstall Velero

We can delete the velero installation by using this command

```
$ kubectl delete namespace/velero clusterrolebinding/velero
$ kubectl delete crds -l component=velero
```

## Reference

Check the [velero doc](https://velero.io/docs/) to find all the supported commands and options for the backup and restore.
