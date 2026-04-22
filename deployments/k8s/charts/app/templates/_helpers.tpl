{{- define "tech-feed-hub.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tech-feed-hub.fullname" -}}
{{- $name := default .Chart.Name .Values.global.nameOverride -}}
{{- if .Values.global.fullnameOverride -}}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "tech-feed-hub.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tech-feed-hub.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tech-feed-hub.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "tech-feed-hub.labels" -}}
helm.sh/chart: {{ include "tech-feed-hub.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: tech-feed-hub
{{ include "tech-feed-hub.selectorLabels" . }}
{{- end -}}

{{- define "tech-feed-hub.componentFullname" -}}
{{- printf "%s-%s" (include "tech-feed-hub.fullname" .root) .component | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tech-feed-hub.componentSelectorLabels" -}}
{{ include "tech-feed-hub.selectorLabels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "tech-feed-hub.componentLabels" -}}
{{ include "tech-feed-hub.labels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "tech-feed-hub.serviceAccountName" -}}
{{- if .Values.global.serviceAccount.create -}}
{{- default (include "tech-feed-hub.fullname" .) .Values.global.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.global.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "tech-feed-hub.imagePullSecrets" -}}
{{- $global := .root.Values.global.imagePullSecrets | default (list) -}}
{{- $local := .service.imagePullSecrets | default (list) -}}
{{- $all := concat $global $local -}}
{{- if gt (len $all) 0 }}
imagePullSecrets:
{{- range $all }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end -}}

{{- define "tech-feed-hub.httpProbe" -}}
httpGet:
  path: {{ .path | quote }}
  port: http
initialDelaySeconds: {{ default 5 .initialDelaySeconds }}
periodSeconds: {{ default 10 .periodSeconds }}
timeoutSeconds: {{ default 2 .timeoutSeconds }}
failureThreshold: {{ default 3 .failureThreshold }}
{{- end -}}

{{- define "tech-feed-hub.extraEnv" -}}
{{- range . }}
- name: {{ .name }}
  value: {{ .value | quote }}
{{- end }}
{{- end -}}

{{- define "tech-feed-hub.extraSecretEnv" -}}
{{- range . }}
- name: {{ .name }}
  valueFrom:
    secretKeyRef:
      name: {{ .secretName }}
      key: {{ .secretKey }}
      optional: {{ default false .optional }}
{{- end }}
{{- end -}}

{{- define "tech-feed-hub.apiGatewayUrl" -}}
{{- printf "http://%s:%v" (include "tech-feed-hub.componentFullname" (dict "root" . "component" "api-gateway")) .Values.services.apiGateway.service.port -}}
{{- end -}}

{{- define "tech-feed-hub.articleServiceUrl" -}}
{{- printf "http://%s:%v" (include "tech-feed-hub.componentFullname" (dict "root" . "component" "article-service")) .Values.services.articleService.service.port -}}
{{- end -}}

{{- define "tech-feed-hub.notificationServiceUrl" -}}
{{- printf "http://%s:%v" (include "tech-feed-hub.componentFullname" (dict "root" . "component" "notification-service")) .Values.services.notificationService.service.port -}}
{{- end -}}
