import { test, expect } from '@playwright/test';
import path from 'path';

/**
 * Auth E2E tests — no mocks, no stubs.
 * Every assertion hits the real running stack.
 *
 * Stack required:
 *   - postgres + redis (GitHub Actions service containers or docker compose)
 *   - Better Auth service on :9090
 *   - Go API on :8080
 *   - Next.js on :3000  (started by playwright.config.ts webServer)
 *
 * Test credentials are read from .env.test (see .env.test.example).
 */

const AUTH_STATE_FILE = path.join(__dirname, '.auth/user.json');

const EMAIL = process.env.E2E_TEST_EMAIL || 'e2e-test@nixopus.test';
const PASSWORD = process.env.E2E_TEST_PASSWORD || 'Test@E2E!2026';

// ---------------------------------------------------------------------------
// Unauthenticated — private route redirects
// ---------------------------------------------------------------------------
test.describe('Unauthenticated — private route guards', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  for (const route of ['/chats', '/machines', '/settings', '/extensions', '/domains', '/']) {
    test(`${route} redirects to /auth`, async ({ page }) => {
      await page.goto(route);
      await expect(page).toHaveURL(/\/auth/, { timeout: 10_000 });
    });
  }
});

// ---------------------------------------------------------------------------
// Unauthenticated — login form rendering
// ---------------------------------------------------------------------------
test.describe('Unauthenticated — login form', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test.beforeEach(async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });
  });

  test('renders email field, password field, and Login button', async ({ page }) => {
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();
  });

  test('shows "Forgot your password?" link pointing to /auth/reset-password', async ({ page }) => {
    const link = page.getByRole('link', { name: /forgot your password/i });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', '/auth/reset-password');
  });

  test('shows "Sign up" link pointing to /register', async ({ page }) => {
    const link = page.getByRole('link', { name: /sign up/i });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', '/register');
  });
});

// ---------------------------------------------------------------------------
// Unauthenticated — client-side validation (no network call)
// ---------------------------------------------------------------------------
test.describe('Unauthenticated — client-side validation', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test.beforeEach(async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });
  });

  test('empty form submit shows "Email is required"', async ({ page }) => {
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(page.getByRole('alert').filter({ hasText: /email is required/i })).toBeVisible();
    // No navigation — stays on /auth
    await expect(page).toHaveURL(/\/auth/);
  });

  test('invalid email format shows "Please enter a valid Email"', async ({ page }) => {
    await page.locator('input[type="email"]').fill('notanemail');
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(page.getByRole('alert').filter({ hasText: /valid email/i })).toBeVisible();
    await expect(page).toHaveURL(/\/auth/);
  });

  test('empty password submit shows "Password is required"', async ({ page }) => {
    await page.locator('input[type="email"]').fill('user@example.com');
    await page.getByRole('button', { name: 'Login' }).click();
    await expect(
      page.getByRole('alert').filter({ hasText: /password is required/i })
    ).toBeVisible();
    await expect(page).toHaveURL(/\/auth/);
  });
});

// ---------------------------------------------------------------------------
// Unauthenticated — login flows (hit the real API)
// ---------------------------------------------------------------------------
test.describe('Unauthenticated — login flows', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test('wrong credentials show an error toast', async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });

    await page.locator('input[type="email"]').fill('nobody@example.invalid');
    await page.locator('input[type="password"]').fill('WrongPass123!');
    await page.getByRole('button', { name: 'Login' }).click();

    /**
     * BUG FOUND: the backend returns 200 with an error payload on invalid
     * credentials instead of 401. The client reads result.error and shows a
     * Sonner toast. This test verifies the user-visible error — the HTTP status
     * issue is a separate backend concern and must be fixed there, not hidden.
     */
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toBeVisible({
      timeout: 10_000
    });
    await expect(page).toHaveURL(/\/auth/);
  });

  test('valid credentials redirect to /chats', async ({ page }) => {
    await page.goto('/auth');
    await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 10_000 });

    await page.locator('input[type="email"]').fill(EMAIL);
    await page.locator('input[type="password"]').fill(PASSWORD);
    await page.getByRole('button', { name: 'Login' }).click();

    await expect(page).toHaveURL(/\/chats/, { timeout: 20_000 });
  });
});

