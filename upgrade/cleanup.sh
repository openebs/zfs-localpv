#!/bin/bash

set -e

if [ -z $1 ]; then
	# default namespace is openebs when all the custom resources are created
	ZFSPV_NAMESPACE="openebs"
else
	ZFSPV_NAMESPACE=$1
fi

echo "Fetching ZFS Volumes"
numVol=`kubectl get zfsvolumes.openebs.io --no-headers -n $ZFSPV_NAMESPACE | wc -l`

if [ $numVol -gt 0 ]; then
	echo "Cleaning the ZFS Volumes($numVol)"
	kubectl get zfsvolumes.openebs.io -n $ZFSPV_NAMESPACE -oyaml > volumes.yaml

	# remove the finalizer from the old CR
	sed -i "/zfs.openebs.io\/finalizer/d" volumes.yaml
	kubectl apply -f volumes.yaml

	# delete the old CR
	kubectl delete -f volumes.yaml
fi

# delete the ZFSVolume CRD definition
kubectl delete crd zfsvolumes.openebs.io

numAttach=`kubectl get volumeattachment --no-headers | grep zfs.csi.openebs.io | wc -l`

if [ $numAttach -gt 0 ]; then
	echo "Cleaning the volumeattachment($numAttach)"
	# delete the volumeattachment object
	kubectl delete volumeattachment --all
fi

echo "Fetching ZFS Snapshots"
numSnap=`kubectl get zfssnapshots.openebs.io --no-headers -n $ZFSPV_NAMESPACE | wc -l`

if [ $numSnap -gt 0 ]; then
	echo "Cleaning the ZFS Snapshot($numSnap)"
	kubectl get zfssnapshots.openebs.io -n $ZFSPV_NAMESPACE -oyaml > snapshots.yaml

	# remove the finalizer from the old CR
	sed -i "/zfs.openebs.io\/finalizer/d" snapshots.yaml
	kubectl apply -f snapshots.yaml

	# delete the old CR
	kubectl delete -f snapshots.yaml
fi

# delete the ZFSSnapshot CRD definition
kubectl delete crd zfssnapshots.openebs.io
