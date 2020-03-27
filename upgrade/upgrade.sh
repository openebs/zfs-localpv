#!/bin/bash

set -e

# ZFSVolumes: create the new CR with apiVersion as zfs.openebs.io and kind as Volume

kubectl get zfsvolumes.openebs.io -n openebs -oyaml > volumes.yaml


# update the group name to zfs.openebs.io
sed -i "s/apiVersion: openebs.io/apiVersion: zfs.openebs.io/g" volumes.yaml
# update the kind to Volume
sed -i "s/kind: ZFSVolume/kind: Volume/g" volumes.yaml
# create the new CR
kubectl apply -f volumes.yaml

rm volumes.yaml


# ZFSSnapshots: create the new CR with apiVersion as zfs.openebs.io and kind as Snapshot

kubectl get zfssnapshots.openebs.io -n openebs -oyaml > snapshots.yaml


# update the group name to zfs.openebs.io
sed -i "s/apiVersion: openebs.io/apiVersion: zfs.openebs.io/g" snapshots.yaml
# update the kind to Snapshot
sed -i "s/kind: ZFSSnapshot/kind: Snapshot/g" snapshots.yaml
# create the new CR
kubectl apply -f snapshots.yaml

rm snapshots.yaml
