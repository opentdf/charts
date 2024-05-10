# platform

![Version: 0.2.3](https://img.shields.io/badge/Version-0.2.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: nightly](https://img.shields.io/badge/AppVersion-nightly-informational?style=flat-square)

A Helm Chart for OpenTDF Platform

**_NOTE:_** Until a stable platform release is available, the current appVersion is set to `nightly`.

## Requirements

In order to run the platform you need a supported identity provider and postgres database. We have included the following charts as dependencies:

- [Keycloak](https://www.keycloak.org/guides)
  - [Helm Chart](https://artifacthub.io/packages/helm/bitnami/keycloak)

- [Postgres](https://www.postgresql.org/)
  - [Helm Chart](https://artifacthub.io/packages/helm/bitnami/postgresql)

If you want to deploy keycloak and postgresql as part of the platform deployment make sure to follow [Playground Configuration](#playground-configuration-optional) section.

## Pre-Requisites

### TLS Certificate

The chart will attempt to generate a TLS certificate if `tls.enabled` is set to `true` and no `tls.secretName` is provided.

Alternatively, you can also provide your own certificate as well. Once you have the certificate and key, you can create a TLS secret with the following command:

```bash
kubectl create secret tls <tls-secret-name> --cert <path-to-cert-file> --key <path-to-key-file>
```

### KAS Secret

In order to run the KAS server, you need to provide the KAS with a set a keys. This will allow the KAS to support nanotdf and tdf3 rewrap functionality.

#### Generate a set of KAS keys

```bash
openssl req -x509 -nodes -newkey RSA:2048 -subj "/CN=kas" -keyout kas-private.pem -out kas-cert.pem -days 365
openssl ecparam -name prime256v1 >ecparams.tmp
openssl req -x509 -nodes -newkey ec:ecparams.tmp -subj "/CN=kas" -keyout kas-ec-private.pem -out kas-ec-cert.pem -days 365
```

To create a secret with the keys, run the following command:

```bash
kubectl create secret generic kas-private-keys --from-file=./kas-cert.pem --from-file=./kas-ec-cert.pem --from-file=./kas-ec-private.pem --from-file=kas-private.pem
```

### HTTP/2 Support

The platform provides both gRPC and RESTful interfaces. Since gRPC requires HTTP/2 support, it's essential to configure your ingress with the appropriate annotations to handle this requirement.

Below are starting examples for configuring various ingress providers:

#### Openshift

Check to see if you have HTTP/2 enabled on your ingress routers
- https://docs.openshift.com/container-platform/4.15/networking/ingress-operator.html#nw-http2-haproxy_configuring-ingress

If no HTTP/2 support is enabled. You can still leverage gRPC by using a passthrough route.

Example values.yaml for passthrough route:
```yaml
ingress:
  enabled: true
  annotations:
    route.openshift.io/termination: "passthrough"
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

Create a Keycloak Configuration File with the following content:

```yaml
baseUrl: &baseUrl http://localhost:8888
serverBaseUrl: &serverBaseUrl http://localhost:8080
customAudMapper: &customAudMapper
  name: audience-mapper
  protocol: openid-connect
  protocolMapper: oidc-audience-mapper
  config:
    included.custom.audience: *serverBaseUrl
    access.token.claim: "true"
    id.token.claim: "true"
realms:
  - realm_repepresentation:
      realm: opentdf
      enabled: true
    custom_realm_roles:
      - name: opentdf-org-admin
      - name: opentdf-admin
      - name: opentdf-readonly
    custom_client_roles:
      tdf-entity-resolution:
        - name: entity-resolution-test-role
    custom_groups:
      - name: mygroup
        attributes:
          mygroupattribute:
            - mygroupvalue
    clients:
      - client:
          clientID: opentdf
          enabled: true
          name: opentdf
          serviceAccountsEnabled: true
          clientAuthenticatorType: client-secret
          secret: secret
          protocolMappers:
            - *customAudMapper
        sa_realm_roles:
          - opentdf-org-admin
      - client:
          clientID: opentdf-sdk
          enabled: true
          name: opentdf-sdk
          serviceAccountsEnabled: true
          clientAuthenticatorType: client-secret
          secret: secret
          protocolMappers:
            - *customAudMapper
        sa_realm_roles:
          - opentdf-readonly
      - client:
          clientID: tdf-entity-resolution
          enabled: true
          name: tdf-entity-resolution
          serviceAccountsEnabled: true
          clientAuthenticatorType: client-secret
          secret: secret
          protocolMappers:
            - *customAudMapper
        sa_client_roles:
          realm-management:
            - view-clients
            - query-clients
            - view-users
            - query-users
      - client:
          clientID: tdf-authorization-svc
          enabled: true
          name: tdf-authorization-svc
          serviceAccountsEnabled: true
          clientAuthenticatorType: client-secret
          secret: secret
          protocolMappers:
            - *customAudMapper
    users:
      - username: sample-user
        enabled: true
        firstName: sample
        lastName: user
        email: sampleuser@sample.com
        credentials:
          - value: testuser123
            type: password
        attributes:
          superhero_name:
            - thor
          superhero_group:
            - avengers
        groups:
          - mygroup
        realmRoles:
          - opentdf-org-admin
        clientRoles:
          realm-management:
            - view-clients
            - query-clients
            - view-users
            - query-users
          tdf-entity-resolution:
            - entity-resolution-test-role
    token_exchanges:
      - start_client: opentdf
        target_client: opentdf-sdk
 ```

 Run the following command to provision keycloak test data

  ```sh
    docker run \                                                                                                                                       ░▒▓ 127 ✘  system Node  at 13:22:53
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
| oci://registry-1.docker.io/bitnamicharts | keycloak | 21.0.3 |
| oci://registry-1.docker.io/bitnamicharts | postgresql | 15.2.4 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment |
| autoscaling.enabled | bool | `false` | Enable horizontal autoscaling |
| autoscaling.maxReplicas | int | `100` | Maximum number of pods to run |
| autoscaling.minReplicas | int | `1` | Minimum number of pods to run |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization percentage |
| db.database | string | `"opentdf"` | The database name |
| db.host | string | `"platform-db"` | The database host |
| db.password.secret.key | string | `"password"` | The key in the secret containing the database user password |
| db.password.secret.name | string | `"opentdf-db-credentials"` | The kubernetes secret containing the database user password |
| db.port | int | `5432` | The database port |
| db.sslmode | string | `"prefer"` | The database ssl mode ( disable, prefer, require, verify-ca, verify-full ) |
| db.user | string | `"postgres"` | The database user |
| fullnameOverride | string | `""` | Overrides the generated fullname |
| image.pullPolicy | string | `"IfNotPresent"` | The image tag to pull policy |
| image.repository | string | `"registry.opentdf.io/platform"` | The image repository to pull from |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | The image pull secrets to use |
| ingress.annotations | object | `{}` | Extra Ingress annotations |
| ingress.className | string | `""` | Ingress class name if required |
| ingress.enabled | bool | `false` | Enable Ingress |
| ingress.hosts | list | `[{"host":"chart-example.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]` | List of hosts for the ingress |
| ingress.tls | list | `[]` | List of tls hosts |
| keycloak.auth.adminPassword | string | `"keycloak123"` |  |
| keycloak.auth.adminUser | string | `"admin"` |  |
| keycloak.externalDatabase.database | string | `"opentdf"` |  |
| keycloak.externalDatabase.existingSecret | string | `"opentdf-db-credentials"` |  |
| keycloak.externalDatabase.existingSecretPasswordKey | string | `"password"` |  |
| keycloak.externalDatabase.host | string | `"platform-db"` |  |
| keycloak.externalDatabase.port | int | `5432` |  |
| keycloak.externalDatabase.user | string | `"postgres"` |  |
| keycloak.fullnameOverride | string | `"platform-keycloak"` |  |
| keycloak.ingress.enabled | bool | `true` |  |
| keycloak.ingress.labels."opentdf.io/keycloak-admin" | string | `"false"` |  |
| keycloak.ingress.selfSigned | bool | `true` |  |
| keycloak.ingress.tls | bool | `true` |  |
| keycloak.keycloakConfigCli.configuration."opentdf.json" | string | `"{\n  \"realm\":\"opentdf\",\n  \"enabled\": true,\n  \"clients\": []\n}\n"` |  |
| keycloak.keycloakConfigCli.enabled | bool | `true` |  |
| keycloak.postgresql.enabled | bool | `false` |  |
| keycloak.proxy | string | `"edge"` |  |
| keycloak.tls.autoGenerated | bool | `true` |  |
| keycloak.tls.enabled | bool | `true` |  |
| logger.level | string | `"info"` | The platform log level ( debug, info, warn, error ) |
| logger.output | string | `"stdout"` | The platform log output |
| logger.type | string | `"json"` | The platform log format ( json, text ) |
| nameOverride | string | `""` | Overrides the chart name |
| nodeSelector | object | `{}` | Target specific nodes in the cluster |
| opa.embedded | bool | `true` | Whether to run an embedded OPA bundle. (!!! Do not enable this in production !!!) |
| playground | bool | `false` |  |
| podAnnotations | object | `{}` | Extra annotations to add to the pod |
| podLabels | object | `{}` | Extra labels to add to the pod |
| podSecurityContext | object | `{}` | The pod security context |
| postgresql.auth.database | string | `"opentdf"` |  |
| postgresql.auth.postgresPassword | string | `"opentdf123"` |  |
| postgresql.auth.usePasswordFiles | bool | `true` |  |
| postgresql.fullnameOverride | string | `"platform-db"` |  |
| postgresql.tls.autoGenerated | bool | `true` |  |
| postgresql.tls.enabled | bool | `true` |  |
| replicaCount | int | `1` | The number of Platform pods to run |
| resources | object | `{}` | Resources to allocate to the container |
| securityContext | object | `{}` | The container security context |
| server.auth.audience | string | `"http://localhost:8080"` | Audience of provided by the identity provider |
| server.auth.issuer | string | `"http://platform-keycloak/realms/opentdf"` | Identity provider issuer |
| server.auth.policy.claim | string | `nil` |  |
| server.auth.policy.csv | string | `nil` |  |
| server.auth.policy.default | string | `nil` |  |
| server.auth.policy.map | string | `nil` |  |
| server.cryptoProvider.standard.ec.key1.privateKeyPath | string | `"/etc/opentdf/kas/kas-ec-private.pem"` |  |
| server.cryptoProvider.standard.ec.key1.publicKeyPath | string | `"/etc/opentdf/kas/kas-ec-cert.pem"` |  |
| server.cryptoProvider.standard.rsa.key1.privateKeyPath | string | `"/etc/opentdf/kas/kas-private.pem"` |  |
| server.cryptoProvider.standard.rsa.key1.publicKeyPath | string | `"/etc/opentdf/kas/kas-cert.pem"` |  |
| server.grpc.reflectionEnabled | bool | `true` | Enables grpc reflection (https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) |
| server.healthCheckz | bool | `true` | Enables Kubernetes Health Checkz |
| server.port | int | `9000` | The server port |
| server.tls.enabled | bool | `true` | Enables tls |
| server.tls.secret | string | `nil` | The server tls certificate. If not set, a self-signed certificate is generated |
| service.port | int | `9000` | The port of the service |
| service.type | string | `"ClusterIP"` | The type of service to create |
| serviceAccount.annotations | object | `{}` | Extra annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| services.authorization.client | string | `"tdf-entity-resolution"` | Client id for the external entity store |
| services.authorization.enabled | bool | `true` | Authorization service enabled |
| services.authorization.legacy | bool | `true` | Keycloak legacy mode?? |
| services.authorization.realm | string | `"opentdf"` | Keycloak Realm |
| services.authorization.secret | string | `nil` | Client secret for the external entity store |
| services.authorization.url | string | `nil` | External entity store (currently only keycloak is supported) |
| services.kas.enabled | bool | `true` | KAS service enabled |
| services.kas.privateKeysSecret | string | `"kas-private-keys"` | KAS secret containing keys kas-private.pem , kas-cert.pem , kas-ec-private.pem , kas-ec-cert.pem |
| tolerations | list | `[]` | Tolerations to apply to the pod (https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |