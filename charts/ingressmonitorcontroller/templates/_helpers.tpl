{{/*
Expand the name of the chart.
*/}}
{{- define "ingress-monitor-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ingress-monitor-controller.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else if .Values.nameOverride }}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ingress-monitor-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ingress-monitor-controller.labels" -}}
helm.sh/chart: {{ include "ingress-monitor-controller.chart" . }}
{{ include "ingress-monitor-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Values.serviceManagedBy }}
{{- if .Values.global.labels }}
{{ toYaml .Values.global.labels }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ingress-monitor-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ingress-monitor-controller.name" . }}
app.kubernetes.io/instance: {{ .Values.name | default .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ingress-monitor-controller.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ingress-monitor-controller.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Verify that CRDs are installed
*/}}
{{- define "verify_crds_exist" -}}
  {{- if .Capabilities.APIVersions.Has "endpointmonitor.stakater.com/v1alpha1" -}}
    true
  {{- end -}}
{{- end -}}
