# Requirements Analysis — User Authentication with JWT

**Agent:** requirements-analyst  
**Phase:** planning  
**Confidence:** HIGH  

---

## Functional Requirements

1. **Registration** — Users register with email + password. Email must be unique. Password minimum 8 characters, must contain uppercase, lowercase, and digit.
2. **Login** — Users authenticate with email + password. Returns access token (15 min TTL) and refresh token (7 day TTL).
3. **Token refresh** — Clients exchange a valid refresh token for a new access/refresh token pair. Old refresh token is immediately invalidated (rotation).
4. **Logout** — Invalidates the user's current refresh token. Access token remains valid until natural expiry.
5. **Logout all devices** — Invalidates all refresh tokens for the user.
6. **Token verification** — All protected endpoints validate the access token on every request.
7. **Role-based access** — Three roles: `admin`, `editor`, `viewer`. Roles stored in JWT claims.
8. **Email verification** — New accounts are `unverified`. Verification email sent on registration. Protected routes require verified status.
9. **Password reset** — User requests reset via email. Reset token expires in 1 hour, single-use.
10. **Rate limiting** — Login endpoint: 5 attempts per 15 minutes per IP. Registration: 3 per hour per IP.
11. **Audit logging** — Log all auth events (login, logout, failed login, password reset) with timestamp, IP, user agent.
12. **MFA stub** — Schema includes `mfa_enabled` and `mfa_secret` columns. MFA flows out of scope for v1, fields reserved.

## Non-Functional Requirements

- Access token must be verified without database lookup (stateless, RS256 signed)
- Refresh token store must survive server restart (database-backed, not in-memory)
- Auth endpoints must respond in < 200ms p99 under normal load
- Passwords must be hashed with bcrypt, cost factor ≥ 12

## Out of Scope (v1)

- OAuth / social login (Google, GitHub)
- MFA enforcement flows
- Account deletion / GDPR export
- Admin user management UI

## Open Questions

1. Should JWT public key be loaded from file or environment variable? → Recommend env var for 12-factor compliance.
2. What is the expected peak concurrent user count? → Affects rate limit thresholds and Redis sizing.
3. Is email sending synchronous or queued? → Recommend queued (SQS/BullMQ) but sync acceptable for MVP.

## Acceptance Criteria

- [ ] `POST /auth/register` creates user, sends verification email, returns 201
- [ ] `POST /auth/login` returns access + refresh token pair, logs event
- [ ] `POST /auth/refresh` rotates refresh token, returns new pair
- [ ] `DELETE /auth/logout` invalidates current refresh token
- [ ] `GET /me` returns user profile, requires valid access token
- [ ] Invalid/expired token returns 401 with `WWW-Authenticate: Bearer` header
- [ ] Rate limit exceeded returns 429 with `Retry-After` header
- [ ] All auth events appear in audit_log table
