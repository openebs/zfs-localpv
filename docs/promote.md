## Promote a clone

A clone made from a snapshot depends on that snapshot, and the snapshot on
its volume. In a chain where each generation is cloned from the last, every
old volume and snapshot has to stay, and the chain only grows. `zfs promote`
reverses one link: the clone takes over the snapshot it was made from (and
every earlier snapshot of its origin), and the origin becomes a clone of it.
After that the old volume and its snapshots can be deleted, and the promoted
volume stands on its own.

The use case that motivated it is a CI build cache. Each run's volume is
cloned from a snapshot of the trunk cache. When a run on the trunk passes,
its volume becomes the new trunk: it is promoted, snapshotted, and the old
trunk is deleted, so the chain stays one generation deep.

### How to ask for it

Create a `ZFSPromote` in the driver's namespace, named after the volume,
with the volume's `ownerNodeID` (from its `ZFSVolume`) and status `Init`:

```yaml
apiVersion: zfs.openebs.io/v1
kind: ZFSPromote
metadata:
  name: pvc-1d2c8e4a-7f0b-4a8e-9c35-2f6f1b0e9d11
  namespace: openebs
spec:
  volumeName: pvc-1d2c8e4a-7f0b-4a8e-9c35-2f6f1b0e9d11
  ownerNodeID: node-1
status: Init
```

The node agent on that node:

1. sets the status to `InProgress`;
2. runs `zfs promote <pool>/<volume>`, unless the volume has no origin
   (it was never a clone, or was already promoted);
3. points every `ZFSSnapshot` whose snapshot now lives under the volume at
   it (the `openebs.io/persistent-volume` label), so restore and delete find
   the snapshot where it is now;
4. sets the status to `Done`.

On an error it sets `Failed` and records a Warning event on the `ZFSPromote`
with the zfs message (`kubectl describe zfspromote -n openebs <name>`).
Every step is safe to repeat, so an agent restart picks up `Init` and
`InProgress` resources from the start.

Wait for `Done` or `Failed`, then delete the `ZFSPromote`; it has no
finalizer and nothing else removes it. A caller needs `create`, `get` and
`delete` on `zfspromotes`, and `get` on `zfsvolumes` to read the
`ownerNodeID`.

If the old origin's PV was already deleted, its `ZFSVolume` was kept only
because it still had snapshots ("marked for deletion"). Once the promote has
moved the last of them, the controller deletes it.

### Things to know

- **Space.** The moved snapshots' space is charged to the promoted volume.
  Under `quotatype: quota` (the default) the quota counts snapshots, so a
  promote fails with "out of space" when the moved snapshots and the
  volume's own data exceed it. Use `quotatype: refquota` for volumes you
  intend to promote.
- **Busy snapshots.** A moved snapshot that is held, for example mounted
  under `.zfs/snapshot`, fails the promote with "dataset is busy".
- **Other clones** of the moved snapshots follow them: they become clones of
  the promoted volume. The old origin cannot be destroyed while clones of
  its remaining snapshots exist, and the promoted volume cannot be destroyed
  while the old origin or other clones depend on its snapshots.
- A clone made from a volume (rather than from a snapshot) is cloned from a
  snapshot that has no `ZFSSnapshot`; the promote moves it but there is no
  record to relabel.
