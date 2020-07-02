# Copyright 2019 The OpenEBS Authors.
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

#!/usr/bin/env bash

set -e

export OPENEBS_NAMESPACE="openebs"
export NodeID=$HOSTNAME

ZFS_OPERATOR=deploy/zfs-operator.yaml
TEST_DIR="tests"

# Prepare env for runnging BDD tests
# Minikube is already running
sudo kubectl apply -f $ZFS_OPERATOR

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(sudo kubectl get pods -l app=openebs-zfs-node -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  sudo kubectl describe po $AgentPOD -n kube-system
  printf "\n\n"
  sudo kubectl logs --tail=${NR} $AgentPOD -n kube-system -c openebs-zfs-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(sudo kubectl get pods -l app=openebs-zfs-controller -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  sudo kubectl describe po $ControllerPOD -n kube-system
  printf "\n\n"
  sudo kubectl logs --tail=${NR} $ControllerPOD -n kube-system -c openebs-zfs-plugin
  printf "\n\n"
}


isPodReady(){
  [ "$(sudo kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n kube-system)" = 'True' ]
}


isDriverReady(){
  for pod in $zfsDriver;do
  isPodReady $pod || return 1
  done
}


waitForZFSDriver() {
  period=120
  interval=1
  
  i=0
  while [ "$i" -le "$period" ]; do
    zfsDriver="$(sudo kubectl get pods -o 'jsonpath={.items[*].metadata.name}' -n kube-system)"
    if isDriverReady $zfsDriver; then
      return 0
    fi

    i=$(( i + interval ))
    echo "Waiting for zfs-driver to be ready..."
    sleep "$interval"
  done

  

  echo "Waited for $period seconds, but all pods are not ready yet."
  return 1
}

# wait for zfs-driver to be up
waitForZFSDriver

cd $TEST_DIR

sudo kubectl get po -n kube-system

set +e

echo "running ginkgo test case"

ginkgo -v

if [ $? -ne 0 ]; then

sudo zpool status

sudo zfs list -t all

sudo zfs get all

echo "******************** ZFS Controller logs***************************** "
dumpControllerLogs 1000

echo "********************* ZFS Agent logs *********************************"
dumpAgentLogs 1000

echo "get all the pods"
sudo kubectl get pods -owide --all-namespaces

echo "get pvc and pv details"
sudo kubectl get pvc,pv -oyaml --all-namespaces

echo "get sc details"
sudo kubectl get sc --all-namespaces -oyaml

echo "get zfs volume details"
sudo kubectl get zfsvolumes.zfs.openebs.io -n openebs -oyaml

exit 1
fi

echo "\n\n######### All test cases passed #########\n\n"
