# E2E Test Report — User Authentication with JWT

**Agent:** e2e-tester  
**Phase:** qa  
**Confidence:** HIGH  

---

## Test Plan

Framework: **Playwright** (TypeScript)  
Target: staging environment with real database and email via Mailhog

---

## Test Suites

### Suite 1: Happy Path — Full Auth Flow

```typescript
test('full registration → login → refresh → logout flow', async ({ page }) => {
  // 1. Register
  await page.goto('/register');
  await page.fill('[name=email]', `test+${Date.now()}@example.com`);
  await page.fill('[name=password]', 'SecurePass1');
  await page.click('button[type=submit]');
  await expect(page.locator('[data-testid=verify-banner]')).toBeVisible();

  // 2. Verify email (fetch token from Mailhog API)
  const token = await getVerificationToken(email);
  await page.goto(`/verify-email?token=${token}`);
  await expect(page.locator('[data-testid=verified-success]')).toBeVisible();

  // 3. Login
  await page.goto('/login');
  await page.fill('[name=email]', email);
  await page.fill('[name=password]', 'SecurePass1');
  await page.click('button[type=submit]');
  await expect(page).toHaveURL('/dashboard');

  // 4. Wait for token expiry simulation (mock clock or wait)
  // Trigger silent refresh by making an authenticated API call
  const resp = await page.evaluate(() =>
    fetch('/api/me', { credentials: 'include' }).then(r => r.status)
  );
  expect(resp).toBe(200);

  // 5. Logout
  await page.click('[data-testid=logout-btn]');
  await expect(page).toHaveURL('/login');

  // 6. Verify refresh token invalidated
  const refreshResp = await page.evaluate(() =>
    fetch('/auth/refresh', { method: 'POST', credentials: 'include' }).then(r => r.status)
  );
  expect(refreshResp).toBe(401);
});
```

### Suite 2: Password Reset Flow

```typescript
test('password reset full flow', async ({ page }) => {
  await page.goto('/forgot-password');
  await page.fill('[name=email]', existingUser.email);
  await page.click('button[type=submit]');
  await expect(page.locator('[data-testid=reset-sent]')).toBeVisible();

  const token = await getResetToken(existingUser.email);
  await page.goto(`/reset-password?token=${token}`);
  await page.fill('[name=password]', 'NewSecurePass2');
  await page.click('button[type=submit]');
  await expect(page.locator('[data-testid=reset-success]')).toBeVisible();

  // Old password no longer works
  await page.goto('/login');
  await page.fill('[name=email]', existingUser.email);
  await page.fill('[name=password]', existingUser.oldPassword);
  await page.click('button[type=submit]');
  await expect(page.locator('[data-testid=login-error]')).toBeVisible();

  // New password works
  await page.fill('[name=password]', 'NewSecurePass2');
  await page.click('button[type=submit]');
  await expect(page).toHaveURL('/dashboard');
});
```

### Suite 3: Rate Limiting (Visual Feedback)

```typescript
test('login rate limit shows user-friendly error', async ({ page }) => {
  await page.goto('/login');
  for (let i = 0; i < 5; i++) {
    await page.fill('[name=email]', 'user@example.com');
    await page.fill('[name=password]', 'wrongpassword');
    await page.click('button[type=submit]');
    await page.waitForResponse(r => r.url().includes('/auth/login'));
  }
  await expect(page.locator('[data-testid=rate-limit-msg]')).toBeVisible();
  await expect(page.locator('[data-testid=rate-limit-msg]')).toContainText('Too many attempts');
});
```

---

## Edge Cases to Cover

- [ ] Browser back after logout does not expose protected page (CSP + no-cache headers)
- [ ] Multiple tabs: logout in one tab → other tabs show session-expired banner within 30s (polling or BroadcastChannel)
- [ ] Session persists across page reload (silent refresh on load)
- [ ] Mobile viewport: login form usable on 375px width
- [ ] Password field shows/hides correctly (eye icon toggle)

---

## Test Infrastructure

- **Mailhog** for email capture in staging
- **Playwright fixtures** for user creation/cleanup
- **`--shard`** flag for parallel CI runs across 4 workers
- Screenshots and video captured on failure (`video: 'retain-on-failure'`)
