# platform

![Version: 0.7.0](https://img.shields.io/badge/Version-0.7.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: nightly](https://img.shields.io/badge/AppVersion-nightly-informational?style=flat-square)

A Helm Chart for OpenTDF Platform

> [!WARNING]
> This chart is currently under active development and breaking changes may occur.

> [!NOTE]
> Until a stable platform release is available, the current appVersion is set to `nightly`.

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
         -it registry.opentdf.io/platform:nightly provision keycloak-from-config -f /home/nonroot/.opentdf/kc-config.yaml -e https://kc-nightly.apps.dsp.shp.virtru.us -u admin -p keycloak123
  ```

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Opentdf |  |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| oci://registry-1.docker.io/bitnamicharts | keycloak | 22.1.1 |
| oci://registry-1.docker.io/bitnamicharts | postgresql | 15.5.21 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment |
| autoscaling.enabled | bool | `false` | Enable horizontal autoscaling |
| autoscaling.maxReplicas | int | `100` | Maximum number of pods to run |
| autoscaling.minReplicas | int | `1` | Minimum number of pods to run |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization percentage |
| configTemplate | string | `"platform.configurationEmpty.tpl"` |  |
| db.database | string | `"opentdf"` | The database name |
| db.host | string | `"platform-db"` | The database host |
| db.password.secret.key | string | `"password"` | The key in the secret containing the database user password |
| db.password.secret.name | string | `"opentdf-db-credentials"` | The kubernetes secret containing the database user password |
| db.port | int | `5432` | The database port |
| db.required | bool | `true` | If database connection info is required for the service (default: true) |
| db.sslmode | string | `"prefer"` | The database ssl mode ( disable, prefer, require, verify-ca, verify-full ) |
| db.user | string | `"opentdf"` | The database user |
| envFrom | list | `[]` | Environment variables from a configmap or secret  |
| extraEnv | list | `[]` | Extra environment variables to add to the container |
| fullnameOverride | string | `""` | Overrides the generated fullname |
| hostAliases | list | `[]` | Host Alias entries to add to pods |
| image.pullPolicy | string | `"IfNotPresent"` | The image tag to pull policy |
| image.repository | string | `"registry.opentdf.io/platform"` | The image repository to pull from |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | The image pull secrets to use |
| ingress.annotations | object | `{}` | Extra Ingress annotations |
| ingress.className | string | `"gce-internal"` | Ingress class name if required |
| ingress.defaultBackend.enabled | bool | `true` |  |
| ingress.enabled | bool | `true` | Enable Ingress |
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
| sdk_config.endpoint | string | `""` | The core platform endpoint |
| sdk_config.existingSecret | object | `{"key":"","name":""}` | Oauth2 Client Secret Kubernetes Secret |
| sdk_config.existingSecret.key | string | `""` | The key in the secret containing the client secret |
| sdk_config.existingSecret.name | string | `""` | The kubernetes secret containing the client secret |
| sdk_config.plaintext | bool | `false` | Plaintext Insecure Connection |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}` | The container security context (https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container) |
| server.auth.audience | string | `"http://localhost:8080"` | Audience of provided by the identity provider |
| server.auth.dpopskew | string | `"1h"` | The amount of drift allowed between the server and the client for the DPoP Proof Token |
| server.auth.issuer | string | `"http://platform-keycloak/realms/opentdf"` | Identity provider issuer |
| server.auth.policy.claim | string | `nil` |  |
| server.auth.policy.csv | string | `nil` |  |
| server.auth.policy.default | string | `nil` |  |
| server.auth.policy.map | string | `nil` |  |
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
| server.grpc.reflectionEnabled | bool | `true` | Enables grpc reflection (https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) |
| server.port | int | `9000` | The server port |
| server.tls.additionalTrustedCerts | list | `[]` | Additional trusted certificates. These can be loaded following [projected volume](https://kubernetes.io/docs/concepts/storage/projected-volumes/) |
| server.tls.enabled | bool | `false` | Enables tls for platform server |
| server.tls.secret | string | `nil` | The server tls certificate. If not set, a self-signed certificate is generated |
| service.annotations | object | `{}` | Extra annotations to add to the service |
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
| services.entityresolution.subgroups | bool | `false` | Subgroups  |
| services.entityresolution.url | string | `nil` | Identity Provider Entity Resolver |
| services.extraServices | object | `{}` | Additional services |
| services.kas.config | object | `{"keyring":[{"alg":"ec:secp256r1","kid":"e1"},{"alg":"rsa:2048","kid":"r1"}]}` | KAS service Configuration as yaml |
| services.kas.config.keyring | list | `[{"alg":"ec:secp256r1","kid":"e1"},{"alg":"rsa:2048","kid":"r1"}]` | Default keys for clients to use |
| services.kas.privateKeysSecret | string | `"kas-private-keys"` | KAS secret containing keys kas-private.pem , kas-cert.pem , kas-ec-private.pem , kas-ec-cert.pem |
| tolerations | list | `[]` | Tolerations to apply to the pod (https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) |
| volumeMountTemplate | string | `"platform.volumeMountsEmpty.tpl"` | Add ability for downstream chart to merge additional volumeMounts |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumeTemplate | string | `"platform.volumesEmpty.tpl"` | Add ability for downstream chart to merge additional volumes |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |