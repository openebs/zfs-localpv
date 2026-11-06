# Volume Encryption

zfs-localpv supports [ZFS native encryption](https://openzfs.github.io/openzfs-docs/man/master/8/zfs-load-key.8.html)
per PersistentVolume. Two families are available:

- **StorageClass-level key (legacy)** — one static key for all volumes of a
  StorageClass, via a `keylocation` file present on every node. See the
  `encryption`, `keyformat`, `keylocation` parameters in
  [storageclasses.md](./storageclasses.md).
- **Per-volume key** — each PV gets its own key, sourced from a Kubernetes
  Secret referenced by the PVC.

With per-volume keys, the key material (a 32-byte key encoded as 64 hex
characters) is **never stored in the ZFSVolume CR or on the node disk**, and is
fed to `zfs create` / `zfs load-key` over stdin. The key is kept in a Kubernetes
Secret, typically in **etcd** (base64, plaintext unless etcd encryption-at-rest is
enabled). See *Security notes*. The dataset is created with `keylocation=prompt`,
so the node reloads the key at mount time (via `zfs load-key`), which lets
encrypted volumes survive node reboots.

Volumes encrypted by the driver carry `openebs.io/encrypted` in the PV's
`spec.csi.volumeAttributes` (never the key), so you can tell how a PV is
encrypted without inspecting the `ZFSVolume` CR:
`kubectl get pv <name> -o jsonpath='{.spec.csi.volumeAttributes.openebs\.io/encrypted}'`.

| Value | Meaning |
|---|---|
| `volume` | per-volume key (Secret referenced by the PVC) |
| `storageclass` | StorageClass-level key (`keylocation`), shared by every volume of the class |
| *(absent)* | not encrypted by the driver; the volume may still be encrypted by its pool or parent dataset |

The authoritative source remains the `ZFSVolume` CR (`spec.encryption`,
`spec.encryptionKeyRef`).

## Prerequisites

- A ZFS pool that supports encryption (OpenZFS >= 0.8).
- The external-provisioner runs with `--extra-create-metadata=true` (default in
  the shipped manifests) so the controller learns the PVC name/namespace.
- RBAC (already included in the Helm chart / operator YAML): the controller and
  node components can `get` Secrets cluster-wide (the referenced Secret can live
  in any PVC namespace).

---

## Key from a Kubernetes Secret (per PVC)

The user brings their own key. The PVC references a Secret in its own namespace
via an annotation.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-zfspv-encrypted
parameters:
  poolname: "zfspv-pool"
  fstype: "ext4"
  encryption: "aes-256-gcm"
provisioner: zfs.csi.openebs.io
```

Create the key Secret (64 hex chars = 32 bytes) in the PVC namespace:

```bash
kubectl create secret generic csi-encryption-key \
  --from-literal=key="$(openssl rand -hex 32)"
```

Reference it from the PVC:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-zfspv-encrypted
  annotations:
    encryption.openebs.io/secret-name: "csi-encryption-key"
spec:
  storageClassName: openebs-zfspv-encrypted
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: 4Gi
```

Notes:
- The Secret must be in the **same namespace** as the PVC.
- The Secret data key is `key`; the value must be 64 hex characters.
- Deleting the PV does **not** delete your Secret.
- A CSI clone, or a restore from a `VolumeSnapshot`, always uses its source's
  key: if its PVC carries the annotation naming another Secret, provisioning
  fails with `InvalidArgument`. Velero is not supported, see *Backup and
  restore (Velero): not supported*.

If the annotation is empty, or references a Secret that is missing or does not
contain a valid 64-char hex key, provisioning fails immediately with a clear
`InvalidArgument` error rather than deferring the failure to the node.
A PVC that cannot be read also fails the provisioning (it is retried), since it
may name its own key.
Likewise, if a key stops matching its dataset (an altered Secret, or a wrong
key), `zfs load-key` is rejected and the mount fails with an error naming the
Secret and giving the reason from ZFS (`Incorrect key provided`), rather than
silently mounting nothing.

Full example: [`deploy/sample/encrypted-pvc.yaml`](../deploy/sample/encrypted-pvc.yaml).

---

## Keeping keys in an external KMS (External Secrets Operator)

The driver only reads keys from Kubernetes Secrets. To keep the source of truth,
or a copy, of the keys in an external KMS (HashiCorp Vault, OpenBao, a cloud KMS,
…), use the [External Secrets Operator](https://external-secrets.io/) (ESO).

**Pull — the key lives in the KMS (Secret mode).** An `ExternalSecret`
materialises the key as the Secret referenced by the PVC annotation. It must
exist before the PVC is created, and its value must never change while the
volume exists, so sync it once:

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: db-0-key
  namespace: apps
spec:
  refreshPolicy: CreatedOnce
  secretStoreRef: { kind: ClusterSecretStore, name: vault }
  target:
    name: db-0-key            # referenced by encryption.openebs.io/secret-name
    creationPolicy: Owner
  data:
    - secretKey: key
      remoteRef: { key: zfs/apps/db-0, property: key }
```

With `creationPolicy: Owner`, deleting the `ExternalSecret` deletes the Secret;
the key itself stays in the KMS.

To keep the key material out of etcd, enable
[etcd encryption at rest with a KMS provider](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/).

---

## Backup and restore (Velero): not supported

Backup and restore of volumes that use a per-volume key are **not supported**. In
zfs-localpv, backup and restore are driven by the `ZFSBackup` / `ZFSRestore` CRs,
which the OpenEBS Velero plugin (`openebs/velero-plugin`) creates. Tested with
Velero 1.14 and velero-plugin 3.6.0:

- **Restore fails.** The plugin rebuilds the restored `ZFSVolume` from the saved
  object with its own copy of the zfs-localpv API types, which predates the
  per-volume key fields. The key source is dropped, so the driver has no key to
  hand to `zfs recv`, and the receive fails with `Cannot use 'prompt' keylocation
  because stdin is in use`.
- **Backups are not encrypted.** The backup path streams a non-raw `zfs send`,
  which ZFS decrypts on the fly: the data leaves the node and reaches the object
  store in clear text, and deleting a volume's key does not make its backups
  unreadable. This applies to the legacy StorageClass encryption too.

Supporting it needs changes outside this driver (an updated Velero plugin) and a
design for encrypted backups: a raw `zfs send -w` keeps the stream encrypted,
but restoring it requires the volume's original key. Until then, back up such
volumes at the application level.

---

## Key rotation (manual)

Automated rotation is intentionally **not** performed by the driver: there is no
standard CSI trigger for it, and an interrupted rotation can make data
unrecoverable. Rotate manually on the owning node, keeping both keys until the
volumes mount with the new one.

A Secret can be shared by several volumes, and a clone uses the encryption root
of its source (`zfs get encryptionroot`). Re-wrap **every** encryption root that
uses the Secret before updating it, or the others cannot be unlocked anymore:

```bash
# 1. Save the current key and a new one somewhere durable, outside the cluster.
OLDKEY=$(kubectl -n <ns> get secret <name> -o jsonpath='{.data.key}' | base64 -d)
NEWKEY=$(openssl rand -hex 32)

# 2. For each encryption root using the Secret: load the current key, unless it
#    is already loaded, then re-wrap it with the new key.
printf '%s' "$OLDKEY" | zfs load-key <pool>/<pv-name>
printf '%s' "$NEWKEY" | zfs change-key -o keyformat=hex -o keylocation=prompt <pool>/<pv-name>

# 3. Only once every root is re-wrapped successfully, update the Secret.
kubectl -n <ns> patch secret <name> -p '{"stringData":{"key":"'"$NEWKEY"'"}}'
```

`zfs change-key` re-wraps the dataset master key; existing data is not
re-encrypted. If the rotation is interrupted, or a `change-key` fails, leave the
Secret as it is: check which saved key each root accepts
(`printf '%s' "$KEY" | zfs load-key -n <pool>/<pv-name>`), re-wrap the roots
still on the old key, and only then put the new key in the Secret.

---

## Upgrade and downgrade

- **Upgrading** adds the optional `encryptionKeyRef` field to the `ZFSVolume`,
  `ZFSSnapshot` and `ZFSRestore` CRDs: apply the new CRDs together with the new
  driver (`helm upgrade` does it, and the operator YAML includes them). Existing
  volumes, including those encrypted with a StorageClass `keylocation`, are not
  changed.
- **Downgrading** to a release without per-volume keys is **not supported** once
  a volume uses one: the older CRD schema prunes the unknown `encryptionKeyRef`
  field from the `ZFSVolume` objects, so the driver no longer knows where the key
  is and those volumes cannot be unlocked at mount. Move their data to other
  volumes before downgrading.

---

## Security notes

- Key material is never written to the ZFSVolume CR or the node disk, and is fed
  to `zfs` over stdin. **The key is stored in a Kubernetes Secret, typically in etcd**
  (base64) — enable
  [etcd encryption at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)
  when key custody matters.
- gRPC request/response logging strips secrets, and the key is passed to `zfs`
  on stdin, so it never appears in process arguments or logs.
- The node is granted cluster-wide `get` (not `list`) on Secrets, because a
  referenced Secret can live in any PVC namespace. The controller already has
  `get`/`list` on Secrets (the CSI sidecars need it) and only reads them.

## Limitations

- The managed per-volume key is always `hex` (a 32-byte key as 64 hex chars).
  The legacy StorageClass `keyformat` (`passphrase`/`raw`/`hex`) is unaffected.
- Key rotation is manual (see above) — the driver performs no automatic rotation
  for either the per-volume or the legacy StorageClass encryption.
- **Backup and restore through Velero are not supported** — see *Backup and
  restore (Velero): not supported*.

The legacy `keylocation`-file StorageClass encryption continues to work unchanged
alongside per-volume keys.
