{{/*
Expand the name of the chart.
*/}}
{{- define "kube-copilot-ui.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Full name.
*/}}
{{- define "kube-copilot-ui.fullname" -}}
{{- if .Release.Name | eq .Chart.Name -}}
{{ .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- else -}}
{{ printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "kube-copilot-ui.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "kube-copilot-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "kube-copilot-ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-copilot-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name.
*/}}
{{- define "kube-copilot-ui.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kube-copilot-ui.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
