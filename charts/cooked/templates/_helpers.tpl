{{/*
Expand the name of the chart.
*/}}
{{- define "cooked.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncated to 63 characters (DNS label limit).
*/}}
{{- define "cooked.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version for the helm.sh/chart label.
*/}}
{{- define "cooked.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Standard labels for all resources.
*/}}
{{- define "cooked.labels" -}}
helm.sh/chart: {{ include "cooked.chart" . }}
{{ include "cooked.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels (immutable — never include version here).
*/}}
{{- define "cooked.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cooked.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common annotations for all resources.
*/}}
{{- define "cooked.annotations" -}}
{{- with .Values.commonAnnotations }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Service account name.
*/}}
{{- define "cooked.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cooked.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container image string from registry/repository/tag/digest.
When using appVersion as the default tag, prepends "v" to match GHCR tag format (v1.3.2).
An explicit image.tag is used as-is.
*/}}
{{- define "cooked.image" -}}
{{- if .Values.image.digest }}
{{- printf "%s/%s@%s" .Values.image.registry .Values.image.repository .Values.image.digest }}
{{- else if .Values.image.tag }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.tag }}
{{- else }}
{{- printf "%s/%s:v%s" .Values.image.registry .Values.image.repository .Chart.AppVersion }}
{{- end }}
{{- end }}
