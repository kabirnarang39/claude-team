# API Design — User Authentication with JWT

**Agent:** api-designer  
**Phase:** architecture  
**Confidence:** HIGH  

---

## Endpoints

### POST /auth/register

**Request**
```json
{
  "email": "user@example.com",
  "password": "SecurePass1"
}
```

**Response 201**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Verification email sent"
}
```

**Errors:** `400` (validation), `409` (duplicate email)

---

### POST /auth/login

**Request**
```json
{
  "email": "user@example.com",
  "password": "SecurePass1"
}
```

**Response 200**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "dGhpcyBpcyBhIHJhbmRvbSB0b2tlbg=="
}
```

**Errors:** `401` (invalid credentials), `429` (rate limited — 5 attempts / 15 min / IP)

---

### POST /auth/refresh

**Request**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJhbmRvbSB0b2tlbg=="
}
```

**Response 200** — same schema as `/auth/login` response (new token pair)

**Errors:** `401` (expired, unknown, or already rotated token)

---

### POST /auth/logout

**Headers:** `Authorization: Bearer <access_token>`

**Request**
```json
{
  "all_devices": false
}
```

**Response 204** — no body

`all_devices: true` invalidates all refresh tokens for the authenticated user.

---

### POST /auth/verify-email

**Request**
```json
{
  "token": "email-verification-token"
}
```

**Response 200**
```json
{
  "message": "Email verified"
}
```

**Errors:** `400` (expired or invalid token)

---

### POST /auth/password-reset/request

**Request**
```json
{
  "email": "user@example.com"
}
```

**Response 200** — always, regardless of whether email exists (prevents user enumeration)

---

### POST /auth/password-reset/confirm

**Request**
```json
{
  "token": "reset-token",
  "new_password": "NewSecurePass1"
}
```

**Response 200**
```json
{
  "message": "Password updated"
}
```

**Errors:** `400` (expired/used token, password fails validation)

---

## Authentication Header

All protected endpoints require:
```
Authorization: Bearer <access_token>
```

Missing or invalid token → `401 Unauthorized`  
Valid token but insufficient role → `403 Forbidden`

---

## Error Response Schema

All error responses follow:
```json
{
  "error": "human_readable_code",
  "message": "Description of what went wrong"
}
```

Example:
```json
{
  "error": "invalid_credentials",
  "message": "Email or password is incorrect"
}
```

---

## Rate Limits

| Endpoint | Limit | Window | Scope |
|----------|-------|--------|-------|
| POST /auth/login | 5 requests | 15 min | Per IP |
| POST /auth/register | 3 requests | 1 hour | Per IP |
| POST /auth/password-reset/request | 3 requests | 1 hour | Per IP |
