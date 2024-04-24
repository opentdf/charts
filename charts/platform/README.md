# OpenTDF Platform

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.0](https://img.shields.io/badge/AppVersion-0.2.0-informational?style=flat-square)

A Helm Chart for OpenTDF Platform

## Pre-Requisites

### Creating TLS Certificate

The chart will attempt to generate a TLS certificate if `tls.enabled` is set to `true` and no `tls.secretName` is provided.

Alternatively, you can also provide your own certificate as well. Once you have the certificate and key, you can create a TLS secret with the following command:

```bash
kubectl create secret tls <tls-secret-name> --cert <path-to-cert-file> --key <path-to-key-file>
```

### Creating KAS Secret

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

The platform exposes a gRPC and Restful interface. gRPC requires HTTP/2 support so you will need to ensure providing the proper ingress annotations.
Below is the start of examples for different ingress providers.

#### Openshift

Check to see if you have HTTP/2 enabled on your ingress routers
- https://docs.openshift.com/container-platform/4.15/networking/ingress-operator.html#nw-http2-haproxy_configuring-ingress

TBD

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment |
| autoscaling.enabled | bool | `false` | Enable horizontal autoscaling |
| autoscaling.maxReplicas | int | `100` | Maximum number of pods to run |
| autoscaling.minReplicas | int | `1` | Minimum number of pods to run |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization percentage |
| db.host | string | `nil` | The database host |
| db.name | string | `"opentdf"` | The database name |
| db.password.secret.key | string | `"password"` | The key in the secret containing the database user password |
| db.password.secret.name | string | `"opentdf-db-credentials"` | The kubernetes secret containing the database user password |
| db.port | int | `5432` | The database port |
| db.sslmode | string | `"prefer"` | The database ssl mode ( disable, prefer, require, verify-ca, verify-full ) |
| db.user | string | `"opentdf"` | The database user |
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
| logger.level | string | `"info"` | The platform log level ( debug, info, warn, error ) |
| logger.output | string | `"stdout"` | The platform log output |
| logger.type | string | `"json"` | The platform log format ( json, text ) |
| nameOverride | string | `""` | Overrides the chart name |
| nodeSelector | object | `{}` | Target specific nodes in the cluster |
| opa.embedded | bool | `true` | Whether to run an embedded OPA bundle. (!!! Do not enable this in production !!!) |
| podAnnotations | object | `{}` | Extra annotations to add to the pod |
| podLabels | object | `{}` | Extra labels to add to the pod |
| podSecurityContext | object | `{}` | The pod security context |
| replicaCount | int | `1` | The number of Platform pods to run |
| resources | object | `{}` | Resources to allocate to the container |
| securityContext | object | `{}` | The container security context |
| server.auth.audience | string | `nil` | Audience of provided by the identity provider |
| server.auth.issuer | string | `nil` | Identity provider issuer |
| server.auth.policy.claim | string | `nil` |  |
| server.auth.policy.csv | string | `nil` |  |
| server.auth.policy.default | string | `nil` |  |
| server.auth.policy.map | string | `nil` |  |
| server.cryptoProvider.standard.ec.key1.privateKeyPath | string | `"/etc/opentdf/kas/kas-ec-private.pem"` |  |
| server.cryptoProvider.standard.ec.key1.publicKeyPath | string | `"/etc/opentdf/kas/kas-ec-cert.pem"` |  |
| server.cryptoProvider.standard.rsa.key1.privateKeyPath | string | `"/etc/opentdf/kas/kas-private.pem"` |  |
| server.cryptoProvider.standard.rsa.key1.publicKeyPath | string | `"/etc/opentdf/kas/kas-cert.pem"` |  |
| server.grpc.reflectionEnabled | bool | `true` | Enables grpc reflection (https://github.com/grpc/grpc/blob/master/doc/server-reflection.md) |
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