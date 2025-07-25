{{ define "platform.volumesEmpty.tpl" }}
{{ end }}
{{ define "platform.volumes.test.tpl" }}
volumes:
  - name: testVolume
    configMap:
      name: {{ include "chart.fullname" . }}
{{ end }}
{{ define "platform.volumes.tpl" }}
volumes:
  - name: config
    configMap:
      name: {{ include "chart.fullname" . }}
  {{- if or (contains "all" .Values.mode) (contains "kas" .Values.mode) }}
  - name: kas-private-keys
    secret:
      secretName: {{ coalesce .Values.services.kas.private_keys_secret .Values.services.kas.privateKeysSecret }}
    {{- if .Values.server.tls.enabled }}
  {{- end }}
  - name: tls
    secret:
      secretName: {{ .Values.server.tls.secret | default (printf "%s-tls" (include "chart.fullname" .)) }}
    {{- end }}
    {{- if or (and .Values.playground .Values.keycloak.ingress.enabled .Values.keycloak.ingress.tls) .Values.server.tls.additionalTrustedCerts }}
  - name: trusted-certs
    projected:
      sources:
      {{- if and .Values.playground .Values.keycloak.ingress.enabled .Values.keycloak.ingress.tls }}
        - secret:
            name: {{ .Values.keycloak.ingress.hostname }}-tls # If the fullnameOverride is set, this will break
            optional: false
            items:
              - key: ca.crt
                path: kc-ca.crt
        {{- end -}}
        {{- with .Values.server.tls.additionalTrustedCerts }}
        {{- toYaml . | nindent 12 }}
        {{- end }}
    {{- end }}
  {{- with .Values.volumes }}
  {{- toYaml . | nindent 2 }}
  {{- end }}
{{ end }}


{{ define "platform.volumeMountsEmpty.tpl" }}
{{ end }}
{{ define "platform.volumeMounts.test.tpl" }}
volumeMounts:
  - name: testVolumeMount
    readOnly: true
    mountPath: /etc/platform/config
{{ end }}
{{ define "platform.volumeMounts.tpl" }}
volumeMounts:
  - name: config
    readOnly: true
    mountPath: /etc/platform/config
  {{- if or (contains "all" .Values.mode ) (contains "kas" .Values.mode) }}
  - name: kas-private-keys
    readOnly: true
    mountPath: /etc/platform/kas
  {{- end }}
  {{- if or (and .Values.playground .Values.keycloak.ingress.enabled .Values.keycloak.ingress.tls) .Values.server.tls.additionalTrustedCerts }}
  - name: trusted-certs
    readOnly: true
    mountPath: /etc/ssl/certs/platform
  {{- end }}
  {{- if .Values.server.tls.enabled }}
  - name: tls
    readOnly: true
    mountPath: /etc/platform/certs
  {{- end -}}
  {{- with .Values.volumeMounts }}
  {{- toYaml . | nindent 2 }}
  {{- end }}
{{ end }}