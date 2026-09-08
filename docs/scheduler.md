# ZFS LocalPV Scheduler

The ZFS driver includes its own scheduler, designed to distribute PVs across nodes to prevent overloading a single node. The driver supports three scheduling algorithms:

1. **VolumeWeighted**: Prioritizes nodes with fewer provisioned volumes.
2. **CapacityWeighted**: Prioritizes nodes with less capacity used from the ZFS pool. This is the default.
3. **SpaceWeighted**: Prioritizes nodes whose ZFS pool has the most free space left.

`VolumeWeighted` and `CapacityWeighted` both order nodes by what has already been put *into* a pool, and neither says anything about what is left in it, so a node with a small and barely used pool outranks a node with a large, moderately used one even though the second has far more room for the volume. `SpaceWeighted` orders by the room left instead. Prefer it when the pools differ widely in size, and keep the default when they are uniform.

For more details on selecting a scheduler via storage class, refer to [this guide](https://github.com/openebs/zfs-localpv/blob/HEAD/docs/storageclasses.md#storageclass-with-k8s-scheduler).

#### Choosing the pool

When the StorageClass names one pool with `poolname`, that is the pool the volume goes into. When it selects a family of pools with [poolpattern](https://github.com/openebs/zfs-localpv/blob/HEAD/docs/storageclasses.md#poolpattern-must-parameter-if-poolname-is-not-set), the node may have several matching pools, and the same algorithm chooses between them: the least used pool under `CapacityWeighted`, the pool holding the fewest volumes under `VolumeWeighted`, and the pool with the most free space under `SpaceWeighted`. The pool that is chosen is recorded on the PersistentVolume and never changes afterwards.

#### Capacity

A volume that reserves space is only placed where the reservation fits: a node qualifies if one of its matching pools has more free space than the volume needs, and the pool chosen on that node is one that fits. If no node qualifies, provisioning fails immediately with `ResourceExhausted` instead of retrying against pools that cannot hold the volume; if no pool matches the StorageClass anywhere, it fails with `FailedPrecondition`.

A volume reserves space when it is a ZVOL (any `fstype` other than `zfs`) that is not thin provisioned, or a dataset created with `thinprovision: "no"`. Thin volumes and datasets that carry only a quota reserve nothing, so they are not restricted by free space and may be provisioned into a pool that is already full.

The free and used figures come from each node agent's periodic report, so the fit is a best effort — `zfs create` on the node remains the final arbiter, and a volume that slips through a stale figure still fails there and is retried by Kubernetes.

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