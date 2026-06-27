# QA Report

## Test Plan

| Area | Cases |
| --- | --- |
| Registration | valid signup, duplicate email, weak password |
| Login | valid credentials, invalid password, unknown email |
| JWT verification | valid token, expired token, invalid signature, wrong audience |
| Refresh | valid rotation, reuse detection, revoked token, expired token |
| Logout | revokes current refresh token, does not revoke unrelated device |
| Rate limits | repeated login failures, repeated refresh failures |

## Recommended Automated Tests

```bash
go test ./internal/auth ./internal/http
npm test -- auth
```

## Risks

- Refresh-token storage must be transactionally safe to avoid race conditions during rotation.
- Client storage decision affects XSS/CSRF trade-offs and should be explicit before implementation.

## Verdict

Proceed after token transport decision is confirmed.
