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
v0.1.0-beta.1.1  # optional follow-up in the beta.1 series
v0.1.0-beta.2    # only when the maintainer advances the beta milestone
v0.1.0-rc.1
v0.1.0
v1.0.0
```

The maintainer selected `v0.1.0-beta.1.2` for the passwordless SDK update,
keeping it on the existing beta.1 line. The reviewed release becomes installable
when the protected workflow creates its immutable tag. The earlier
`v0.1.0-beta.1` release remains available.

Every customer-visible change updates the changelog. Preview releases may
change incompatibly, but breaking changes require migration guidance. Stable
deprecation windows and supported-version commitments are published before
`v1.0.0`; urgent security retirements remain an exception.

## Publication

Go has no required centralized publish command. A release is published by
merging a reviewed release change, creating an immutable semantic-version tag,
pushing the commit and tag, and allowing Go tooling or a module proxy to
discover the repository.

Release Please prepares version and changelog changes. Tags omit the package
component so Go receives `v<version>` tags. The annotated `version.go` constant
is updated with the release manifest. Preserve these release-owned files when
exporting new source. If repository policy prevents the Actions token from
opening a PR, a maintainer opens the prepared release branch with the
`autorelease: pending` label; review and the protected publishing gate still
apply. The protected release workflow runs Go 1.27.1 certification and the hosted Identity, Events,
Notifications, Storage, and Webhooks proof before it may create later tags.
The one-time bootstrap workflow created `v0.1.0-beta.1` with
`AETHER_GO_SDK_BOOTSTRAP_RELEASE_ENABLED=true`. Bootstrap is now disabled and
`AETHER_GO_SDK_RELEASE_ENABLED=true` enables reviewed Release Please updates.
A newly configured repository must leave both switches unset or false until
its hosted sandbox endpoints, credentials, audiences, and Storage namespace
are configured. Every later release still requires certification and hosted
proof for its candidate.

The current preview is installable with:

```bash
go get github.com/aetherplatform/aether-go@v0.1.0-beta.1.3
```

After each release, a clean external project must fetch the exact new tag,
verify checksums, compile the public packages, and complete the hosted journey.

## First release evidence

`v0.1.0-beta.1` was published on 2026-09-12 from commit
`4d9b887600b9cbc57884ab381dbb760d7f6f3b3a` after the protected
[bootstrap workflow](https://github.com/aetherplatform/aether-go/actions/runs/34666284010)
passed certification and hosted proof. An external consumer successfully
fetched the tag, verified checksums, compiled the public packages, and passed
the five-platform hosted journey. Bootstrap is now disabled and the permanent
release workflow is enabled; both retain the required reviewer. No npm package
is required for Go.

## Configurable resend release evidence

`v0.1.0-beta.1.3` was published on 2026-09-14 from commit
`eb2656f9e2730a939a72c454e34c60e770b67669` after the protected
[release workflow](https://github.com/aetherplatform/aether-go/actions/runs/34840803955)
passed certification and hosted five-platform proof. A clean external project
fetched the exact version through `proxy.golang.org`, verified module checksums,
compiled every public package, and verified `aether.Version`. This release adds
acceptance of the server-configured 1–300 second resend interval; the server
default remains 60 seconds.
