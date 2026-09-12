# Aether Go SDK Versioning and Release Policy

Status: locked for the initial public beta

The public module is `github.com/aetherplatform/aether-go`, published from the
`aetherplatform/aether-go` repository under Apache License 2.0. `Aether` is the
public publisher name. Security reports go to `security@useaether.co`; SDK
support goes to `support@useaether.co`.

## Public Boundary

The first beta contains Core, the server-only client-credentials package, and
the approved public contracts for Identity, Events, Notifications, Storage,
and Webhooks. Operator and partner-private routes, hosted Identity forms,
internal native authentication, service-JWT signing, NATS contracts,
deployment configuration, and backend source remain private.

## Go Support

The module declares Go 1.26 compatibility and is continuously tested on Go
1.26 and Go 1.27.1. Generation, security analysis, hosted certification, and
release jobs use Go 1.27.1. Go 1.26 support is reevaluated before `1.0.0`.

Consumers must use patched toolchains: at least Go 1.26.6 on the 1.26 line or
Go 1.27.1 on the 1.27 line as of 2026-09-12. The language directive in `go.mod`
is a compatibility floor, not a recommendation to use an outdated patch
release. Keep current with Go security updates and rebuild applications to
incorporate standard-library fixes.

## Semantic Versioning

The initial progression is:

```text
v0.1.0-beta.1
v0.1.0-beta.2
v0.1.0-rc.1
v0.1.0
v1.0.0
```

Every customer-visible change updates the changelog. Preview releases may
change incompatibly, but breaking changes require migration guidance. Stable
deprecation windows and supported-version commitments are published before
`v1.0.0`; urgent security retirements remain an exception.

## Publication

Go has no required centralized publish command. A release is published by
merging a reviewed release change, creating an immutable semantic-version tag,
pushing the commit and tag, and allowing Go tooling or a module proxy to
discover the repository.

Release Please prepares version and changelog changes. The protected release
workflow runs Go 1.27.1 certification and the hosted Identity, Events,
Notifications, Storage, and Webhooks proof before it may create later tags.
The one-time bootstrap workflow creates `v0.1.0-beta.1` only when
`AETHER_GO_SDK_BOOTSTRAP_RELEASE_ENABLED=true`. After verification, the
bootstrap switch is disabled and `AETHER_GO_SDK_RELEASE_ENABLED=true` enables
reviewed Release Please updates. Both switches remain unset or false until all
hosted sandbox endpoints, credentials, audiences, and the Storage namespace
are configured.

After release, a clean external project must successfully run:

```bash
go get github.com/aetherplatform/aether-go@v0.1.0-beta.1
```
