# Acceptance Criteria

## Scope

Implement email/password authentication with short-lived access tokens and rotating refresh tokens.

## Criteria

1. Given a new user submits a valid email and password, when registration succeeds, then the system creates the user and returns an authenticated session.
2. Given an existing user submits valid credentials, when login succeeds, then the system returns an access token and refresh token.
3. Given an access token is expired and a refresh token is valid, when refresh is requested, then the system rotates the refresh token and returns a new token pair.
4. Given a refresh token has already been used, when it is submitted again, then the system rejects it and records a security event.
5. Given a user logs out, when the active refresh token is revoked, then future refresh attempts fail.
6. Given invalid credentials are submitted repeatedly, when the rate limit threshold is reached, then the endpoint returns a rate-limit response.
7. Given a protected endpoint receives no valid access token, when the request is evaluated, then it returns unauthorized.
8. Given the JWT signing key changes, when old tokens are verified past the grace window, then they fail validation.

## Out Of Scope

- OAuth providers
- MFA enrollment
- Password reset email delivery
- Admin user management

## Open Questions

- Confirm preferred token transport: HTTP-only cookies or Authorization header.
- Confirm session lifetime policy for remembered devices.
