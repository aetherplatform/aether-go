# Aether Identity Go SDK

The `identity` package supports Aether's public email/SMS passwordless API and
OAuth/OIDC authorization-code flow with S256 PKCE. Passwordless endpoints require
operator enablement for your registered client and configured delivery. This
branch prepares a beta release; hosted availability must be verified separately.

`Client` is safe to compile for browsers (`GOOS=js`). Set its public `ClientID` for
passwordless and OAuth exchange, refresh and revocation. `ConfidentialClient`
uses registered Basic credentials and is excluded from browser builds.
Introspection and client-credentials grants remain confidential-only.

```go
client, err := identity.NewClient(identity.Config{
    BaseURL: "https://your-configured-auth-host.example",
    ClientID: "registered-public-client-id",
})
if err != nil { return err }
proof, err := identity.GeneratePKCE()
if err != nil { return err }
state, err := identity.GenerateState()
if err != nil { return err }
scope := "openid email offline_access"
started, err := client.StartPasswordless(ctx, identity.PasswordlessStartRequest{
    Identifier: email, Channel: identity.ChannelEmail,
    RedirectURI: "https://app.example/callback", Scope: &scope, State: &state,
    CodeChallenge: proof.Challenge,
})
if err != nil { return err }
// Preserve started.Transaction, started.ChallengeID, proof.Verifier and state
// in application-owned storage while the user reads and enters their code.
result, err := client.VerifyPasswordless(ctx, identity.PasswordlessVerifyRequest{
    Transaction: started.Transaction, ChallengeID: started.ChallengeID,
    Identifier: email, Channel: identity.ChannelEmail, Code: userEnteredCode,
    CodeVerifier: proof.Verifier,
})
if err != nil { return err }
switch outcome := result.(type) {
case *identity.PasswordlessAuthorized:
    if outcome.State == nil || *outcome.State != state { return errors.New("state mismatch") }
    tokens, err := client.ExchangeOAuthToken(ctx, identity.TokenRequest{
        GrantType: identity.GrantAuthorizationCode, Code: outcome.Code,
        RedirectURI: "https://app.example/callback", CodeVerifier: proof.Verifier,
    })
    if err != nil { return err }
    // Store tokens using your application's session policy. Never log them.
    _ = tokens
case *identity.PasswordlessHostedCompletionRequired:
    // Open outcome.ContinuationURL in the browser. On your registered callback:
    // callback, err := identity.ValidateOAuthCallback(callbackURL, redirectURI, state)
    // Exchange callback.Code with the original verifier after validation succeeds.
case *identity.PasswordlessCustomCompletionRequired:
    // Render outcome.Requirements only for a client approved for custom completion.
    // Call CompletePasswordless with outcome.Continuation and the original verifier
    // after the user approves or denies. Registration approval also sends the
    // verified identifier, profile and exact required terms/privacy versions.
}
```

`CompletePasswordless` returns either `*PasswordlessAuthorized` or
`*PasswordlessDenied`. Denial is a handled result (`access_denied`), not a token.
A hosted-only client cannot use custom completion to bypass the hosted page.
`PasswordlessRegistrationConsent` and `PasswordlessRegistrationProfile` provide
named types for registration data.

After hosted completion, `ValidateOAuthCallback` verifies the exact redirect,
original state, query uniqueness and success/error shape. It accepts HTTPS
callbacks, including Universal Links/App Links. Loopback HTTP callbacks must be
explicitly registered for development. Store browser transaction material in
memory by default; mobile applications can use secure application storage. The
SDK does not persist tokens or transactions and never installs an SSO cookie.

Codes expire after five minutes; transactions expire after ten minutes. Resend
only after an explicit user action and the returned 60-second cooldown. Calling
`StartPasswordless` again supersedes the earlier pending challenge. All sends,
verification, completion, code exchange and refresh are single attempts; even a
network interruption requires application reconciliation or a fresh user action.
For rate limits, inspect `aether.Error.Retry.RetryAfterSeconds`.

Refresh using `ExchangeOAuthToken` with `GrantRefreshToken`, storing the returned
successor refresh token atomically. Reuse revokes the family. For logout, revoke
the current refresh token with `RevokeOAuthToken` and clear application storage;
refresh-token revocation invalidates that family. Revoking only an access token
does not revoke the refresh token. Disabling passwordless stops new starts while
existing token revocation stays available.

For backend screens, construct `NewConfidentialClient` with `ClientID` and
`ClientSecret` and call the same passwordless methods. Keep those credentials on
the server. Use `clientcredentials.Provider` for cached machine tokens.
Hosted continuation pages are browser UI and are never called as JSON SDK APIs.
