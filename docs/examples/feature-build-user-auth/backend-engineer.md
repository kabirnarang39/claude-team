# Backend Implementation Report — User Authentication with JWT

**Agent:** backend-engineer  
**Phase:** engineering  
**Confidence:** HIGH  

---

## Implementation Plan

### Stack
- **Language:** Go 1.22
- **HTTP:** `net/http` + `chi` router
- **DB:** PostgreSQL via `pgx/v5`
- **JWT:** `golang-jwt/jwt/v5` (RS256)
- **Password hashing:** `golang.org/x/crypto/bcrypt` (cost 12)
- **Key storage:** environment variable `AUTH_PRIVATE_KEY_PEM`

---

## Core Components

### 1. Token Generation

```go
func (s *Service) IssueAccessToken(user *User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "role":  user.Role,
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(15 * time.Minute).Unix(),
        "jti":   uuid.New().String(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.privateKey)
}
```

### 2. Refresh Token Rotation

```go
func (s *Service) RotateRefreshToken(ctx context.Context, raw string) (*TokenPair, error) {
    hash := sha256Hex(raw)
    rt, err := s.store.GetRefreshToken(ctx, hash)
    if err != nil || rt.ExpiresAt.Before(time.Now()) {
        return nil, ErrInvalidToken
    }
    // Grace window: allow recently rotated tokens (handles concurrent refresh)
    if rt.ReplacedAt != nil && time.Since(*rt.ReplacedAt) > 5*time.Second {
        return nil, ErrInvalidToken
    }
    newRaw := secureRandom(32)
    newHash := sha256Hex(newRaw)
    if err := s.store.RotateRefreshToken(ctx, rt.ID, newHash, 7*24*time.Hour); err != nil {
        return nil, err
    }
    access, _ := s.IssueAccessToken(rt.User)
    return &TokenPair{AccessToken: access, RefreshToken: newRaw}, nil
}
```

### 3. Middleware

```go
func AuthMiddleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            claims, err := verifyJWT(raw, publicKey)
            if err != nil {
                http.Error(w, `{"error":"unauthorized"}`, 401)
                return
            }
            ctx := context.WithValue(r.Context(), ctxKeyClaims, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| bcrypt cost | 12 | ~400ms on commodity hardware; acceptable for login |
| Refresh token storage | SHA-256 hash of raw token | Raw token never stored; hash breach is useless without raw |
| Concurrent refresh | 5s grace window | Prevents logout on mobile retry without weakening security significantly |
| Error messages | Generic on auth failure | Prevents user enumeration on login and password reset |

---

## Risks & Open Items

- **Concurrent refresh race** mitigated by grace window; load-test at ≥ 100 concurrent refresh requests to validate
- **Key rotation runbook** not yet written — needed before production
- **Token blocklist** not implemented for access tokens; 15 min TTL is the revocation mechanism; document this limitation in runbook
