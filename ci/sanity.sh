#!/bin/bash

# Copyright 2020 The OpenEBS Authors
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

set -ex
test_repo="kubernetes-csi"

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(kubectl get pods -l app=openebs-zfs-node -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  kubectl describe po "$AgentPOD" -n kube-system
  printf "\n\n"
  kubectl logs --tail="${NR}" "$AgentPOD" -n kube-system -c openebs-zfs-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(kubectl get pods -l app=openebs-zfs-controller -o jsonpath='{.items[0].metadata.name}' -n kube-system)
  kubectl describe po "$ControllerPOD" -n kube-system
  printf "\n\n"
  kubectl logs --tail="${NR}" "$ControllerPOD" -n kube-system -c openebs-zfs-plugin
  printf "\n\n"
}

function dumpAllLogs() {
	echo "========================= Dump All logs ========================"
	dumpControllerLogs 1000
	dumpAgentLogs 1000
}

function initializeCSISanitySuite() {
	echo "=============== Initialize CSI Sanity test suite ==============="
       cat <<EOT >> /tmp/parameters.json
{
        "node": "$HOSTNAME",
        "poolname": "zfspv-pool",
        "wait": "yes",
        "thinprovision": "yes"
}
EOT

	sudo rm -rf /tmp/csi.sock
	CSI_TEST_REPO="https://github.com/$test_repo/csi-test.git"
	CSI_REPO_PATH="$GOPATH/src/github.com/$test_repo/csi-test"
	if [ ! -d "$CSI_REPO_PATH" ] ; then
		git clone -b "v4.0.1" "$CSI_TEST_REPO" "$CSI_REPO_PATH"
	else
		cd "$CSI_REPO_PATH"
		git pull "$CSI_REPO_PATH"
	fi

	cd "$CSI_REPO_PATH/cmd/csi-sanity"
	make clean
	make

	UUID=$(kubectl get pod -n kube-system openebs-zfs-controller-0 -o 'jsonpath={.metadata.uid}')
	SOCK_PATH=/var/lib/kubelet/pods/"$UUID"/volumes/kubernetes.io~empty-dir/socket-dir/csi.sock

	sudo chmod -R 777 /var/lib/kubelet
	sudo ln -s "$SOCK_PATH" /tmp/csi.sock
	sudo chmod -R 777 /tmp/csi.sock
}

function startTestSuite() {
	echo "================== Start csi-sanity test suite ================="
	./csi-sanity --ginkgo.v --csi.controllerendpoint=///tmp/csi.sock --csi.endpoint=/var/lib/kubelet/plugins/zfs-localpv/csi.sock --csi.testvolumeparameters=/tmp/parameters.json --csi.testsnapshotparameters=/tmp/parameters.json
	if [ $? -ne 0 ];
	then
		dumpAllLogs
		exit 1
	fi
	exit 0
}

initializeCSISanitySuite

# do not exit in case of error, let us print the logs
set +e

startTestSuite
