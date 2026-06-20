# Frontend Implementation Report — User Authentication with JWT

**Agent:** frontend-engineer  
**Phase:** engineering  
**Confidence:** HIGH  

---

## Implementation Plan

### Stack
- **Framework:** React 18 + TypeScript
- **State:** Zustand (auth store)
- **HTTP client:** `axios` with interceptors
- **Forms:** React Hook Form + Zod validation
- **Routing:** React Router v6

---

## Component Structure

```
src/
  auth/
    AuthProvider.tsx       # Context + token refresh orchestration
    useAuth.ts             # Hook exposing login/logout/user
    authStore.ts           # Zustand store (access token in memory)
  pages/
    LoginPage.tsx
    RegisterPage.tsx
    ResetPasswordPage.tsx
    VerifyEmailPage.tsx
  api/
    authClient.ts          # Axios instance with interceptors
```

---

## Token Storage Strategy

**Access token:** in-memory only (Zustand store). Never `localStorage` — XSS risk.  
**Refresh token:** `HttpOnly; Secure; SameSite=Strict` cookie, set by backend.

This means the frontend never reads the refresh token directly — it's sent automatically by the browser on `POST /auth/refresh`.

---

## Silent Refresh

```typescript
// authClient.ts
let refreshPromise: Promise<void> | null = null;

axiosInstance.interceptors.response.use(
  res => res,
  async error => {
    if (error.response?.status !== 401 || error.config._retry) {
      return Promise.reject(error);
    }
    // Deduplicate concurrent refresh calls
    if (!refreshPromise) {
      refreshPromise = authStore.getState().refreshTokens()
        .finally(() => { refreshPromise = null; });
    }
    await refreshPromise;
    error.config._retry = true;
    error.config.headers.Authorization = `Bearer ${authStore.getState().accessToken}`;
    return axiosInstance(error.config);
  }
);
```

Deduplication ensures a burst of 401s triggers exactly one refresh request, not N.

---

## Form Validation

```typescript
const loginSchema = z.object({
  email: z.string().email("Valid email required"),
  password: z.string().min(1, "Password required"),
});

const registerSchema = z.object({
  email: z.string().email("Valid email required"),
  password: z
    .string()
    .min(8, "At least 8 characters")
    .regex(/[A-Z]/, "At least one uppercase letter")
    .regex(/[0-9]/, "At least one number"),
});
```

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Access token location | Memory only | XSS cannot steal in-memory tokens |
| Refresh token location | HttpOnly cookie | JS cannot read it; browser sends it automatically |
| Concurrent 401 handling | Deduplication via shared Promise | Prevents multiple simultaneous refresh calls |
| Auth state persistence | Re-fetch on page load via silent refresh | No token in localStorage; page reload triggers `/auth/refresh` |

---

## Risks

- **Page load flicker:** user appears logged-out for ~200ms until silent refresh completes. Mitigate with a loading skeleton on `AuthProvider` mount.
- **Tab sync:** access token is per-tab (in-memory). Multiple tabs each maintain their own access token refreshed from the shared `HttpOnly` cookie — this is acceptable.
