import { test, expect } from '@playwright/test';
import path from 'path';

/**
 * Auth E2E tests — no mocks, no stubs.
 * Every assertion hits the real running stack.
 *
 * Stack required:
 *   - docker compose -f docker-compose-test.yml up -d
 *   - Go API running on :8080
 *   - Next.js running on :3000  (started by playwright.config.ts webServer)
 *
 * Test credentials are read from .env.test (see .env.test.example).
 */

const AUTH_STATE_FILE = path.join(__dirname, '.auth/user.json');

const EMAIL = process.env.E2E_TEST_EMAIL || 'e2e-test@nixopus.test';
const PASSWORD = process.env.E2E_TEST_PASSWORD || 'Test@E2E!2026';

// ---------------------------------------------------------------------------
// Unauthenticated flows — fresh browser context (no cookies)
// ---------------------------------------------------------------------------
test.describe('Unauthenticated', () => {
  // Force a completely empty storage state so there is no bleed from other tests.
  test.use({ storageState: { cookies: [], origins: [] } });

  test('visiting a private route redirects to /auth', async ({ page }) => {
    await page.goto('/chats');
    // Middleware must redirect before the page renders — no JS needed.
    await expect(page).toHaveURL(/\/auth/, { timeout: 10_000 });
  });

  test('visiting /machines redirects to /auth', async ({ page }) => {
    await page.goto('/machines');
    await expect(page).toHaveURL(/\/auth/, { timeout: 10_000 });
  });

  test('visiting / redirects to /auth', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/auth/, { timeout: 10_000 });
  });

  test('login page renders email and password fields', async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();
  });

  test('wrong credentials show an error toast', async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });

    await page.locator('input[type="email"]').fill('nobody@example.invalid');
    await page.locator('input[type="password"]').fill('WrongPass123!');
    await page.getByRole('button', { name: 'Login' }).click();

    /**
     * BUG FOUND: the backend returns a 200 with error payload on invalid credentials
     * instead of a 401. The client reads result.error and shows a toast.
     * This test verifies the user-visible error appears — the HTTP status issue
     * is a separate backend concern and should be fixed there, not hidden here.
     *
     * Sonner renders toasts with [data-sonner-toast][data-type="error"].
     */
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toBeVisible({
      timeout: 10_000
    });

    // URL must NOT change — user stays on /auth after failed login.
    await expect(page).toHaveURL(/\/auth/);
  });

  test('valid credentials redirect to /chats', async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });

    await page.locator('input[type="email"]').fill(EMAIL);
    await page.locator('input[type="password"]').fill(PASSWORD);
    await page.getByRole('button', { name: 'Login' }).click();

    // Real session cookie must be set and middleware must allow /chats.
    await expect(page).toHaveURL(/\/chats/, { timeout: 20_000 });
  });
});

// ---------------------------------------------------------------------------
// Authenticated flows — reuse the session saved by global.setup.ts
// ---------------------------------------------------------------------------
test.describe('Authenticated', () => {
  test.use({ storageState: AUTH_STATE_FILE });

  test('visiting /auth redirects to /chats', async ({ page }) => {
    await page.goto('/auth');
    // Middleware sees better-auth.session_token and sends authenticated users
    // away from landing auth routes. This is the real middleware — not stubbed.
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });
  });

  test('visiting / redirects to /chats', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });
  });

  test('/chats page renders without crashing', async ({ page }) => {
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });

    /**
     * BUG FOUND: app/chats/page.tsx loads ChatPage via next/dynamic with ssr:false,
     * which means the chat UI only mounts after hydration. The loading.tsx skeleton
     * should show during this window. We assert the page shell is present (not blank)
     * but we do NOT assert chat-specific content since that depends on AI config.
     */
    // The root layout body must be present — confirms no crash in shell providers.
    await expect(page.locator('body')).not.toBeEmpty();

    // No uncaught error boundary should be showing.
    await expect(page.locator('text=Something went wrong')).not.toBeVisible();
  });
});
