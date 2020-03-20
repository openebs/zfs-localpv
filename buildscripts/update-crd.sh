#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
# Copyright 2020 The OpenEBS Authors.
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

#set -o errexit
set -o nounset
set -o pipefail

## find or download controller-gen
CONTROLLER_GEN=$(which controller-gen)

if [ "$CONTROLLER_GEN" = "" ]
then
  echo "ERROR: failed to get controller-gen, Please run make bootstrap to install it";
  exit 1;
fi

SCRIPT_ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)

$CONTROLLER_GEN crd:trivialVersions=true,preserveUnknownFields=false paths=${SCRIPT_ROOT}/pkg/apis/openebs.io/zfs/v1alpha1 output:crd:artifacts:config=deploy/yamls

## rename the crd yamls
mv deploy/yamls/zfs.openebs.io_zfssnapshots.yaml deploy/yamls/zfssnapshot-crd.yaml
mv deploy/yamls/zfs.openebs.io_zfsvolumes.yaml deploy/yamls/zfsvolume-crd.yaml

## create the operator file using all the yamls
cat deploy/yamls/namespace.yaml deploy/yamls/zfsvolume-crd.yaml deploy/yamls/zfssnapshot-crd.yaml deploy/yamls/zfs-operator.yaml > deploy/zfs-operator.yaml

# To use your own boilerplate text use:
#   --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
