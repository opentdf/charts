# Changelog

## [0.14.0](https://github.com/opentdf/charts/compare/platform-0.13.0...platform-0.14.0) (2025-08-13)


### Features

* **kas:** Add registered_kas_uri to values ([#148](https://github.com/opentdf/charts/issues/148)) ([0f637f3](https://github.com/opentdf/charts/commit/0f637f3c9553f5c51dcc2c314f9fc9f60a8a0756))

## [0.13.0](https://github.com/opentdf/charts/compare/platform-0.12.0...platform-0.13.0) (2025-08-11)


### Features

* Add init containers support to platform deployment ([#142](https://github.com/opentdf/charts/issues/142)) ([c7072f6](https://github.com/opentdf/charts/commit/c7072f626a5ffaba90c286fe2c1ea3f4b7b947bd))

## [0.12.0](https://github.com/opentdf/charts/compare/platform-0.11.0...platform-0.12.0) (2025-07-22)


### Features

* **deployments:** SRE-5532 Adding deployment annotations ([#139](https://github.com/opentdf/charts/issues/139)) ([1adffeb](https://github.com/opentdf/charts/commit/1adffebfa85f1f8915af443ac249d7948823839a))

## [0.11.0](https://github.com/opentdf/charts/compare/platform-0.10.1...platform-0.11.0) (2025-07-02)


### Features

* **platform:** Adds server.public_hostname ([#136](https://github.com/opentdf/charts/issues/136)) ([a778c2c](https://github.com/opentdf/charts/commit/a778c2c906600504a33b751d0628db6807902d4e))

## [0.10.1](https://github.com/opentdf/charts/compare/platform-0.10.0...platform-0.10.1) (2025-07-01)


### Bug Fixes

* Expose http server options ([#133](https://github.com/opentdf/charts/issues/133)) ([433a57b](https://github.com/opentdf/charts/commit/433a57b4c95d6171f38d054d2e831ef43735a989))

## [0.10.0](https://github.com/opentdf/charts/compare/platform-0.9.0...platform-0.10.0) (2025-06-25)


### Features

* **platform:** Expose grpc server options in platform chart ([#130](https://github.com/opentdf/charts/issues/130)) ([d2ee0c5](https://github.com/opentdf/charts/commit/d2ee0c5a9596720387c36cf8dd12640bea9857a7))

## [0.9.0](https://github.com/opentdf/charts/compare/platform-0.8.5...platform-0.9.0) (2025-06-24)


### Features

* updated platform configuration options ([#127](https://github.com/opentdf/charts/issues/127)) ([03809ae](https://github.com/opentdf/charts/commit/03809ae2a7e4ae04477ec9690c12539bca9e9942))

## [0.8.5](https://github.com/opentdf/charts/compare/platform-0.8.4...platform-0.8.5) (2025-04-14)


### Bug Fixes

* update sdk_config to match platform configuration ([#117](https://github.com/opentdf/charts/issues/117)) ([c0f65fb](https://github.com/opentdf/charts/commit/c0f65fb1a0e86d77cf3803a80f8a98ca6fd4d83e))

## [0.8.4](https://github.com/opentdf/charts/compare/platform-0.8.3...platform-0.8.4) (2025-01-29)


### Bug Fixes

* add override for appprotocol ([#110](https://github.com/opentdf/charts/issues/110)) ([03f08bf](https://github.com/opentdf/charts/commit/03f08bf8d7199389db89e5700118ac5b32d70f65))

## [0.8.3](https://github.com/opentdf/charts/compare/platform-0.8.2...platform-0.8.3) (2024-12-06)


### Bug Fixes

* auth policy to set backward compatible group to role mapppings ([#106](https://github.com/opentdf/charts/issues/106)) ([172b0ec](https://github.com/opentdf/charts/commit/172b0ec983d58ef5654a8dcea111e0a4ed24aa67))

## [0.8.2](https://github.com/opentdf/charts/compare/platform-0.8.1...platform-v0.8.2) (2024-12-04)


### Bug Fixes

* update postgres chart version 16.2.5 ([#95](https://github.com/opentdf/charts/issues/95)) ([768c407](https://github.com/opentdf/charts/commit/768c4072f215733ef57bb68e2ae617666a9cfeaf))

## [0.8.1](https://github.com/opentdf/charts/compare/platform-0.8.0...platform-v0.8.1) (2024-12-04)


### Bug Fixes

* set appVersion to v0.4.32 ([#92](https://github.com/opentdf/charts/issues/92)) ([f94d5dc](https://github.com/opentdf/charts/commit/f94d5dc64353c14821537ed9761f7e08a2500aeb))

## [0.8.0](https://github.com/opentdf/charts/compare/platform-0.7.1...platform-v0.8.0) (2024-10-17)


### Features

* Add pod disruption budget capability ([#87](https://github.com/opentdf/charts/issues/87)) ([549a779](https://github.com/opentdf/charts/commit/549a77962f7913e75cfcffb82a9a2d10c8970695))

## [0.7.1](https://github.com/opentdf/charts/compare/platform-0.7.0...platform-v0.7.1) (2024-10-16)


### Bug Fixes

* **ci:** Add public_client_id to server.auth config ([#89](https://github.com/opentdf/charts/issues/89)) ([6829c87](https://github.com/opentdf/charts/commit/6829c8783566a955bb76287bdd640a7b2507ebc8))
* **core:** deprecate keycloak-from-config ([#81](https://github.com/opentdf/charts/issues/81)) ([da5a685](https://github.com/opentdf/charts/commit/da5a6856a0b14fcc6eab5810d4cf5a9518d68abe)), closes [#80](https://github.com/opentdf/charts/issues/80)
* port naming/app protocol tls.enabled ([#90](https://github.com/opentdf/charts/issues/90)) ([613f67e](https://github.com/opentdf/charts/commit/613f67e860972dc8d27bf0fc6fb55daa89bf8b7a))

## [0.7.0](https://github.com/opentdf/charts/compare/platform-0.6.2...platform-v0.7.0) (2024-08-20)


### Features

* add ability to merge custom volumes and volumeMount templates ([0666d4f](https://github.com/opentdf/charts/commit/0666d4f8e250cf1b002c77beda846aed51e9b54e))
* platform configuration template for child charts ([#76](https://github.com/opentdf/charts/issues/76)) ([02f8839](https://github.com/opentdf/charts/commit/02f8839076ba7c2eb48e049af160488bb5333ce0))
* platform mode and global sdk configuration support ([#73](https://github.com/opentdf/charts/issues/73)) ([0666d4f](https://github.com/opentdf/charts/commit/0666d4f8e250cf1b002c77beda846aed51e9b54e))


### Bug Fixes

* add auth skew configuration ([0666d4f](https://github.com/opentdf/charts/commit/0666d4f8e250cf1b002c77beda846aed51e9b54e))
* add optional required field for db to drive db configuration ([#78](https://github.com/opentdf/charts/issues/78)) ([81f64ca](https://github.com/opentdf/charts/commit/81f64ca04c6b322a3b2cb3132487c1969a3f0a29))
* correct h2c service appProtocol ([#77](https://github.com/opentdf/charts/issues/77)) ([25a1ae8](https://github.com/opentdf/charts/commit/25a1ae862937707fb42a28d037b97849a575e914))
* support openshift compatibility for service appProtocol ([81f64ca](https://github.com/opentdf/charts/commit/81f64ca04c6b322a3b2cb3132487c1969a3f0a29))
* trusted-cert volume and mount defined when not expected ([#75](https://github.com/opentdf/charts/issues/75)) ([73cde0e](https://github.com/opentdf/charts/commit/73cde0e2dbecb51ff426c396f2daf7ef071f4834))
* upgrade keycloak chart to 22.1.1 ([0666d4f](https://github.com/opentdf/charts/commit/0666d4f8e250cf1b002c77beda846aed51e9b54e))
* upgrade postgresql chart to 15.5.21 ([0666d4f](https://github.com/opentdf/charts/commit/0666d4f8e250cf1b002c77beda846aed51e9b54e))

## [0.6.2](https://github.com/opentdf/charts/compare/platform-0.6.1...platform-v0.6.2) (2024-07-25)


### Bug Fixes

* only mount keycloak cert if ingress and tls enabled ([5cacbb8](https://github.com/opentdf/charts/commit/5cacbb814060f531ea3cccbea2375f41e33c4d20))
* **platform:** extra certs overwriting files in /etc/ssl/certs ([#70](https://github.com/opentdf/charts/issues/70)) ([5cacbb8](https://github.com/opentdf/charts/commit/5cacbb814060f531ea3cccbea2375f41e33c4d20))
* remove defaultMode on trusted-certs volume ([5cacbb8](https://github.com/opentdf/charts/commit/5cacbb814060f531ea3cccbea2375f41e33c4d20))
* set default security context values for container ([5cacbb8](https://github.com/opentdf/charts/commit/5cacbb814060f531ea3cccbea2375f41e33c4d20))
* set pod security context defaults ([5cacbb8](https://github.com/opentdf/charts/commit/5cacbb814060f531ea3cccbea2375f41e33c4d20))

## [0.6.1](https://github.com/opentdf/charts/compare/platform-0.6.0...platform-v0.6.1) (2024-07-16)


### Bug Fixes

* allow enablement of pprof in service ([#68](https://github.com/opentdf/charts/issues/68)) ([060f9f6](https://github.com/opentdf/charts/commit/060f9f65c5de91e8eac2f5bf9ef1b95aee8c72d1))

## [0.6.0](https://github.com/opentdf/charts/compare/platform-0.5.0...platform-v0.6.0) (2024-07-15)


### Features

* **common:** add templating for hostaliases ([#66](https://github.com/opentdf/charts/issues/66)) ([efc773c](https://github.com/opentdf/charts/commit/efc773cb38f6a5db6f6fff083245dd2aafa0a3cf))

## [0.5.0](https://github.com/opentdf/charts/compare/platform-0.4.1...platform-v0.5.0) (2024-07-10)


### ⚠ BREAKING CHANGES

* yaml config for kas ([#52](https://github.com/opentdf/charts/issues/52))

### Features

* ability to add extra service definitions ([#62](https://github.com/opentdf/charts/issues/62)) ([e35695e](https://github.com/opentdf/charts/commit/e35695e2137446faf5a07cb5312d121a9f389ab2))
* ability to load envFrom a configmap or secret ([e35695e](https://github.com/opentdf/charts/commit/e35695e2137446faf5a07cb5312d121a9f389ab2))
* ability to set extra environment variables ([e35695e](https://github.com/opentdf/charts/commit/e35695e2137446faf5a07cb5312d121a9f389ab2))
* add cors configuration support ([#60](https://github.com/opentdf/charts/issues/60)) ([f175e88](https://github.com/opentdf/charts/commit/f175e88b04e7846e343ec0d460c2cf365af47dcd))
* **platform:** add headless service ([#55](https://github.com/opentdf/charts/issues/55)) ([c7c8611](https://github.com/opentdf/charts/commit/c7c8611168fb4185bce5015d3c3d089688fff57a))
* **platform:** additional certs projected volumes ([#57](https://github.com/opentdf/charts/issues/57)) ([d1205b4](https://github.com/opentdf/charts/commit/d1205b4b77602587fd28c3ce69734ce49a8c7152))
* **platform:** propagate playground setting to platform ([#56](https://github.com/opentdf/charts/issues/56)) ([2aaa497](https://github.com/opentdf/charts/commit/2aaa49749bc24df9ee29fcf1f3e729c8a47e3e2a))
* yaml config for kas ([#52](https://github.com/opentdf/charts/issues/52)) ([e5b4c1a](https://github.com/opentdf/charts/commit/e5b4c1a1e7320d2738ba04693c9927d225a5d64c))


### Bug Fixes

* enable extra keycloak features needed by platform ([#63](https://github.com/opentdf/charts/issues/63)) ([d877541](https://github.com/opentdf/charts/commit/d8775412be5dc967477a29ea191c972a3a035a4d))
* **platform:** update crypto provider and kas keyring configuration ([d1205b4](https://github.com/opentdf/charts/commit/d1205b4b77602587fd28c3ce69734ce49a8c7152))
* sync service configuration with platform ([#64](https://github.com/opentdf/charts/issues/64)) ([63dc6ba](https://github.com/opentdf/charts/commit/63dc6ba38ea283e6a7a1881bfe5a4e224c8e6206))
* unquote cors config values ([#61](https://github.com/opentdf/charts/issues/61)) ([1bb67ff](https://github.com/opentdf/charts/commit/1bb67ffc0118de644452a0fe7d89ee2bcdb58f1f))

## [0.4.1](https://github.com/opentdf/charts/compare/platform-0.4.0...platform-v0.4.1) (2024-06-06)


### Bug Fixes

* **platform:** update kas key paths to new config format ([#48](https://github.com/opentdf/charts/issues/48)) ([9c4bd5b](https://github.com/opentdf/charts/commit/9c4bd5bf64dfe8114e2aff03fbaa8aff70ecb09d))

## [0.4.0](https://github.com/opentdf/charts/compare/platform-0.3.0...platform-v0.4.0) (2024-05-20)


### Miscellaneous Chores

* release 0.4.0 ([#46](https://github.com/opentdf/charts/issues/46)) ([97cd502](https://github.com/opentdf/charts/commit/97cd50274a2047e68ed67ebefa775f5b657d4b5b))

## [0.3.0](https://github.com/opentdf/charts/compare/platform-0.2.6...platform-v0.3.0) (2024-05-15)


### Features

* add entity resolution service to values ([#41](https://github.com/opentdf/charts/issues/41)) ([3eb303d](https://github.com/opentdf/charts/commit/3eb303d2a9e372ac835679e56b393f5c8ff0bf6b))

## [0.2.6](https://github.com/opentdf/charts/compare/platform-0.2.5...platform-v0.2.6) (2024-05-15)


### Bug Fixes

* change file paths to be more generic ([#42](https://github.com/opentdf/charts/issues/42)) ([acf9487](https://github.com/opentdf/charts/commit/acf94873ca44daaf3b94917ef0a9babc65e45418))

## [0.2.5](https://github.com/opentdf/charts/compare/platform-0.2.4...platform-v0.2.5) (2024-05-14)


### Bug Fixes

* remove container command ([#39](https://github.com/opentdf/charts/issues/39)) ([0af92ca](https://github.com/opentdf/charts/commit/0af92cab269b7770092c1b9f1e68e629d2cb2d48))

## [0.2.4](https://github.com/opentdf/charts/compare/platform-0.2.3...platform-v0.2.4) (2024-05-14)


### Bug Fixes

* set appVersion to nightly ([#37](https://github.com/opentdf/charts/issues/37)) ([54ec9b1](https://github.com/opentdf/charts/commit/54ec9b1faf4c3f4627b8cb94381dde2fa41df245))

## [0.2.3](https://github.com/opentdf/charts/compare/platform-0.2.2...platform-v0.2.3) (2024-05-07)


### Bug Fixes

* changes service to http2 from http ([#32](https://github.com/opentdf/charts/issues/32)) ([a548707](https://github.com/opentdf/charts/commit/a548707c04ef2a62803713bb3d8d3e810f37b690))

## [0.2.2](https://github.com/opentdf/charts/compare/platform-0.2.1...platform-v0.2.2) (2024-05-02)


### Bug Fixes

* platform optional health check ([#28](https://github.com/opentdf/charts/issues/28)) ([4a977d1](https://github.com/opentdf/charts/commit/4a977d1577b02056e63e68809c8580dc57a01a89))

## [0.2.1](https://github.com/opentdf/charts/compare/platform-0.2.0...platform-v0.2.1) (2024-04-26)


### Bug Fixes

* platform chart readme ([#22](https://github.com/opentdf/charts/issues/22)) ([fc0b23c](https://github.com/opentdf/charts/commit/fc0b23c0fd824900f6f96e39a87cf90e1e29f961))

## [0.2.0](https://github.com/opentdf/charts/compare/platform-v0.1.0...platform-v0.2.0) (2024-04-26)


### Features

* initial platform helm chart setup ([#1](https://github.com/opentdf/charts/issues/1)) ([a3eba3f](https://github.com/opentdf/charts/commit/a3eba3fa508a7d0c7b39aa0845c997bbefa3b608))


### Bug Fixes

* change version ([#20](https://github.com/opentdf/charts/issues/20)) ([5a22524](https://github.com/opentdf/charts/commit/5a225244624e577a3842566bc75abd6ec2034a48))
