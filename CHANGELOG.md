# Changelog

All notable customer-visible changes are recorded here.

## Unreleased

### Added

- Typed email/SMS passwordless start, verification and approved custom completion
  for public and confidential clients, with explicit authorization/hosted/custom
  outcomes generated from the canonical Identity contract.
- Secretless public-client authorization-code exchange, refresh rotation and
  token revocation, plus secure S256 PKCE and strict callback/state helpers.
- Identity operation metadata and payload generation, public package consumers,
  and browser compilation coverage for the new flow.

### Security

- Authentication mutations never retry automatically or follow HTTP redirects.
  Sensitive request/response details, transport causes and untrusted error
  metadata are omitted from authentication errors; retry timing is retained.

Release preparation follows the existing release-please prerelease workflow.
Existing versions and tags remain immutable. Publication and hosted email/SMS
verification require separate release/environment checks.

## [0.1.0-beta.1.2](https://github.com/aetherplatform/aether-go/compare/aether-go-v0.1.0-beta.1...aether-go-v0.1.0-beta.1.2) (2026-09-13)


### Features

* add Identity Events Notifications and Webhooks ([b95e201](https://github.com/aetherplatform/aether-go/commit/b95e2012f0658e48583a0704192ddfd3a1887542))
* **identity:** prepare public passwordless OAuth beta ([51936f2](https://github.com/aetherplatform/aether-go/commit/51936f22f411587f1ce59ec8f7ffbea5d5ecc142))
* **identity:** prepare public passwordless OAuth beta ([bc8b2c2](https://github.com/aetherplatform/aether-go/commit/bc8b2c2c607a4a53ec1636117e19919bdf38156f))
* publish Aether Go SDK beta source ([b39d25d](https://github.com/aetherplatform/aether-go/commit/b39d25d86b4e7512449bae65aced5f13348a9488))


### Bug Fixes

* preserve standalone export tooling ([882b8d3](https://github.com/aetherplatform/aether-go/commit/882b8d34964ef111d3966c6a1952e3107d56b711))
* **release:** configure identity before creating the beta tag ([d3c26a5](https://github.com/aetherplatform/aether-go/commit/d3c26a5896f2d799fe05e0cd045f7eaf0d8c257b))
* **release:** set Git identity for the first Go beta tag ([4d9b887](https://github.com/aetherplatform/aether-go/commit/4d9b887600b9cbc57884ab381dbb760d7f6f3b3a))
* restore keyed webhook retries and add explicit retry opt-out ([c5b327a](https://github.com/aetherplatform/aether-go/commit/c5b327ab23d29e4332d1ec1c86bb556497c6ecb7))
* restore keyed webhook retries and add retry opt-out ([217e5a2](https://github.com/aetherplatform/aether-go/commit/217e5a2a3d0cad6a0525e9a7e456add24768f343))


### Miscellaneous Chores

* **release:** retain the beta.1 line for passwordless ([c6f01a4](https://github.com/aetherplatform/aether-go/commit/c6f01a4b93aa18e6e163706ada672b6baed989bc))

## 0.1.0-beta.1

Published 2026-09-12 as `v0.1.0-beta.1`.

### Fixed

- Retry webhook event publishing only with a nonblank body idempotency key;
  generate the retry-safety binding from the public contract.
- Add `DisableRetries` to API, Identity, and client-credentials configuration
  while preserving the existing zero-value defaults.
- Correct sandbox URLs and support contacts, document installation before the
  first tag, and synchronize the exported transport fixtures.

### Added

- Standard-library Core transport with typed errors and bounded retries.
- Server-only OAuth client-credentials provider.
- OAuth/OIDC Identity client with PKCE URL construction, discovery, userinfo,
  confidential token exchange, revocation, and introspection.
- Typed clients for the 3 Events, 29 Notifications, and 19 Webhooks operations
  approved by their public route inventories.
- Constant-time Webhooks signature verification against the shared contract
  vectors.
- Typed client for the 18 approved public Storage operations.
- Direct and multipart upload, download, pagination, and availability helpers.
- Source-only export, browser-boundary checks, and gated release automation.
