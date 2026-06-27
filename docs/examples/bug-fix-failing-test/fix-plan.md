# Fix Plan

1. Add a concurrency regression test that submits the same refresh token twice.
2. Replace read-then-update with a conditional update.
3. If affected rows is zero, treat the token as reused or revoked.
4. Revoke the refresh-token family on reuse.
5. Run auth unit tests with race detection.

## Expected Test Command

```bash
go test -race ./internal/auth
```
