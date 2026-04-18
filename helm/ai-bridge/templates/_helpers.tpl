{{/*
Expand the name of the chart.
*/}}
{{- define "ai-bridge.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ai-bridge.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "ai-bridge.backend.fullname" -}}
{{- printf "%s-backend" (include "ai-bridge.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "ai-bridge.frontend.fullname" -}}
{{- printf "%s-frontend" (include "ai-bridge.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ai-bridge.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "ai-bridge.backend.labels" -}}
{{ include "ai-bridge.labels" . }}
app.kubernetes.io/name: {{ include "ai-bridge.name" . }}-backend
app.kubernetes.io/component: backend
{{- end }}

{{- define "ai-bridge.frontend.labels" -}}
{{ include "ai-bridge.labels" . }}
app.kubernetes.io/name: {{ include "ai-bridge.name" . }}-frontend
app.kubernetes.io/component: frontend
{{- end }}

{{- define "ai-bridge.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ai-bridge.name" . }}-backend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "ai-bridge.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ai-bridge.name" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Full image reference with optional registry prefix.
Usage: include "ai-bridge.image" (dict "registry" .Values.imageRegistry "repository" .Values.backend.image.repository "tag" .Values.backend.image.tag)
*/}}
{{- define "ai-bridge.image" -}}
{{- $repo := .repository -}}
{{- if .registry -}}
{{- printf "%s/%s:%s" .registry $repo .tag -}}
{{- else -}}
{{- printf "%s:%s" $repo .tag -}}
{{- end -}}
{{- end }}
