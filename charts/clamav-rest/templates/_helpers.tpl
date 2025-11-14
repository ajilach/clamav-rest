{{- define "clamav-rest.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "clamav-rest.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "clamav-rest.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "clamav-rest.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "clamav-rest.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "clamav-rest.selectorLabels" -}}
app.kubernetes.io/name: {{ include "clamav-rest.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

