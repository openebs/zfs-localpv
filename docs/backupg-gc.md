# Backup garbage collector

The backup garbage collector is a component of the OpenEBS ZFS LocalPV CSI driver that is responsible for cleaning up old incremental backups that are orphaned.

It removes the backups that cannot be restored because parent snapshots are missing. This is done to free up space and ensure that the backup system remains efficient and manageable.

Furthermore, when using velero-plugin with the incremental backup feature, the garbage collector will be necessary to keep the backup storage clean and keep the frequency of full backups.

## Configuration

The backup garbage collector can be configured by setting the following environment variable in the `zfs-controller` deployment:

```yaml
env:
  - name: OPENEBS_IO_ENABLE_BACKUP_GC
    value: "true"
```
