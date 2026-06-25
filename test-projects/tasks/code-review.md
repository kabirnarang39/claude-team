Security-focused code review of our authentication middleware.

Files to review:
- internal/auth/middleware.go — JWT validation, token extraction
- internal/auth/refresh.go — refresh token rotation logic
- internal/auth/session.go — session store (Redis)

Key concerns:
- Are we validating JWT algorithm (alg field)? Algorithm confusion attacks?
- Refresh token rotation — is old token invalidated atomically?
- Are tokens appearing in logs, error messages, or headers?
- Missing rate limiting on /auth/refresh endpoint?
- CORS headers on auth endpoints — too permissive?
- Session fixation after login?

We had a pen test last year that flagged JWT alg:none bypass. Verify that's fixed.
