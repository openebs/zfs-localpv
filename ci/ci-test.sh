#!/usr/bin/env bash

set -eu

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
SNAP_CLASS=deploy/sample/zfssnapclass.yaml
export OPENEBS_NAMESPACE=${OPENEBS_NAMESPACE:-openebs}
TEST_DIR="$SCRIPT_DIR"/../tests
DEFAULT_TEST_PROFILES="regular"
ALLOWED_TEST_PROFILES="regular custom-node-id all"

CRDS_TO_DELETE_ON_CLEANUP="zfsrestores.zfs.openebs.io zfssnapshots.zfs.openebs.io zfsvolumes.zfs.openebs.io zfsbackups.zfs.openebs.io zfsnodes.zfs.openebs.io"

help() {
  cat <<EOF >&2
Usage: $(basename "${0}") [COMMAND] [OPTIONS]

Commands:
  run                          Run the tests.
  load                         Build and load the image into the K8s cluster.
  install                      Install helm chart and wait for it to be ready.
  clean                        Clean the leftovers.

Options:
  -h, --help                   Display this text.

Options for run:
  -r, --reset                  Clean before running the tests.
  -x, --no-cleanup             Don't cleanup after running the tests.
  -b, --build-always           Build and load the images before running the tests. [ By default image is built if not present only ]
  -p, --profile      "<str>"   The test profile to run [ default: $DEFAULT_TEST_PROFILES ] [ allowed: $ALLOWED_TEST_PROFILES ]

Examples:
  $(basename "${0}") run -rxb -p "all"
EOF
}

echo_err() {
  echo -e "ERROR: $1" >&2
}

needs_help() {
  [ -n "${1:-}" ] && echo_err "$1\n"
  help
  exit 1
}

die() {
  echo_err "FATAL: $1"
  exit 1
}

cleanup_loop_zfs() {
  for device in $(losetup -l -J | jq -r '.loopdevices[]|select(."back-file" == "/tmp/disk.img")|.name'); do
    echo "Found stale loop device: $device"

    sudo "$(which zpool)" destroy -f zfspv-pool || :
    sudo losetup -d "$device" 2>/dev/null || :
    rm "/tmp/disk.img" || :
  done
}

cleanup() {
  set +e

  echo "Cleaning up test resources"

  if kubectl get nodes 2>/dev/null; then
    kubectl delete deployment -lrole=test -n "$OPENEBS_NAMESPACE"
    kubectl delete pod -lrole=test --force -n "$OPENEBS_NAMESPACE"
    kubectl delete pvc -n "$OPENEBS_NAMESPACE" --all

    sleep 3

    # shellcheck disable=SC2068
    for cr in ${CRDS_TO_DELETE_ON_CLEANUP[@]}; do
      kubectl delete "$cr" -n "$OPENEBS_NAMESPACE" --all
    done

    if helm uninstall zfs-localpv -n "$OPENEBS_NAMESPACE" --ignore-not-found --timeout=1m --wait; then
      # shellcheck disable=SC2086
      kubectl delete crds $CRDS_TO_DELETE_ON_CLEANUP
      kubectl delete -f "${SNAP_CLASS}"

      kubectl delete pod -lrole=openebs-zfs --force -n "$OPENEBS_NAMESPACE"
    fi
  fi

  cleanup_loop_zfs

  set -e
  # always return true
  return 0
}

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(kubectl get pods -l app=openebs-zfs-node -o jsonpath='{.items[0].metadata.name}' -n "$OPENEBS_NAMESPACE")
  kubectl describe po "$AgentPOD" -n "$OPENEBS_NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$AgentPOD" -n "$OPENEBS_NAMESPACE" -c openebs-zfs-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(kubectl get pods -l app=openebs-zfs-controller -o jsonpath='{.items[0].metadata.name}' -n "$OPENEBS_NAMESPACE")
  kubectl describe po "$ControllerPOD" -n "$OPENEBS_NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$ControllerPOD" -n "$OPENEBS_NAMESPACE" -c openebs-zfs-plugin
  printf "\n\n"
}

