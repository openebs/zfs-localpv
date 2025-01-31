# ZFS LocalPV Scheduler

The ZFS driver includes its own scheduler, designed to distribute PVs across nodes to prevent overloading a single node. The driver supports two scheduling algorithms:

1. **VolumeWeighted**: Prioritizes nodes with fewer provisioned volumes.
2. **CapacityWeighted**: Prioritizes nodes with less capacity used from the ZFS pool.

For more details on selecting a scheduler via storage class, refer to [this guide](https://github.com/openebs/zfs-localpv/blob/HEAD/docs/storageclasses.md#storageclass-with-k8s-scheduler).

Once the driver selects a node, it creates a Persistent Volume (PV) and a `ZFSVolume` custom resource containing node information. A watcher process monitors this `ZFSVolume` resource and provisions a ZFS dataset (ZVOL) with the specified properties on the chosen node.

Currently, the scheduler does not consider factors such as CPU or memory availability, focusing solely on ZFS volumes and pool capacity. If you require CPU/memory constraints or node affinity rules, Kubernetes' native scheduler should be used.

#### Using Kubernetes Scheduler

To leverage Kubernetes' scheduler, set `volumeBindingMode` to `WaitForFirstConsumer`. This delays volume binding until the application pod is scheduled, ensuring that PV creation aligns with pod placement.

##### Example: Storage Class with Kubernetes Scheduler

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv
allowVolumeExpansion: true
parameters:
  recordsize: "128k"
  compression: "off"
  dedup: "off"
  fstype: "zfs"
  poolname: "zfspv-pool"
provisioner: zfs.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```

> **Note:** Once a PV is created for a specific node, any application using that PV will always be scheduled on the same node. PVs remain bound to their respective nodes.

### Conclusion

The scheduling mechanism (ZFS driver or Kubernetes) is only relevant during deployment. Once a PV is created, the application cannot move to another node since the data resides on the node where the PV was initially provisioned.