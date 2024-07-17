#!/bin/bash

set -ex
test_repo="kubernetes-csi"

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(kubectl get pods -l app=openebs-zfs-node -o jsonpath='{.items[0].metadata.name}' -n openebs)
  kubectl describe po "$AgentPOD" -n openebs
  printf "\n\n"
  kubectl logs --tail="${NR}" "$AgentPOD" -n openebs -c openebs-zfs-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(kubectl get pods -l app=openebs-zfs-controller -o jsonpath='{.items[0].metadata.name}' -n openebs)
  kubectl describe po "$ControllerPOD" -n openebs
  printf "\n\n"
  kubectl logs --tail="${NR}" "$ControllerPOD" -n openebs -c openebs-zfs-plugin
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
	CSI_REPO_PATH="$(go env GOPATH)/src/github.com/$test_repo/csi-test"
	if [ ! -d "$CSI_REPO_PATH" ] ; then
		git clone -b "v4.0.1" "$CSI_TEST_REPO" "$CSI_REPO_PATH"
	else
		cd "$CSI_REPO_PATH"
		git pull "$CSI_REPO_PATH"
	fi

	cd "$CSI_REPO_PATH/cmd/csi-sanity"
	make clean
	make

	UUID=$(kubectl get pod -n openebs -l "openebs.io/component-name=openebs-zfs-controller" -o 'jsonpath={.items[0].metadata.uid}')
	SOCK_PATH=/var/lib/kubelet/pods/"$UUID"/volumes/kubernetes.io~empty-dir/socket-dir/csi.sock

	sudo chmod -R 777 /var/lib/kubelet
	sudo ln -s "$SOCK_PATH" /tmp/csi.sock
	sudo chmod -R 777 /tmp/csi.sock
}

function startTestSuite() {
	echo "================== Start csi-sanity test suite ================="
	if ! ./csi-sanity --ginkgo.v --csi.controllerendpoint=///tmp/csi.sock --csi.endpoint=/var/lib/kubelet/plugins/zfs-localpv/csi.sock --csi.testvolumeparameters=/tmp/parameters.json --csi.testsnapshotparameters=/tmp/parameters.json;
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
