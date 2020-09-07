
# Copyright © 2020 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#!/bin/bash

# do not provision/deprovision anything while running the upgrade script.

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
	# ZFSVolumes: create the new CR with apiVersion as zfs.openebs.io and kind as Volume

	kubectl get zfsvolumes.openebs.io -n $ZFSPV_NAMESPACE -oyaml > volumes.yaml

	# update the group name to zfs.openebs.io
	sed -i "s/apiVersion: openebs.io/apiVersion: zfs.openebs.io/g" volumes.yaml
	# create the new CR
	kubectl apply -f volumes.yaml

	rm volumes.yaml
fi

echo "Fetching ZFS Snapshots"
numSnap=`kubectl get zfssnapshots.openebs.io --no-headers -n $ZFSPV_NAMESPACE | wc -l`

if [ $numSnap -gt 0 ]; then
	# ZFSSnapshots: create the new CR with apiVersion as zfs.openebs.io and kind as Snapshot

	kubectl get zfssnapshots.openebs.io -n $ZFSPV_NAMESPACE -oyaml > snapshots.yaml


	# update the group name to zfs.openebs.io
	sed -i "s/apiVersion: openebs.io/apiVersion: zfs.openebs.io/g" snapshots.yaml
	# create the new CR
	kubectl apply -f snapshots.yaml

	rm snapshots.yaml
fi
