{{/*
    This returns a "1" if the CRD is absent in the cluster
    Usage:
      {{- if (include "crdIsAbsent" (list <crd-name>)) -}}
      # CRD Yaml
      {{- end -}}
*/}}
{{- define "crdIsAbsent" -}}
    {{- $crdName := index . 0 -}}
    {{- $crd := lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" $crdName -}}
    {{- $output := "1" -}}
    {{- if $crd -}}
        {{- $output = "" -}}
    {{- end -}}

    {{- $output -}}
{{- end -}}