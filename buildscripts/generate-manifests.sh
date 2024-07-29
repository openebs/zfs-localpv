#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
ROOT_DIR="$SCRIPT_DIR/.."
DEPLOY_YAML_DIR="$ROOT_DIR/deploy/yamls"
HELM_CHART_DIR="$ROOT_DIR/deploy/helm/charts/"
CRD_CHART_TEMPLATE_DIR="$HELM_CHART_DIR/charts/crds/templates"
CONTROLLER_GEN=$(which controller-gen)
RELEASE_NAME="openebs"
RELEASE_NAMESPACE="kube-system"

if [ "$CONTROLLER_GEN" = "" ]; then
  echo "ERROR: failed to get controller-gen, Please run make bootstrap to install it";
  exit 1;
fi

$CONTROLLER_GEN crd:trivialVersions=false,preserveUnknownFields=false paths=./pkg/apis/... output:crd:artifacts:config="$DEPLOY_YAML_DIR"

for FILE in "$DEPLOY_YAML_DIR"/zfs.openebs.io_*; do
  BASE_NAME=$(basename "$FILE" | sed -e 's/^zfs.openebs.io_//' -e 's/s\.yaml$/.yaml/')
  NEW_FILE="$DEPLOY_YAML_DIR/${BASE_NAME%.yaml}-crd.yaml"
  # Rename the files generated in the format <name>-crd.yaml
  mv "$FILE" "$NEW_FILE"

  # Copy the files to the crd subchart templates, in format <name>.yaml
  TARGET_FILE="$CRD_CHART_TEMPLATE_DIR/${BASE_NAME%.yaml}.yaml"
  cp "$NEW_FILE" "$TARGET_FILE"

  # Append the helm configuration to enable disable the keep the crds on uninstall.
  awk '/controller-gen.kubebuilder.io\/version:/ { print; print "    {{- include \"crds.extraAnnotations\" .Values.zfsLocalPv | nindent 4 }}"; next }1' "$TARGET_FILE" > "$TARGET_FILE.tmp" && mv "$TARGET_FILE.tmp" "$TARGET_FILE"
  # Append the helm configuration to enable disable the installation of the crds.
  awk 'BEGIN { print "{{- if .Values.zfsLocalPv.enabled -}}" } { print } END { if (NR > 0) print "{{- end -}}" }' "$TARGET_FILE" > "$TARGET_FILE.tmp" && mv "$TARGET_FILE.tmp" "$TARGET_FILE"
done

# Generate the zfs-operator.yaml using the helm chart.
helm template "$RELEASE_NAME" "$HELM_CHART_DIR" -n "$RELEASE_NAMESPACE" --set analytics.installerType="zfs-operator" > "$DEPLOY_YAML_DIR"/../zfs-operator.yaml
