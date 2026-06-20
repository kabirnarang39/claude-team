# QA Report — User Authentication with JWT

**Agent:** qa-engineer  
**Phase:** qa  
**Confidence:** HIGH  

---

## Test Plan

### Coverage Targets
- Unit: token generation, password hashing, validation logic — **90%+**
- Integration: all 7 auth endpoints against real PostgreSQL — **100% happy path, 80%+ error cases**
- E2E: full registration → login → token refresh → logout flow — **see e2e-tester report**

---

## Test Cases

### Registration

| ID | Scenario | Expected |
|----|----------|----------|
| REG-01 | Valid email + strong password | 201, `user_id` returned, verification email queued |
| REG-02 | Duplicate email | 409 `email_already_registered` |
| REG-03 | Password < 8 chars | 400 `password_too_weak` |
| REG-04 | Password missing uppercase | 400 `password_too_weak` |
| REG-05 | Invalid email format | 400 `invalid_email` |
| REG-06 | Empty body | 400 |
| REG-07 | Rate limit (4th request same IP / hour) | 429 |

### Login

| ID | Scenario | Expected |
|----|----------|----------|
| LOG-01 | Correct credentials | 200, access + refresh token |
| LOG-02 | Wrong password | 401 `invalid_credentials` (no specifics) |
| LOG-03 | Unknown email | 401 `invalid_credentials` (same message, no enumeration) |
| LOG-04 | Unverified email | 401 `email_not_verified` |
| LOG-05 | 5th attempt same IP / 15 min | 429, `Retry-After` header set |
| LOG-06 | Access token expires after 15 min | 401 on protected endpoints |

### Token Refresh

| ID | Scenario | Expected |
|----|----------|----------|
| REF-01 | Valid refresh token | 200, new token pair; old token invalidated |
| REF-02 | Expired refresh token | 401 |
| REF-03 | Already-rotated token (replay) | 401 (outside 5s grace window) |
| REF-04 | Concurrent refresh with same token | One succeeds, other succeeds within 5s grace window |
| REF-05 | Tampered refresh token | 401 |

### Logout

| ID | Scenario | Expected |
|----|----------|----------|
| LGT-01 | Single device logout | 204; refresh token invalidated |
| LGT-02 | All devices logout | 204; all refresh tokens for user invalidated |
| LGT-03 | Logout with already-expired refresh token | 204 (idempotent) |

### Email Verification

| ID | Scenario | Expected |
|----|----------|----------|
| VER-01 | Valid token within 24h | 200, account verified |
| VER-02 | Token used twice | 400 `token_already_used` |
| VER-03 | Expired token | 400 `token_expired` |

### Password Reset

| ID | Scenario | Expected |
|----|----------|----------|
| RST-01 | Request for existing email | 200 (reset email queued) |
| RST-02 | Request for non-existent email | 200 (no user enumeration) |
| RST-03 | Confirm with valid token + strong password | 200 |
| RST-04 | Confirm with expired token | 400 |
| RST-05 | Confirm with used token | 400 |
| RST-06 | New password fails complexity | 400 `password_too_weak` |

---

## Integration Test Setup

- Dockerised PostgreSQL (`postgres:16-alpine`) spun up per test run via `testcontainers-go`
- Migrations applied before each test suite
- Each test case runs in a transaction, rolled back on completion
- Email sending mocked via interface injection

---

## Risks

- Concurrent refresh race (REF-04) is timing-sensitive; run with `-count=10` to expose flakiness
- Rate limiting tests need to control clock or stub IP — use `httptest` with `X-Forwarded-For` override
