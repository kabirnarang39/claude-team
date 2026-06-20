# Architecture Decision Record — User Authentication with JWT

**Agent:** senior-architect  
**Phase:** architecture  
**Confidence:** HIGH  

---

## Decision: JWT (RS256) + Opaque Refresh Tokens backed by PostgreSQL

### Context

We need a stateless, scalable authentication mechanism that supports multi-device sessions, token rotation, and future MFA. The two main candidates are:

1. **JWT-only (symmetric, HS256)** — simple, single-service
2. **JWT access token + opaque refresh token (RS256)** — industry standard for multi-service, zero-DB access token verification

### Decision

Use **RS256-signed JWT access tokens** (15 min TTL) paired with **opaque random refresh tokens** stored in PostgreSQL (7 day TTL with rotation).

### Rationale

| Concern | Choice | Why |
|---------|--------|-----|
| Access token verification | RS256 JWT | Stateless; future microservices verify with public key, no DB call |
| Refresh token storage | Opaque in DB | Can be invalidated server-side; JWTs cannot be revoked without a blocklist |
| Symmetric vs asymmetric | Asymmetric | Private key stays in auth service; other services only need public key |
| Session revocation | DB lookup on refresh | Refresh is rare (~every 15 min); DB cost is acceptable |

### Rejected Alternative: Sessions + Redis

Redis sessions would require every service to reach the session store on every request. Adds infrastructure dependency and latency. Not appropriate for our target scale.

---

## System Architecture

```
Client
  │
  ├─ POST /auth/register  ──► Auth Service ──► PostgreSQL (users, refresh_tokens)
  ├─ POST /auth/login     ──► Auth Service ──► PostgreSQL
  │     └─ returns: { access_token (JWT), refresh_token (opaque) }
  │
  ├─ GET /api/...  (with Authorization: Bearer <JWT>)
  │     └─► API Gateway verifies JWT signature with public key (no DB)
  │
  └─ POST /auth/refresh   ──► Auth Service ──► PostgreSQL (lookup + rotate token)
```

---

## Key Design Decisions

### Token Payload (JWT Claims)
```json
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "role": "editor",
  "iat": 1700000000,
  "exp": 1700000900,
  "jti": "unique-token-id"
}
```
`jti` enables future token revocation without a full blocklist.

### Refresh Token Schema
```sql
CREATE TABLE refresh_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,  -- SHA-256 of token; raw token never stored
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  replaced_by UUID REFERENCES refresh_tokens(id)  -- rotation chain
);
```

### Concurrent Refresh Race Mitigation
Short grace window (5 s): when a refresh token is rotated, it remains valid for 5 seconds. Handles mobile retries on 5xx without logging users out. Implemented via `replaced_at` timestamp + grace period check.

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Refresh token DB becomes bottleneck | Low | Medium | Index on `token_hash`; periodic cleanup of expired rows |
| Private key compromise | Very low | High | Key stored in secrets manager; rotation runbook documented |
| JWT theft (XSS) | Medium | High | Short TTL (15 min); consider `HttpOnly` cookie for access token |

---

## Next Steps

1. API designer: define endpoints and request/response schemas
2. DBA: finalise schema and indices
3. Backend: implement bcrypt (cost 12), RS256 signing, token rotation logic
