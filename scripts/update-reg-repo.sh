#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
ROOT_DIR="$SCRIPT_DIR/.."
CHART_DIR="$ROOT_DIR/deploy/helm/charts"
VALUES_YAML="$CHART_DIR/values.yaml"

source "$SCRIPT_DIR/yq_utils.sh"
source "$SCRIPT_DIR/log.sh"

help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --registry                                The registry to be updated to.
  --repository                              The repository to be updated to.

Examples:
  $(basename "$0") --registry ghcr.io --repository openebs/dev/zfs-driver
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case $1 in
    -h|--help)
      help
      exit 0
      ;;
    --registry)
      shift
      NEW_REGISTRY=$1
      shift
      ;;
    --repository)
      shift
      NEW_REPOSITORY=$1
      shift
      ;;
    *)
      help
      log_fatal "Unknown option: $1"
      ;;
  esac
done

if [ -z "${NEW_REGISTRY:-}" ]; then
  log_fatal "Missing required flag: --registry"
fi

if [ -z "${NEW_REPOSITORY:-}" ]; then
  log_fatal "Missing required flag: --repository"
fi

yq_ibl ".zfsPlugin.image.registry = \"$NEW_REGISTRY\"" "$VALUES_YAML"
yq_ibl ".zfsPlugin.image.repository = \"$NEW_REPOSITORY\"" "$VALUES_YAML"
