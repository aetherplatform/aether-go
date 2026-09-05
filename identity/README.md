# Aether Identity Go SDK

The `identity` package implements Aether's public OAuth 2.0 Authorization Code
with PKCE and OpenID Connect contracts.

The browser-safe surface builds authorization URLs, reads discovery/JWKS/scope
metadata, and calls userinfo with an explicit user-access token provider. It
never accepts `principal_id` and never treats browser cookies as API bearer
credentials.

Confidential token exchange, revocation, and introspection helpers require HTTP
Basic client authentication and are excluded from `GOOS=js` builds. Hosted
login, passwordless registration, consent forms, and native internal
authentication are intentionally not SDK APIs.

Use the root `clientcredentials.Provider` for cached, single-flight machine
token acquisition. The lower-level confidential Identity client intentionally
does not retry authorization-code or rotating-refresh-token exchanges.
