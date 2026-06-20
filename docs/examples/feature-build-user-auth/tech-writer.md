# Product Requirements Document — User Authentication with JWT

**Agent:** tech-writer  
**Phase:** planning  
**Confidence:** HIGH  

---

## Overview

This document captures product requirements for a JWT-based user authentication system. It is the authoritative reference for the engineering, QA, and DevOps teams working on this feature.

**Scope:** Registration, login, token refresh, logout, password reset, email verification, role-based access.  
**Out of scope (v1):** MFA flows, OAuth/SSO, social login.

---

## User Stories

| ID | As a… | I want to… | So that… |
|----|--------|------------|----------|
| US-01 | new user | register with my email and password | I can access the platform |
| US-02 | registered user | log in and receive a JWT | I can authenticate subsequent requests |
| US-03 | logged-in user | refresh my access token silently | my session continues without re-login |
| US-04 | logged-in user | log out from this device | my session is ended securely |
| US-05 | logged-in user | log out from all devices | I can revoke access if my account is compromised |
| US-06 | new user | verify my email address | my account is activated |
| US-07 | user who forgot password | reset it via email | I can regain access without contacting support |
| US-08 | admin | assign roles to users | I can control what each user can do |

---

## Acceptance Criteria

### Registration (US-01)
- Accepts `email` (valid format, unique) and `password` (≥8 chars, ≥1 uppercase, ≥1 digit)
- Returns `201 Created` with `user_id`
- Sends verification email immediately
- Duplicate email returns `409 Conflict`

### Login (US-02)
- Accepts `email` + `password`
- Returns access token (15 min TTL) and refresh token (7 day TTL)
- Failed login returns `401` (no indication which field is wrong)
- Brute-force protection: 5 attempts per 15 min per IP → `429 Too Many Requests`

### Token Refresh (US-03)
- Accepts valid refresh token
- Returns new access + refresh token pair; old refresh token is invalidated
- Expired or unknown refresh token returns `401`

### Logout (US-04 / US-05)
- Single device: invalidates current refresh token
- All devices: invalidates all refresh tokens for user
- Endpoint succeeds even if token already invalidated (idempotent)

### Email Verification (US-06)
- Token in email is single-use, expires in 24 hours
- Clicking link: `200 OK`, account moves to `verified`
- Expired/used token: `400 Bad Request` with clear message

### Password Reset (US-07)
- Request reset: accepts email, always returns `200` (no user enumeration)
- Reset token: single-use, 1-hour TTL
- New password must meet same complexity rules as registration

---

## Open Questions

| # | Question | Owner | Due |
|---|----------|-------|-----|
| 1 | Should refresh tokens be stored in `HttpOnly` cookie or returned in response body? | Backend + Frontend | Before API design |
| 2 | Do we need SSO/OAuth in v2? Affects token schema design now. | Product | Sprint planning |
| 3 | What is the acceptable p99 latency for login endpoint under load? | Platform | Before perf testing |

---

## Out of Scope (v1)

- MFA (TOTP / SMS) — schema reserves columns, flows deferred
- OAuth2 / OIDC social login
- IP allowlisting
- Device fingerprinting
