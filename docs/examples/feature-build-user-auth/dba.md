# Database Report — User Authentication with JWT

**Agent:** dba  
**Phase:** engineering  
**Confidence:** HIGH  

---

## Schema

```sql
-- Users
CREATE TABLE users (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email             TEXT        NOT NULL UNIQUE,
  password_hash     TEXT        NOT NULL,
  role              TEXT        NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin','editor','viewer')),
  email_verified    BOOLEAN     NOT NULL DEFAULT false,
  mfa_enabled       BOOLEAN     NOT NULL DEFAULT false,
  mfa_secret        TEXT,
  failed_login_count INT        NOT NULL DEFAULT 0,
  last_failed_login TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Refresh tokens
CREATE TABLE refresh_tokens (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash   TEXT        NOT NULL UNIQUE,  -- SHA-256, hex-encoded
  expires_at   TIMESTAMPTZ NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  replaced_by  UUID        REFERENCES refresh_tokens(id),
  replaced_at  TIMESTAMPTZ
);

-- Email verification tokens
CREATE TABLE email_tokens (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT        NOT NULL UNIQUE,
  kind       TEXT        NOT NULL CHECK (kind IN ('verify_email','password_reset')),
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## Indices

```sql
-- Refresh token lookup (hot path: every token rotation)
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Cleanup: expired token sweep job
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)
  WHERE replaced_by IS NULL;

-- Email token lookup
CREATE INDEX idx_email_tokens_token_hash ON email_tokens(token_hash)
  WHERE used_at IS NULL;

-- User lookup by email (login hot path)
-- Covered by UNIQUE constraint; no extra index needed.
```

---

## Migration Strategy

All schema changes via sequential numbered migration files (`001_initial.sql`, `002_...`). Applied by `golang-migrate` on server startup. Rollback scripts maintained alongside each forward migration.

---

## Maintenance

### Expired Token Cleanup

Run nightly via `pg_cron` or application job:
```sql
DELETE FROM refresh_tokens WHERE expires_at < now() - INTERVAL '1 day';
DELETE FROM email_tokens    WHERE expires_at < now() - INTERVAL '1 day';
```

### Updated_at Trigger

```sql
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## Performance Notes

- `token_hash` lookup is point-query on unique index — O(log n), negligible at scale
- `users(email)` covered by unique constraint index — no additional index needed
- Refresh token table will grow; cleanup job critical to prevent unbounded growth
- At 10k DAU with 15min token TTL: ~960k rotations/day. Index lookup stays fast up to ~10M rows before vacuum/cleanup becomes critical.