dump_logs() {
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
  kubectl get zfsvolumes.zfs.openebs.io -n "$OPENEBS_NAMESPACE" -oyaml

  echo "get zfs snapshot details"
  kubectl get zfssnapshots.zfs.openebs.io -n "$OPENEBS_NAMESPACE" -oyaml
}

isPodReady(){
  [ "$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n "$OPENEBS_NAMESPACE")" = 'True' ] &&
  [ "$(kubectl get po "$1" -o 'jsonpath={.metadata.deletionTimestamp}' -n "$OPENEBS_NAMESPACE")" = "" ]
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
    zfsDriver="$(kubectl get pods -l role=openebs-zfs -o 'jsonpath={.items[*].metadata.name}' -n "$OPENEBS_NAMESPACE")"
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

helm_install() {
  cleanup_loop_zfs
  truncate -s 100G /tmp/disk.img
  sudo zpool create zfspv-pool "$(sudo losetup -f /tmp/disk.img --show)"

  # Setup encryption key for testing encrypted volumes
  echo "Setting up encryption key for encrypted volume tests"
  sudo mkdir -p /etc/zfs/keys
  sudo dd if=/dev/urandom of=/etc/zfs/keys/zfspool0.key bs=32 count=1
  sudo chmod 600 /etc/zfs/keys/zfspool0.key

  helm install zfs-localpv ./deploy/helm/charts -n "$OPENEBS_NAMESPACE" --create-namespace --set zfsPlugin.image.pullPolicy=Never --set analytics.enabled=false --set backupGC.enabled=true --set 'zfsNode.defaultFormatOptions.ext4=-b 2048'
  kubectl apply -f "$SNAP_CLASS"

  waitForZFSDriver

  kubectl get pods -n "$OPENEBS_NAMESPACE"
}

runTestSuite() {
  local coverageFile=$1
  local labelFilter="$2"

  # wait for zfs-driver to be up
  waitForZFSDriver

  cd "$TEST_DIR"

  echo "running ginkgo test case with coverage ${coverageFile}"

  if ! ginkgo -v -coverprofile="${coverageFile}" --label-filter="${labelFilter}" -covermode=atomic; then
    dump_logs
    [ "$CLEAN_AFTER" = "true" ] && cleanup
    exit 1
  fi
}

prepareCustomNodeIdEnv() {
  for node in $(kubectl get nodes -n "$OPENEBS_NAMESPACE" -o jsonpath='{.items[*].metadata.name}'); do
    local zfsNode
    zfsNode=$(kubectl get zfsnode -n "$OPENEBS_NAMESPACE" -o jsonpath="{.items[?(@.metadata.ownerReferences[0].name=='${node}')].metadata.name}")
    echo "Relabeling node ${node} with ${node}-custom-id"
    kubectl label node "${node}" openebs.io/nodeid="${node}-custom-id" --overwrite

    local nodeDriver
    nodeDriver=$(kubectl get pods -l name=openebs-zfs-node -o jsonpath="{.items[?(@.spec.nodeName=='${node}')].metadata.name}" -n "$OPENEBS_NAMESPACE")
    echo "Restarting ${nodeDriver} on ${node} to pick up the new node id"
    kubectl delete pod "${nodeDriver}" -n "$OPENEBS_NAMESPACE"

    echo "Deleting old zfsnode ${zfsNode}"
    kubectl delete zfsnode "${zfsNode}" -n "$OPENEBS_NAMESPACE"
  done

  kubectl get po -n "$OPENEBS_NAMESPACE"
}

load_k3s() {
  if [ "${CI_K3S:-}" = "true" ]; then
    local img="${1:-}"
    if [ -z "${1:-}" ]; then
      repo="$(make image-repo -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
      tag="$(make image-tag -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
      img="$repo:$tag"
    fi
    docker save "$img" | ctr images import -
  fi
}

load_image() {
  make zfs-driver-image
  load_k3s "${1:-}"
}

maybe_load_image() {
  local repo tag img did cid

  if [ "$BUILD_ALWAYS" = "true" ]; then
    load_image
    return 0
  fi

  repo="$(make image-repo -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
  tag="$(make image-tag -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
  img="$repo:$tag"

  did="$(docker image ls --no-trunc --format json | jq -r --arg repo "$repo" --arg tag "$tag" 'select(.Repository == $repo and .Tag == $tag)|.ID')"
  if [ -z "$did" ]; then
    make zfs-driver-image
  fi

  if ! [ "${CI_K3S:-}" = "true" ]; then
    return 0
  fi

  cid="$(crictl image --output json | jq -r --arg image "docker.io/$(make image-ref -s -C "$SCRIPT_DIR"/.. 2>/dev/null)" '.images[]|select(.repoTags[0] == $image)|.id')"
  # If image not present, or different to the docker source, then rebuild it!
  if [ -z "$cid" ] || [ "$cid" != "$did" ]; then
    load_image "$img"
    return 0
  fi

  return 0
}

set_test_profiles() {
  local profiles="$1"
  local set_profiles=""

  for profile in $profiles; do
    if [[ ! " $ALLOWED_TEST_PROFILES " =~ " $profile " ]]; then
      needs_help "$profile is not allowed!"
    fi
    if [[ ! " $set_profiles " =~ " $profile " ]]; then
      set_profiles="$set_profiles $profile"
    fi
  done

  echo "$set_profiles"
}

has_profile() {
  local profiles="$1"
  local profile="$2"

  [[ " $profiles " =~ " $profile " ]] || [[ " $profiles " =~ " all " ]]
}

# allow override
if [ -z "${KUBECONFIG:-}" ]
then
  export KUBECONFIG="${HOME}/.kube/config"
fi

COMMAND=
CLEAN_BEFORE="false"
CLEAN_AFTER="true"
BUILD_ALWAYS="false"
TEST_PROFILES="$DEFAULT_TEST_PROFILES"

while test $# -gt 0; do
  arg="$1"
  case "$arg" in
    run | clean | load | install)
      [ -n "$COMMAND" ] && needs_help "Can't specify two commands"
      COMMAND="$1"
      ;;
    -r | --reset)
      CLEAN_BEFORE="true"
      ;;
    -x | --no-cleanup)
      CLEAN_AFTER="false"
      ;;
    -b | --build-always)
      BUILD_ALWAYS="true"
      ;;
    -h | --help)
      needs_help
      ;;
    -p | --profile)
      test $# -lt 2 && needs_help "Missing value for the profile argument '$arg'."
      TEST_PROFILES="$(set_test_profiles "$2")"
      shift
      ;;
    -*)
      singleLetterOpts="${1:1}"
      while [ -n "$singleLetterOpts" ]; do
        case "${singleLetterOpts:0:1}" in
          r)
            CLEAN_BEFORE="true"
            ;;
          x)
            CLEAN_AFTER="false"
            ;;
          b)
            BUILD_ALWAYS="true"
            ;;
          *)
            needs_help "Unrecognized argument $singleLetterOpts"
            ;;
        esac
        singleLetterOpts="${singleLetterOpts:1}"
      done
      ;;
    *)
      needs_help "Unrecognized argument $1"
      ;;
  esac
  shift
done

case "$COMMAND" in
  clean)
    cleanup
    ;;
  load)
    load_image
    ;;
  install)
    helm_install
    ;;
  run)
    # trap "cleanup 2>/dev/null" EXIT
    [ "$CLEAN_BEFORE" = "true" ] && cleanup

    maybe_load_image

    helm_install

    if has_profile "$TEST_PROFILES" "regular"; then
      runTestSuite bdd_coverage.txt "!custom-node-id"
    fi

    if has_profile "$TEST_PROFILES" "custom-node-id"; then
      prepareCustomNodeIdEnv
      runTestSuite bdd_coverage_custom-node-id.txt "custom-node-id"
    fi

    printf "\n\n"
    echo "######### All test cases passed #########"

    [ "$CLEAN_AFTER" = "true" ] && cleanup
    ;;
  *)
    needs_help "Missing Command"
    ;;
esac
