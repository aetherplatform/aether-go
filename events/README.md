# Aether Events Go SDK

The `events` package provides typed, read-only access to Aether's public event
catalog and governed JSON schemas. It exposes three allowlisted operations and
does not expose NATS credentials or consumer-group administration.

Use a client-access token with `events:catalog/*:read` for catalog discovery and
the corresponding exact schema capability when reading a concrete schema.
