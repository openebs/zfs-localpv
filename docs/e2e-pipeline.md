# E2e test cases for ZFS-LocaPV

* Automated test cases into e2e-pipelines
https://gitlab.openebs.ci/openebs/e2e-nativek8s/pipelines/

1. Validation of ZFS-LocalPV provisioner.
2. Provision and Deprovision of ZFS-volume with Percona-mysql application (Both ext4 and zfs file system).
3. Validation of ZFS-LocalPV snapshot.
4. Validation of ZFS-LocalPV clone.

* Manual test cases

1. Check for the parent volume; it should not be deleted when volume snapshot is present.
2. Check for the clone volume; it should contain only that snapshot content from which it is cloned.
3. Test case for the scheduler to verify it is doing volume count based scheduling.
4. Test case for zfs-volume properties change and validate that changes are applied to the corresponding volume. (Only compression and dedup properties as of now)
5. Verify the data-persistence after draining the node.

* Test cases planned for future

1. Validation of volume resize support for zfs-LocalPV.
2. Add manually tested cases into the pipelines.