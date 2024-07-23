#!/usr/bin/env bash

set -e

SNAP_CLASS=deploy/sample/zfssnapclass.yaml
TEST_DIR="tests"

# Prepare env for running BDD tests
# Minikube is already running
helm install zfs-localpv ./deploy/helm/charts -n "$OPENEBS_NAMESPACE" --create-namespace --set zfsPlugin.pullPolicy=Never --set analytics.enabled=false
kubectl apply -f "$SNAP_CLASS"

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


isPodReady(){
  [ "$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n openebs)" = 'True' ]
}


isDriverReady(){
  for pod in $zfsDriver;do
    isPodReady "$pod" || return 1
  done
}


waitForZFSDriver() {
  period=120
  interval=1
  
  i=0
  while [ "$i" -le "$period" ]; do
    zfsDriver="$(kubectl get pods -l role=openebs-zfs -o 'jsonpath={.items[*].metadata.name}' -n openebs)"
    if isDriverReady "$zfsDriver"; then
      return 0
    fi

    i=$(( i + interval ))
    echo "Waiting for zfs-driver to be ready..."
    sleep "$interval"
  done

  echo "Waited for $period seconds, but all pods are not ready yet."
  return 1
}

runTestSuite() {
  local coverageFile=$1
  local labelFilter="$2"

  # wait for zfs-driver to be up
  waitForZFSDriver

  cd $TEST_DIR

  kubectl get po -n openebs

  set +e

  echo "running ginkgo test case with coverage ${coverageFile}"

  if ! ginkgo -v -coverprofile="${coverageFile}" --label-filter="${labelFilter}" -covermode=atomic; then

  sudo zpool status

  sudo zfs list -t all

  sudo zfs get all

  echo "******************** ZFS Controller logs***************************** "
  dumpControllerLogs 1000

  echo "********************* ZFS Agent logs *********************************"
  dumpAgentLogs 1000

  echo "get all the pods"
  kubectl get pods -owide --all-namespaces

  echo "get pvc and pv details"
  kubectl get pvc,pv -oyaml --all-namespaces

  echo "get snapshot details"
  kubectl get volumesnapshot.snapshot -oyaml --all-namespaces

  echo "get sc details"
  kubectl get sc --all-namespaces -oyaml

  echo "get zfs volume details"
  kubectl get zfsvolumes.zfs.openebs.io -n openebs -oyaml

  echo "get zfs snapshot details"
  kubectl get zfssnapshots.zfs.openebs.io -n openebs -oyaml

  exit 1
  fi
}

runTestSuite bdd_coverage.txt "!custom-node-id"

prepareCustomNodeIdEnv() {
  for node in $(kubectl get nodes -n openebs -o jsonpath='{.items[*].metadata.name}'); do
      local zfsNode=$(kubectl get zfsnode -n openebs -o jsonpath="{.items[?(@.metadata.ownerReferences[0].name=='${node}')].metadata.name}")
      echo "Relabeling node ${node} with ${node}-custom-id"
      kubectl label node "${node}" openebs.io/nodeid="${node}-custom-id" --overwrite

      local nodeDriver=$(kubectl get pods -l app=openebs-zfs-node -o jsonpath="{.items[?(@.spec.nodeName=='${node}')].metadata.name}" -n kube-system)
      echo "Restarting ${nodeDriver} on ${node} to pick up the new node id"
      kubectl delete pod "${nodeDriver}" -n kube-system

      echo "Deleting old zfsnode ${zfsNode}"
      kubectl delete zfsnode "${zfsNode}" -n openebs
  done
}

prepareCustomNodeIdEnv
runTestSuite bdd_coverage_custom-node-id.txt "custom-node-id"

printf "\n\n"
echo "######### All test cases passed #########"
exit 0
