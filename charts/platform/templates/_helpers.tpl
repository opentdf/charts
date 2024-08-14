{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "chart.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "chart.labels" -}}
helm.sh/chart: {{ include "chart.chart" . }}
{{ include "chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "chart.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "chart.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}



{{- define "var_dump" -}}
{{- . | mustToPrettyJson | printf "\nThe JSON output of the dumped var is: \n%s" | fail }}
{{- end -}}

{{- define "platform.configFileName" -}}
{{- printf "%s.yaml" ( .Values.configFileKey | default "opentdf" ) }}
{{- end -}}

{{- define "platform.envVarPrefix" -}}
{{- printf "%s" ( .Values.configFileKey | default "opentdf" | upper ) }}
{{- end -}}

{{- define "sdk_config.validate" -}}
{{- if and ( .Values.sdk_config.clientsecret) ( .Values.sdk_config.existingSecret.name) ( .Values.sdk_config.existingSecret.key)}}
{{- fail "You cannot set both clientsecret and existingSecret in sdk_config." }}
{{- end -}}
{{- if and (not (or (contains "core" .Values.mode) (contains "all" .Values.mode))) (and (not .Values.sdk_config.clientsecret) (not .Values.sdk_config.existingSecret.name) (not .Values.sdk_config.existingSecret.key)) }}
{{- fail "Mode does not contain 'core' or 'all'. You must configure the sdk_config" }}
{{- end }}
{{- end -}}

{{- /*
platform.util.merge will merge two YAML templates and output the result.
This takes an array of three values:
- the top context
- the template name of the overrides (destination)
- the template name of the base (source)
*/ -}}
{{- define "platform.util.merge.list" -}}
{{- $top := first . -}}
{{- $filterKey := (index . 1) }}
{{- $overrides := fromYaml (include (index . 2) $top) | default (dict) -}}
{{- $tpl := fromYaml (include (index . 3) $top) | default (dict) -}}

{{- $mergedList := index $tpl $filterKey | default (list) -}}

{{- range $key, $values := $overrides -}}
  {{- if kindIs "slice" $values }}
    {{- range $key2, $value := $values }}
        {{- $mergedList = append $mergedList $value -}}
    {{- end }}
  {{- end -}}
{{- end -}}

{{- (dict $filterKey $mergedList) | toYaml }}

{{- end -}}