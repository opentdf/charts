{{- define "platform.configurationEmpty.tpl" }}
{{ end }}
{{- define "platform.configuration.test.tpl" }}
services:
  testService:
    name: test
    db:
     host: {{ .Values.db.host }}
{{ end }}
{{- define "platform.configuration.tpl" }}
services:
  {{- if or (contains "all" .Values.mode ) (contains "core" .Values.mode ) }}
  {{- with (.Values.services).entityresolution }}
  {{- if .url }}
  entityresolution:
  {{ . | toYaml | nindent 8 }}
  {{- end }}
  {{- end }}
  {{- end }}
  {{- if or (contains "all" .Values.mode ) (contains "kas" .Values.mode) }}
  kas:
  {{- .Values.services.kas.config | toYaml | nindent 8 }}
  {{- end }}
  {{- if or (contains "all" .Values.mode) (contains "core" .Values.mode) }}
  {{- if .Values.services.authorization }}
  authorization:
  {{- .Values.services.authorization | toYaml | nindent 8 }}
  {{- end }}
  {{- end }}
  {{- with .Values.services.extraServices }}
  {{- . | toYaml | nindent 2}}
  {{- end }}
{{ end }}