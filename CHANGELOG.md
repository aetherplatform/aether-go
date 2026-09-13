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
