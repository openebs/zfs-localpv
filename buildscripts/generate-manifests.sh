#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# Make unmatched globs expand to nothing (prevents mv errors when no matches)
shopt -s nullglob

# Zero-width space (U+200B) sanitizer
ZWSP=$'\u200b'
sanitize() { printf '%s' "${1//$ZWSP/}"; }

# Resolve important paths and sanitize any hidden U+200B characters
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]:-"$0"}")" && pwd -P)"
SCRIPT_DIR="$(sanitize "${SCRIPT_DIR}")"
ROOT_DIR="$(sanitize "${SCRIPT_DIR}/..")"
DEPLOY_YAML_DIR="$(sanitize "${ROOT_DIR}/deploy/yamls")"
HELM_CHART_DIR="$(sanitize "${ROOT_DIR}/deploy/helm/charts")"
CRD_CHART_TEMPLATE_DIR="$(sanitize "${HELM_CHART_DIR}/charts/crds/templates")"

RELEASE_NAME="openebs"
RELEASE_NAMESPACE="kube-system"

CONTROLLER_GEN="$(command -v controller-gen || true)"
HELM_BIN="$(command -v helm || true)"

if [[ -z "${CONTROLLER_GEN}" ]]; then
  echo "ERROR: controller-gen not found. Please run 'make bootstrap' to install it." >&2
  exit 1
fi

if [[ -z "${HELM_BIN}" ]]; then
  echo "ERROR: helm not found. Please install Helm (https://helm.sh)." >&2
  exit 1
fi

# Ensure output directories exist
mkdir -p -- "${DEPLOY_YAML_DIR}" "${CRD_CHART_TEMPLATE_DIR}"

# If a stray 'yamls' directory with a trailing U+200B exists, consolidate and remove it
if [[ -d "${DEPLOY_YAML_DIR}${ZWSP}" ]]; then
  shopt -s dotglob nullglob
  mkdir -p -- "${DEPLOY_YAML_DIR}"
  mv -- "${DEPLOY_YAML_DIR}${ZWSP}/"* "${DEPLOY_YAML_DIR}/" 2>/dev/null || true
  rmdir -- "${DEPLOY_YAML_DIR}${ZWSP}" 2>/dev/null || true
fi

echo "+ Generating ZFS LocalPV CRDs"
# Generate v1 CRDs into the sanitized path
"${CONTROLLER_GEN}" crd:crdVersions=v1 \
  paths=./pkg/apis/... \
  output:crd:artifacts:config="${DEPLOY_YAML_DIR}"

# Rename and copy generated CRDs
for FILE in "${DEPLOY_YAML_DIR}"/zfs.openebs.io_*; do
  BASE_NAME="$(basename -- "${FILE}" | sed -e 's/^zfs.openebs.io_//' -e 's/s\.yaml$/.yaml/')"
  NEW_FILE="${DEPLOY_YAML_DIR}/${BASE_NAME%.yaml}-crd.yaml"
  mv -- "${FILE}" "${NEW_FILE}"

  TARGET_FILE="${CRD_CHART_TEMPLATE_DIR}/${BASE_NAME%.yaml}.yaml"
  install -m 0644 -- "${NEW_FILE}" "${TARGET_FILE}"

  # Append Helm annotations and enable/disable wrapper
  awk '/controller-gen.kubebuilder.io\/version:/ { print; print "    {{- include \"crds.extraAnnotations\" .Values.zfsLocalPv | nindent 4 }}"; next }1' \
    "${TARGET_FILE}" > "${TARGET_FILE}.tmp" && mv -- "${TARGET_FILE}.tmp" "${TARGET_FILE}"

  awk 'BEGIN { print "{{- if .Values.zfsLocalPv.enabled -}}" } { print } END { if (NR > 0) print "{{- end -}}" }' \
    "${TARGET_FILE}" > "${TARGET_FILE}.tmp" && mv -- "${TARGET_FILE}.tmp" "${TARGET_FILE}"
done

# Render the operator bundle
"${HELM_BIN}" template "${RELEASE_NAME}" "${HELM_CHART_DIR}" -n "${RELEASE_NAMESPACE}" \
  --set analytics.installerType="zfs-operator" \
  --set crds.zfsLocalPv.keep=false \
  --set crds.csi.volumeSnapshots.keep=false \
  --set enableHelmMetaLabels=false \
  > "${ROOT_DIR}/deploy/zfs-operator.yaml"

echo "+ Manifests written to: ${DEPLOY_YAML_DIR} and ${ROOT_DIR}/deploy/zfs-operator.yaml"