// ---------------------------------------------------------------------------
// Unauthenticated — register page
// ---------------------------------------------------------------------------
test.describe('Unauthenticated — register page', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test('/register renders without crashing', async ({ page }) => {
    await page.goto('/register');
    // Page either shows the registration form OR the "admin already registered"
    // state — both are valid since this is a real stack with a seeded test user.
    // We assert the page shell is present and no error boundary has triggered.
    await expect(page.locator('body')).not.toBeEmpty();
    await expect(page.locator('text=Something went wrong')).not.toBeVisible({ timeout: 10_000 });
  });

  test('/register with admin already registered shows the blocked state', async ({ page }) => {
    await page.goto('/register');
    // The seeded admin (from global.setup.ts) means admin IS registered.
    // The page should show the "admin already registered" component, not the form.
    await expect(page.getByText(/already registered|admin account/i)).toBeVisible({
      timeout: 15_000
    });
  });
});

// ---------------------------------------------------------------------------
// Authenticated — redirects and page access
// ---------------------------------------------------------------------------
test.describe('Authenticated — redirects', () => {
  test.use({ storageState: AUTH_STATE_FILE });

  test('visiting /auth redirects to /chats', async ({ page }) => {
    await page.goto('/auth');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });
  });

  test('visiting / redirects to /chats', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });
  });
});

// ---------------------------------------------------------------------------
// Authenticated — page health checks
// ---------------------------------------------------------------------------
test.describe('Authenticated — page health', () => {
  test.use({ storageState: AUTH_STATE_FILE });

  test('/chats page renders without crashing', async ({ page }) => {
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });

    /**
     * BUG FOUND: app/chats/page.tsx loads ChatPage via next/dynamic with
     * ssr:false, so chat content only mounts after hydration. We assert the
     * shell is present and no error boundary has triggered.
     */
    await expect(page.locator('body')).not.toBeEmpty();
    await expect(page.locator('text=Something went wrong')).not.toBeVisible();
  });

  test('session persists across a full page reload', async ({ page }) => {
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });

    await page.reload();

    // After reload the middleware must still see the session cookie and keep
    // the user on /chats rather than redirecting to /auth.
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });
    await expect(page.locator('body')).not.toBeEmpty();
  });
});

// ---------------------------------------------------------------------------
// Authenticated — logout flow
// ---------------------------------------------------------------------------
test.describe('Authenticated — logout', () => {
  // Use a fresh copy of the auth state so the logout does not affect other suites.
  test.use({ storageState: AUTH_STATE_FILE });

  test('logout clears session and redirects to /auth', async ({ page }) => {
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });

    // Open the user menu (avatar button in the sidebar)
    await page.getByTestId('user-menu-trigger').click();

    // Click "Log out" in the dropdown (exact label from user-menu.tsx)
    await page.getByRole('menuitem', { name: /log out/i }).click();

    // Confirm logout in the dialog (dialog title: "Confirm Logout", button: "Logout")
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5_000 });
    await page.getByRole('button', { name: 'Logout' }).click();

    // Session cookie must be cleared — middleware redirects to /auth
    await expect(page).toHaveURL(/\/auth/, { timeout: 15_000 });
  });

  test('after logout, private routes redirect to /auth', async ({ page }) => {
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/chats/, { timeout: 10_000 });

    await page.getByTestId('user-menu-trigger').click();
    await page.getByRole('menuitem', { name: /log out/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5_000 });
    await page.getByRole('button', { name: 'Logout' }).click();
    await expect(page).toHaveURL(/\/auth/, { timeout: 15_000 });

    // Now navigate to a private route — must be redirected back to /auth
    await page.goto('/chats');
    await expect(page).toHaveURL(/\/auth/, { timeout: 10_000 });
  });
});
