#!/bin/bash

set -e

kubectl get zfsvolumes.openebs.io -n openebs -oyaml > volumes.yaml

# remove the finalizer from the old CR
sed -i "/zfs.openebs.io\/finalizer/d" volumes.yaml
kubectl apply -f volumes.yaml

# delete the old CR
kubectl delete -f volumes.yaml

# delete the CRD definition
kubectl delete crd zfsvolumes.openebs.io


kubectl get zfssnapshots.openebs.io -n openebs -oyaml > snapshots.yaml

# remove the finalizer from the old CR
sed -i "/zfs.openebs.io\/finalizer/d" snapshots.yaml
kubectl apply -f snapshots.yaml

# delete the old CR
kubectl delete -f snapshots.yaml

# delete the CRD definition
kubectl delete crd zfssnapshots.openebs.io
