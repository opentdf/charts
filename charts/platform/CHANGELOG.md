# Changelog

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
