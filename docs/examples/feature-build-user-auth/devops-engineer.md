# DevOps Report — User Authentication with JWT

**Agent:** devops-engineer  
**Phase:** devops  
**Confidence:** HIGH  

---

## Deliverables

### Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /auth-service ./cmd/auth

FROM gcr.io/distroless/static-debian12
COPY --from=builder /auth-service /auth-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/auth-service"]
```

Multi-stage build. Distroless base: no shell, no package manager, minimal CVE surface. Non-root user.

---

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL DSN (`postgres://user:pass@host/db?sslmode=require`) |
| `AUTH_PRIVATE_KEY_PEM` | Yes | RS256 private key (PEM). Inject from secrets manager. |
| `AUTH_PUBLIC_KEY_PEM` | Yes | RS256 public key. Used by other services for JWT verification. |
| `AUTH_BCRYPT_COST` | No | Default: 12 |
| `AUTH_ACCESS_TOKEN_TTL` | No | Default: `15m` |
| `AUTH_REFRESH_TOKEN_TTL` | No | Default: `168h` (7 days) |
| `SMTP_HOST` | Yes | Email relay host |
| `SMTP_FROM` | Yes | From address for verification/reset emails |
| `APP_BASE_URL` | Yes | Base URL for email links (e.g. `https://app.example.com`) |

Keys stored in **AWS Secrets Manager** / **Vault** — never in environment files or repos.

---

### CI/CD Pipeline

```yaml
# .github/workflows/auth.yml
name: Auth Service

on:
  push:
    paths: ['services/auth/**']

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out | grep total

  build-push:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v5
        with:
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/org/auth-service:${{ github.sha }}
```

---

### Helm Chart Values (excerpt)

```yaml
image:
  repository: ghcr.io/org/auth-service
  tag: ""  # set per deploy

replicaCount: 2
resources:
  requests: { cpu: 100m, memory: 64Mi }
  limits:   { cpu: 500m, memory: 128Mi }

readinessProbe:
  httpGet: { path: /health, port: 8080 }
  initialDelaySeconds: 5

env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef: { name: auth-secrets, key: database-url }
  - name: AUTH_PRIVATE_KEY_PEM
    valueFrom:
      secretKeyRef: { name: auth-secrets, key: private-key-pem }
```

---

## Monitoring

| Signal | Tool | Alert threshold |
|--------|------|-----------------|
| `auth_login_errors_total` (4xx rate) | Prometheus | > 10% of requests for 5 min |
| `auth_login_latency_p99` | Prometheus | > 500ms for 2 min |
| Failed login spike | Grafana alert | > 50 failures/min (potential brute force) |
| Refresh token DB row count | Prometheus | > 5M rows (cleanup job may be failing) |

---

## Runbook: Key Rotation

1. Generate new RS256 key pair
2. Deploy new public key to all services (read old + new during transition)
3. Update `AUTH_PRIVATE_KEY_PEM` secret, rolling restart auth service
4. All new tokens signed with new key; old tokens expire within 15 min
5. Remove old public key from services after 20 min
