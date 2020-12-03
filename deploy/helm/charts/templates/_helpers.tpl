{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "openebs-zfslocalpv.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openebs-zfslocalpv.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "openebs-zfslocalpv.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openebs-zfslocalpv.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "openebs-zfslocalpv.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Define meta labels for openebs-zfslocalpv components
*/}}
{{- define "openebs-zfslocalpv.common.metaLabels" -}}
chart: {{ template "openebs-zfslocalpv.chart" . }}
heritage: {{ .Release.Service }}
openebs.io/version: {{ .Values.release.version | quote }}
{{- end -}}

{{/*
Create match labels for openebs-zfslocalpv controller
*/}}
{{- define "openebs-zfslocalpv.controller.matchLabels" -}}
app: {{ .Values.controller.componentName | quote }}
release: {{ .Release.Name }}
component: {{ .Values.controller.componentName | quote }}
{{- end -}}

{{/*
Create component labels for openebs-zfslocalpv controller
*/}}
{{- define "openebs-zfslocalpv.controller.componentLabels" -}}
openebs.io/component-name: {{ .Values.controller.componentName | quote }}
{{- end -}}


{{/*
Create labels for openebs-zfslocalpv controller
*/}}
{{- define "openebs-zfslocalpv.controller.labels" -}}
{{ include "openebs-zfslocalpv.common.metaLabels" . }}
{{ include "openebs-zfslocalpv.controller.matchLabels" . }}
{{ include "openebs-zfslocalpv.controller.componentLabels" . }}
{{- end -}}

{{/*
Create match labels for openebs-zfslocalpv node daemon
*/}}
{{- define "openebs-zfslocalpv.zfsNode.matchLabels" -}}
name: {{ .Values.zfsNode.componentName | quote }}
release: {{ .Release.Name }}
{{- end -}}

{{/*
Create component labels openebs-zfslocalpv node daemon
*/}}
{{- define "openebs-zfslocalpv.zfsNode.componentLabels" -}}
openebs.io/component-name: {{ .Values.zfsNode.componentName | quote }}
{{- end -}}


{{/*
Create labels for openebs-zfslocalpv node daemon
*/}}
{{- define "openebs-zfslocalpv.zfsNode.labels" -}}
{{ include "openebs-zfslocalpv.common.metaLabels" . }}
{{ include "openebs-zfslocalpv.zfsNode.matchLabels" . }}
{{ include "openebs-zfslocalpv.zfsNode.componentLabels" . }}
{{- end -}}




