# Changelog

All notable customer-visible changes are recorded here.

## 0.1.0-beta.1

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
