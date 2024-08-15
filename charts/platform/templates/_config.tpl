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
  entityresolution: 
  {{ .Values.services.entityresolution | toYaml | nindent 8 }}
  {{- end }}
  {{- if or (contains "all" .Values.mode ) (contains "kas" .Values.mode) }}
  kas:
  {{- .Values.services.kas.config | toYaml | nindent 8 }}
  {{- end }}
  {{- if or (contains "all" .Values.mode) (contains "core" .Values.mode) }}
  authorization:
  {{ .Values.services.authorization | toYaml | nindent 8 }}
  {{- end }}
  {{- with .Values.services.extraServices }}
  {{- . | toYaml | nindent 2}}
  {{- end }}
{{ end }}