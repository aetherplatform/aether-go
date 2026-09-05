# Aether Webhooks Go SDK

The `webhooks` package provides all 19 customer-facing subscription, delivery,
replay, and managed inbound-endpoint operations.

`VerifySignature` verifies `X-Aether-Signature` against the exact raw request
body using HMAC-SHA256, a five-minute default tolerance, multiple `v1`
signatures during rotation, and constant-time comparison. Verify the bytes
before parsing or transforming the request body.
