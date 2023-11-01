# CRD Upgrade

The following CRDs were upgraded in v2.3.0 to support 'zstd' ZFS pool compression variants:
- zfsvolumes.zfs.openebs.io
- zfssnapshots.zfs.openebs.io
- zfsrestores.zfs.openebs.io

Using zstd-3 or similar compression algorithms (zstd with the '-x' suffix) through the StorageClass may lead to a failure when provisioning a volume, if the CRDs are not upgraded.

Error log when trying to use 'zstd-3' without upgrading CRDs:
```
zfs.csi.openebs.io_openebs-zfslocalpv-zfs-localpv-controller-0_2332b051-644b-485c-80e1-be9aed3a5827  failed to provision volume with StorageClass "openebs-zfspv": rpc error: code = Internal desc = not able to provision the volume, nodes [node-0-125210 node-1-125210 node-2-125210], err : ZFSVolume.zfs.openebs.io "pvc-c9f7ad46-9efb-4e2a-87f9-149a0d1cacae" is invalid: spec.compression: Invalid value: "zstd-3": spec.compression in body should match '^(on|off|lzjb|zstd|gzip|gzip-[1-9]|zle|lz4)$'
```

Upgrade the CRDs using the following commands:
```
# Upgrade ZFSVolumes CRD
curl -LO https://raw.githubusercontent.com/openebs/zfs-localpv/zfs-localpv-2.3.1/deploy/yamls/zfsvolume-crd.yaml
kubectl patch crd zfsvolumes.zfs.openebs.io --patch-file zfsvolume-crd.yaml

# Upgrade ZFSSnapshots CRD
curl -LO https://raw.githubusercontent.com/openebs/zfs-localpv/zfs-localpv-2.3.1/deploy/yamls/zfssnapshot-crd.yaml
kubectl patch crd zfssnapshots.zfs.openebs.io --patch-file zfssnapshot-crd.yaml

# Upgrade ZFSRestores CRD
curl -LO https://raw.githubusercontent.com/openebs/zfs-localpv/zfs-localpv-2.3.1/deploy/yamls/zfsrestore-crd.yaml
kubectl patch crd zfsrestores.zfs.openebs.io --patch-file zfsrestore-crd.yaml
```

Delete and recreate the PVC.