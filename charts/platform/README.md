# platform

![Version: 0.10.1](https://img.shields.io/badge/Version-0.10.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.7.0](https://img.shields.io/badge/AppVersion-v0.7.0-informational?style=flat-square)

A Helm Chart for OpenTDF Platform

> [!WARNING]
> This chart is currently under active development and breaking changes may occur.

## Requirements

In order to run the platform you need a supported identity provider and postgres database. We have included the following charts as dependencies:

- [Keycloak](https://www.keycloak.org/guides)
  - [Helm Chart](https://artifacthub.io/packages/helm/bitnami/keycloak)

- [Postgres](https://www.postgresql.org/)
  - [Helm Chart](https://artifacthub.io/packages/helm/bitnami/postgresql)

If you want to deploy keycloak and postgresql as part of the platform deployment make sure to follow [Playground Configuration](#playground-configuration-optional) section.

## Pre-Requisites

### KAS Secret

In order to run the KAS server, you need to provide the KAS with a set a keys. This will allow the KAS to support nanotdf and tdf3 rewrap functionality.

#### Generate a set of KAS keys

```bash
openssl req -x509 -nodes -newkey RSA:2048 -subj "/CN=kas" -keyout kas-private.pem -out kas-cert.pem -days 365
openssl req -x509 -nodes -newkey ec:<(openssl ecparam -name secp256r1) -subj "/CN=kas" -keyout kas-ec-private.pem -out kas-ec-cert.pem -days 365
```

To create a secret with the keys, run the following command:

```bash
kubectl create secret generic kas-private-keys --from-file=./kas-cert.pem --from-file=./kas-ec-cert.pem --from-file=./kas-ec-private.pem --from-file=kas-private.pem
```

### HTTP/2 Support

The platform provides both gRPC and RESTful interfaces. Since gRPC requires HTTP/2 support, it's essential to configure your ingress with the appropriate annotations to handle this requirement.

Below are starting examples for configuring various ingress providers:

#### Openshift

Example values.yaml for edge terminated route:
```yaml
ingress:
  enabled: true
  annotations:
    route.openshift.io/termination: "edge"
  hosts:
    - host: platform.apps.okd.example.com
      paths:
        - backend:
            name: platform
            port:
              number: 9000
          path: ""
          pathType: ImplementationSpecific
```

### Playground Configuration (Optional)

If you want to test the platform quickly without having to worry about setting up a database and identity provider, you can use the following configuration:

> [!WARNING]
> :warning: ** Keycloak might need to be configured differently depending on your environment setup. Use the linked charts to dig into other configuration options.

```yaml
playground: true # Enable playground mode

# Configure keycloak ingress for access to keycloak
keycloak:
  ingress:
    annotations: {}
      # route.openshift.io/termination: edge
    hostname: # Your keycloak hostname (e.g. keycloak.example.com)
```

## Post Install

If everything is healthy you should be able to access the platform at the hostname you provided in the `ingress.hostname` field.

```sh
export PLATFORM_HOST=$(kubectl get ingress -l app.kubernetes.io/name=platform -o jsonpath='{.items[0].spec.rules[0].host}'  ) 
grpcurl -insecure $PLATFORM_HOST:443 kas.AccessService/PublicKey
```

### Post Install with Playground enabled

Download the [keycloak_data.yaml](https://raw.githubusercontent.com/opentdf/platform/main/service/cmd/keycloak_data.yaml)

 Run the following command to provision keycloak test data

  ```sh
    docker run \
         -v ./kc-config.yaml:/home/nonroot/.opentdf/kc-config.yaml \
         -it registry.opentdf.io/platform:nightly provision keycloak -f /home/nonroot/.opentdf/kc-config.yaml -e https://kc-nightly.apps.dsp.shp.virtru.us -u admin -p keycloak123
  ```

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Opentdf |  |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| oci://registry-1.docker.io/bitnamicharts | keycloak | 22.1.1 |
| oci://registry-1.docker.io/bitnamicharts | postgresql | 16.2.5 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment |
| autoscaling.enabled | bool | `false` | Enable horizontal autoscaling |
| autoscaling.maxReplicas | int | `100` | Maximum number of pods to run |
| autoscaling.minReplicas | int | `1` | Minimum number of pods to run |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization percentage |
| cache | object | `{"ristretto":{"max_cost":"1gb"}}` | Platform Cache Manager Settings |
| cache.ristretto.max_cost | string | `"1gb"` | Maximum cost (i.e. 1mb, 1gb) for the cache (default: 1gb) |
| configTemplate | string | `"platform.configurationEmpty.tpl"` |  |
| db.connect_timeout_seconds | int | `15` | Connection timeout duration (seconds). |
| db.database | string | `"opentdf"` | The database name |
| db.host | string | `"platform-db"` | The database host |
| db.password.secret.key | string | `"password"` | The key in the secret containing the database user password |
| db.password.secret.name | string | `"opentdf-db-credentials"` | The kubernetes secret containing the database user password |
| db.pool.health_check_period_seconds | int | `60` | Interval seconds per health check. |
| db.pool.max_connection_count | int | `4` | Maximum number of connections per pool. |
| db.pool.max_connection_idle_seconds | int | `1800` | Maximum seconds allowed for idle connection. |
| db.pool.max_connection_lifetime_seconds | int | `3600` | Maximum seconds per connection lifetime. |
| db.pool.min_connection_count | int | `0` | Minimum number of connections per pool. |
| db.pool.min_idle_connections_count | int | `0` | Minimum number of idle connections per pool. |
| db.port | int | `5432` | The database port |
| db.required | bool | `true` | If database connection info is required for the service (default: true) |
| db.runMigration | bool | `true` | Whether to run the database migration or not. |
| db.schema | string | `"opentdf"` | The schema for the database. |
| db.sslmode | string | `"prefer"` | The database ssl mode ( disable, prefer, require, verify-ca, verify-full ) |
| db.user | string | `"opentdf"` | The database user |
| envFrom | list | `[]` | Environment variables from a configmap or secret |
| extraEnv | list | `[]` | Extra environment variables to add to the container |
| fullnameOverride | string | `""` | Overrides the generated fullname |
| hostAliases | list | `[]` | Host Alias entries to add to pods |
| image.pullPolicy | string | `"IfNotPresent"` | The image tag to pull policy |
| image.repository | string | `"registry.opentdf.io/platform"` | The image repository to pull from |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | The image pull secrets to use |
| ingress.annotations | object | `{}` | Extra Ingress annotations |
| ingress.className | string | `""` | Ingress class name if required |
| ingress.enabled | bool | `false` | Enable Ingress |
| ingress.hosts | list | `[{"host":"chart-example.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]` | List of hosts for the ingress |
| ingress.tls | list | `[]` | List of tls hosts |
| keycloak.auth.adminUser | string | `"admin"` |  |
| keycloak.externalDatabase.database | string | `"opentdf"` |  |
| keycloak.externalDatabase.existingSecret | string | `"opentdf-db-credentials"` |  |
| keycloak.externalDatabase.existingSecretPasswordKey | string | `"password"` |  |
| keycloak.externalDatabase.host | string | `"platform-db"` |  |
| keycloak.externalDatabase.port | int | `5432` |  |
| keycloak.externalDatabase.user | string | `"opentdf"` |  |
| keycloak.extraEnvVars[0].name | string | `"KC_FEATURES"` |  |
| keycloak.extraEnvVars[0].value | string | `"preview,token-exchange"` |  |
| keycloak.fullnameOverride | string | `"platform-keycloak"` |  |
| keycloak.ingress.enabled | bool | `true` |  |
| keycloak.ingress.labels."opentdf.io/keycloak-admin" | string | `"false"` |  |
| keycloak.ingress.selfSigned | bool | `true` |  |
| keycloak.ingress.tls | bool | `true` |  |
| keycloak.keycloakConfigCli.configuration."opentdf.json" | string | `"{\n  \"realm\":\"opentdf\",\n  \"enabled\": true,\n  \"clients\": []\n}\n"` |  |
| keycloak.keycloakConfigCli.enabled | bool | `true` |  |
| keycloak.postgresql.enabled | bool | `false` |  |
| keycloak.proxyHeaders | string | `"xforwarded"` |  |
| keycloak.tls.autoGenerated | bool | `true` |  |
| keycloak.tls.enabled | bool | `true` |  |
| logger.level | string | `"info"` | The platform log level ( debug, info, warn, error ) |
| logger.output | string | `"stdout"` | The platform log output |
| logger.type | string | `"json"` | The platform log format ( json, text ) |
| mode | string | `"all"` | Mode defines the set of services to run (all, core, kas). Example mode: core,kas |
| nameOverride | string | `""` | Overrides the chart name |
| nodeSelector | object | `{}` | Target specific nodes in the cluster |
| playground | bool | `false` |  |
| podAnnotations | object | `{}` | Extra annotations to add to the pod |
| podDisruptionBudget.enabled | bool | `false` | Enable pod disruption budget |
| podDisruptionBudget.maxUnavailable | string | `nil` | Maximum number of pods that can be unavailble |
| podDisruptionBudget.minAvailable | string | `nil` | Minimum number of pods that must be available |
| podLabels | object | `{}` | Extra labels to add to the pod |
| podSecurityContext | object | `{"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}` | The pod security context (https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod) |
| postgresql.auth.database | string | `"opentdf"` |  |
| postgresql.auth.enablePostgresUser | bool | `false` |  |
| postgresql.auth.existingSecret | string | `"opentdf-db-credentials"` |  |
| postgresql.auth.secretKeys.userPasswordKey | string | `"password"` |  |
| postgresql.auth.usePasswordFiles | bool | `true` |  |
| postgresql.auth.username | string | `"opentdf"` |  |
| postgresql.fullnameOverride | string | `"platform-db"` |  |
| postgresql.tls.autoGenerated | bool | `true` |  |
| postgresql.tls.enabled | bool | `true` |  |
| replicaCount | int | `1` | The number of Platform pods to run |
| resources | object | `{}` | Resources to allocate to the container |
| sdk_config.client_id | string | `""` | Oauth2 Client Id |
| sdk_config.client_secret | string | `""` | Oauth2 Client Secret |
| sdk_config.connections | object | `{"core":{"endpoint":"","insecure":false,"plaintext":false},"entityresolution":{"endpoint":"","insecure":false,"plaintext":false}}` | Connection info to support different modes of operation. More connections can be added by adding a new section underneath connections. |
| sdk_config.connections.core | object | `{"endpoint":"","insecure":false,"plaintext":false}` | Defines connection info to an external Core Platform Service (e.g. This would be used if running a standalone KAS) |
| sdk_config.connections.core.endpoint | string | `""` | The core platform endpoint |
| sdk_config.connections.core.insecure | bool | `false` | Whether to verify the certificate of the core platform endpoint |
| sdk_config.connections.core.plaintext | bool | `false` | Plaintext Insecure Connection |
| sdk_config.connections.entityresolution | object | `{"endpoint":"","insecure":false,"plaintext":false}` | Defines connection info to an external Entity Resolution Service |
| sdk_config.connections.entityresolution.endpoint | string | `""` | The entity resolution endpoint |
| sdk_config.connections.entityresolution.insecure | bool | `false` | Whether to verify the certificate of the entity resolution endpoint |
| sdk_config.connections.entityresolution.plaintext | bool | `false` | Plaintext Insecure Connection |
| sdk_config.existingSecret | object | `{"key":"","name":""}` | Oauth2 Client Secret Kubernetes Secret |
| sdk_config.existingSecret.key | string | `""` | The key in the secret containing the client secret |
| sdk_config.existingSecret.name | string | `""` | The kubernetes secret containing the client secret |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}` | The container security context (https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container) |
| server.auth.audience | string | `"http://localhost:8080"` | Audience of provided by the identity provider |
| server.auth.dpopskew | string | `"1h"` | The amount of drift allowed between the server and the client for the DPoP Proof Token |
| server.auth.issuer | string | `"http://platform-keycloak/realms/opentdf"` | Identity provider issuer |
| server.auth.policy.csv | string | `nil` |  |
| server.auth.policy.extension | string | `"g, opentdf-admin, role:admin\ng, opentdf-standard, role:standard\n"` |  |
| server.auth.policy.groups_claim | string | `nil` |  |
| server.auth.policy.model | string | `nil` |  |
| server.auth.policy.username_claim | string | `nil` |  |
| server.auth.public_client_id | string | `"opentdf-public"` | The oidc client id, leveraged by otdfctl |
| server.auth.skew | string | `"1m"` | The amount of drift allowed between the server and the client for the Access Token |
| server.cors.allowcredentials | bool | `true` | Allow credentials |
| server.cors.allowedheaders | list | `["Accept","Authorization","Content-Type","X-CSRF-Token","X-Request-ID"]` | The allowed request headers |
| server.cors.allowedmethods | list | `["GET","POST","PUT","DELETE","OPTIONS"]` | The allowed request methods |
| server.cors.allowedorigins | list | `[]` | The allowed origins |
| server.cors.enabled | bool | `false` | Enable CORS (default: false) |
| server.cors.exposedheaders | list | `["Link"]` | List of response headers that browsers are allowed to access |
| server.cors.maxage | int | `3600` | The max age |
| server.cryptoProvider.standard.keys | list | `[{"alg":"rsa:2048","cert":"/etc/platform/kas/kas-cert.pem","kid":"r1","private":"/etc/platform/kas/kas-private.pem"},{"alg":"ec:secp256r1","cert":"/etc/platform/kas/kas-ec-cert.pem","kid":"e1","private":"/etc/platform/kas/kas-ec-private.pem"}]` | List of key pairs to load into the platform. (Currently only leveraged by KAS) |
| server.disableHealthChecks | bool | `false` | Disable Kubernetes Health Checks. (Useful for debugging) |
| server.enable_pprof | bool | `false` | Enable go pprof |
| server.grpc.maxCallRecvMsgSize | string | `nil` | This is the maximum size of a message that can be received by the server, ex. 4194304 (4MB), if unset use application's default (4MB). |
| server.grpc.maxCallSendMsgSize | string | `nil` | This is the maximum size of a message that can be sent by the server, ex. 4194304 (4MB), if unset use application's default (4MB). |
| server.grpc.reflectionEnabled | bool | `true` | Enables grpc reflection (https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) |
| server.http.idleTimeout | string | `nil` | The maximum amount of time to wait for the next request when keep-alives are enabled, ex. 30s, a negative value indicates no timeout, if unset or zero readTimeout is used. |
| server.http.maxHeaderBytes | string | `nil` | The maximum number of bytes the server will read parsing the request header's keys and values, including the request line, if unset or zero use application's default (1MB). |
| server.http.readHeaderTimeout | string | `nil` | The amount of time allowed to read request headers, ex. 30s, a negative value indicates no timeout, if unset or zero readTimeout is used. |
| server.http.readTimeout | string | `nil` | The maximum duration for reading the entire request including the body, ex. 30s, a negative value indicates no timeout, if unset or zero use application's default (5s). |
| server.http.writeTimeout | string | `nil` | The maximum duration before timing out writes of the response, ex. 30s, a negative value indicates no timeout, if unset or zero use application's default (10s). |
| server.port | int | `9000` | The server port |
| server.public_hostname | string | `""` | The client facing name for the KAS. This is baked into Key Access Objects, and required for key management with the policy service. |
| server.tls.additionalTrustedCerts | list | `[]` | Additional trusted certificates. These can be loaded following [projected volume](https://kubernetes.io/docs/concepts/storage/projected-volumes/) |
| server.tls.enabled | bool | `false` | Enables tls for platform server |
| server.tls.secret | string | `nil` | The server tls certificate. If not set, a self-signed certificate is generated |
| service.annotations | object | `{}` | Extra annotations to add to the service |
| service.appProtocol | string | `""` | explicit appProtocol for the service |
| service.port | int | `9000` | The port of the service |
| service.type | string | `"ClusterIP"` | The type of service to create |
| serviceAccount.annotations | object | `{}` | Extra annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| services.authorization | object | `{}` |  |
| services.entityresolution.clientid | string | `nil` | Client Id for Entity Resolver |
| services.entityresolution.clientsecret | string | `nil` | Client Secret for Entity Resolver |
| services.entityresolution.realm | string | `nil` | Entity Resolver Realm |
| services.entityresolution.subgroups | bool | `false` | Subgroups |
| services.entityresolution.url | string | `nil` | Identity Provider Entity Resolver |
| services.extraServices | object | `{}` | Additional services |
| services.kas.config | object | `{"keyring":[{"alg":"ec:secp256r1","kid":"e1"},{"alg":"rsa:2048","kid":"r1"}],"preview_features":{"ec_tdf_enabled":false,"key_management":false},"root_key":null}` | KAS service Configuration as yaml |
| services.kas.config.keyring | list | `[{"alg":"ec:secp256r1","kid":"e1"},{"alg":"rsa:2048","kid":"r1"}]` | Default keys for clients to use |
| services.kas.config.preview_features | object | `{"ec_tdf_enabled":false,"key_management":false}` | Preview feature enablement |
| services.kas.config.preview_features.ec_tdf_enabled | bool | `false` | Whether tdf based ecc support is enabled. |
| services.kas.config.preview_features.key_management | bool | `false` | Whether new key management features are enabled. |
| services.kas.privateKeysSecret | string | `"kas-private-keys"` | KAS secret containing keys @deprecated Use `private_keys_secret` instead. This value will be removed in a future release. |
| services.kas.private_keys_secret | string | `""` | KAS secret containing keys kas-private.pem , kas-cert.pem , kas-ec-private.pem , kas-ec-cert.pem |
| services.kas.root_key_secret | object | `{"key":"root_key","name":"kas-root-key"}` | Key needed when key_management feature is enabled (openssl rand 32 -hex) openssl rand 32 -hex | kubectl create secret generic kas-root-key --from-file=root_key=/dev/stdin |
| tolerations | list | `[]` | Tolerations to apply to the pod (https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) |
| trace.enabled | bool | `false` | Enable distributed tracing |
| trace.provider.file.compress | string | `nil` | Enable compression of trace files |
| trace.provider.file.maxAge | string | `nil` | Maximum age of files in days |
| trace.provider.file.maxBackups | string | `nil` | Maximum number of backup files |
| trace.provider.file.maxSize | string | `nil` | Maximum file size in MB |
| trace.provider.file.path | string | `nil` | Path to trace file output |
| trace.provider.file.prettyPrint | string | `nil` | Enable pretty-printed JSON |
| trace.provider.name | string | `nil` | Tracing provider (file or otlp) |
| trace.provider.otlp.endpoint | string | `nil` | Endpoint URL for the collector |
| trace.provider.otlp.headers | object | `{}` | Headers to include in OTLP requests |
| trace.provider.otlp.insecure | bool | `false` | Whether to use an insecure connection |
| trace.provider.otlp.protocol | string | `nil` | Protocol to use (grpc or http/grpc) |
| volumeMountTemplate | string | `"platform.volumeMountsEmpty.tpl"` | Add ability for downstream chart to merge additional volumeMounts |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumeTemplate | string | `"platform.volumesEmpty.tpl"` | Add ability for downstream chart to merge additional volumes |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |