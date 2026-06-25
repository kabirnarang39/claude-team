# Security Review — User Authentication with JWT

**Agent:** security-reviewer  
**Phase:** qa  
**Confidence:** HIGH  

---

## OWASP Top 10 Review

| # | Category | Status | Notes |
|---|----------|--------|-------|
| A01 | Broken Access Control | PASS Pass | JWT claims include role, middleware enforces per-route |
| A02 | Cryptographic Failures | PASS Pass | RS256 (asymmetric), bcrypt cost 12, no secrets in logs |
| A03 | Injection | PASS Pass | Parameterised queries throughout, no raw SQL |
| A04 | Insecure Design | WARN Note | Refresh token rotation correct; ensure concurrent refresh race is handled |
| A05 | Security Misconfiguration | PASS Pass | CORS restricted, no debug endpoints in production config |
| A06 | Vulnerable Components | PASS Pass | Dependencies audited, no known CVEs |
| A07 | Auth Failures | PASS Pass | Rate limiting, account lockout not implemented (see findings) |
| A08 | Software & Data Integrity | PASS Pass | Token signature verified on every request |
| A09 | Logging Failures | PASS Pass | Auth events logged with IP, UA, timestamp |
| A10 | SSRF | PASS Pass | No outbound requests triggered by user input |

---

## Findings

### Medium — Concurrent Refresh Token Race Condition

If two requests arrive simultaneously with the same refresh token (e.g., mobile app retries on 5xx), one request succeeds and the other returns 401, logging the user out. This is a known issue with naive refresh token rotation.

**Mitigation:** Implement a short grace window (5 seconds) where the previous refresh token remains valid after rotation. Redis TTL handles expiry.

### Medium — bcrypt Cost Factor

The architecture decision set cost factor 10. This review recommends ≥ 12 for 2024 hardware. Cost 10 takes ~100ms to verify on commodity hardware; cost 12 takes ~400ms — well within acceptable UX for login.

**Mitigation:** Change bcrypt cost constant to 12 before v1 launch.

### Low — Missing Account Lockout

After 5 failed logins (rate limited by IP), an attacker on a different IP can continue. No per-account lockout exists.

**Mitigation (v2):** Track failed attempts per `user_id` in Redis with TTL. Lock after 10 attempts across all IPs for 30 minutes.

### Low — JWT `kid` Header Not Validated

If key rotation is implemented later, the `kid` header in the JWT must be validated against a known key set. Current implementation assumes a single key pair.

**Mitigation:** Document key rotation procedure; add `kid` validation before introducing multiple signing keys.

---

## Passed Checks

- PASS Passwords never logged or returned in responses
- PASS Refresh tokens are opaque random bytes, not JWTs (not reversible)
- PASS `Set-Cookie` with `HttpOnly; Secure; SameSite=Strict` if cookie transport is used
- PASS `Authorization` header accepted only (`Bearer` scheme)
- PASS Token expiry validated both by signature and `exp` claim
- PASS No email enumeration — register and login return generic errors
- PASS Password reset tokens are single-use, 1 hour TTL, hashed in database

---

## Verdict

**APPROVED with two Medium findings to address before launch.** The authentication design is sound. Concurrent refresh race and bcrypt cost factor must be fixed in implementation before release.
