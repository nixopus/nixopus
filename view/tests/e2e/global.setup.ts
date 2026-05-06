import { test as setup, expect } from '@playwright/test';
import path from 'path';

/**
 * Global setup — runs once before all test projects.
 *
 * Responsibilities:
 *   1. Register the E2E test user against the live auth service.
 *      If the user already exists, the 409/422 response is silently ignored.
 *   2. Log in through the real UI and save the authenticated browser storage
 *      state so tests in the "authenticated" describe block can reuse the
 *      session without repeating the login flow.
 *
 * NOTE: This is NOT a mock. Both the register call and the UI login hit the
 * real running stack (Auth service → API → Next.js). Any misconfiguration in
 * the stack will surface here, not be swallowed.
 */

const AUTH_STATE_FILE = path.join(__dirname, '.auth/user.json');

const EMAIL = process.env.E2E_TEST_EMAIL || 'e2e-test@nixopus.test';
const PASSWORD = process.env.E2E_TEST_PASSWORD || 'Test@E2E!2026';

setup('seed test user', async ({ request }) => {
  /**
   * Register via the Next.js BFF proxy at /api/auth/sign-up/email
   * (same path the real UI uses; avoids direct CORS issues with the auth service).
   *
   * Accepted outcomes:
   *   200 / 201  — user created, continue
   *   409        — user already exists, continue
   *   Any other  — real error; fail loudly so the CI run surfaces it
   */
  const res = await request.post('/api/auth/sign-up/email', {
    data: {
      email: EMAIL,
      password: PASSWORD,
      name: 'E2E Test User'
    }
  });

  if (!res.ok()) {
    const body = await res.text();
    const alreadyExists =
      res.status() === 409 ||
      body.toLowerCase().includes('already') ||
      body.toLowerCase().includes('exists');

    if (!alreadyExists) {
      throw new Error(
        `Seed failed — could not register test user.\n` +
          `Status: ${res.status()}\n` +
          `Body: ${body}\n\n` +
          `Make sure the full stack is running and PASSWORD_LOGIN_ENABLED=true.`
      );
    }
    // User already exists — that is fine, move on.
  }
});

setup('authenticate test user and save browser state', async ({ page }) => {
  await page.goto('/auth');

  // Wait for the login form to be ready (config endpoint must resolve first)
  await expect(page.locator('input[type="email"]')).toBeVisible({ timeout: 15_000 });

  await page.locator('input[type="email"]').fill(EMAIL);
  await page.locator('input[type="password"]').fill(PASSWORD);
  await page.getByRole('button', { name: 'Login' }).click();

  // Real redirect — no mocking. If the stack is misconfigured this will time out.
  await page.waitForURL(/\/chats/, { timeout: 20_000 });

  // Persist cookies + localStorage so authenticated tests can skip login.
  await page.context().storageState({ path: AUTH_STATE_FILE });
});
