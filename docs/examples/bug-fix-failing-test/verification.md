# Verification

## Commands

```bash
go test -race ./internal/auth
go test ./internal/http
```

## Expected Result

- Refresh-token replay test passes consistently.
- Valid refresh flow still rotates tokens.
- Reuse of a consumed token revokes the token family.

## Residual Risk

Database isolation behavior should be checked against the production database engine, not only the local test database.
