#!/usr/bin/env bash

# Write output to error output stream.
echo_stderr() {
  echo -e "${1}" >&2
}

die()
{
  local _return="${2:-1}"
  echo_stderr "$1"
  exit "${_return}"
}

set -euo pipefail

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
ROOT_DIR="$SCRIPT_DIR/.."
CHART_DIR="$ROOT_DIR/deploy/helm/charts"
VALUES_YAML="$CHART_DIR/values.yaml"

NEW_REGISTRY="ghcr.io"
NEW_REPOSITORY="openebs/dev"

source "$SCRIPT_DIR/yq_utils.sh"

help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --registry                                The registry to be updated to.
  --repository                              The repository to be updated to.

Examples:
  $(basename "$0") --registry ghcr.io --repository openebs/dev
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
      die "Unknown option: $1"
      ;;
  esac
done

if [ -z "${NEW_REGISTRY:-}" ]; then
  die "Missing required flag: --registry"
fi

if [ -z "${NEW_REPOSITORY:-}" ]; then
  die "Missing required flag: --repository"
fi

yq_ibl ".zfsPlugin.image.registry = \"$NEW_REGISTRY\"" "$VALUES_YAML"
yq_ibl ".zfsPlugin.image.repository = \"$NEW_REPOSITORY\"" "$VALUES_YAML"
