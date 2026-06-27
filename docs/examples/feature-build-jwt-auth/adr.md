# ADR: JWT Access Tokens With Rotating Refresh Tokens

## Status

Proposed

## Context

The application needs stateless authorization for API requests while retaining server-side control over long-lived sessions.

## Decision

Use short-lived signed JWT access tokens and store hashed refresh tokens server-side. Refresh tokens rotate on every use. Reuse of an old refresh token is treated as a possible theft signal and revokes the refresh-token family.

## Consequences

Positive:

- API authorization can remain stateless for normal requests.
- Refresh-token rotation gives the server a revocation point.
- Token theft is easier to detect than with non-rotating refresh tokens.

Trade-offs:

- The refresh path requires a database or cache lookup.
- Key rotation and token-family invalidation need operational runbooks.
- Clients must handle refresh failure by returning to login.

## Implementation Notes

- Store only refresh-token hashes.
- Keep access-token TTL short.
- Rate-limit login and refresh endpoints.
- Add audit events for login, logout, refresh, reuse detection, and failed verification.
