import { defineConfig, devices } from '@playwright/test';
import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(__dirname, '.env.test') });

/**
 * Playwright E2E configuration.
 *
 * Local usage:
 *   1. cp view/.env.test.example view/.env.test  (fill in your credentials)
 *   2. Start the full stack: docker compose -f docker-compose-test.yml up -d --wait
 *   3. Start the API:        cd api && go run main.go
 *   4. Build the view:       cd view && npm run build
 *   5. Run tests:            cd view && npx playwright test
 *
 * The webServer block starts `next start` automatically and stops it after tests.
 * Set reuseExistingServer=true (default for non-CI) to skip restart if already running.
 */
export default defineConfig({
  testDir: './tests/e2e',

  // Run tests sequentially — auth state is shared between some tests.
  fullyParallel: false,
  workers: 1,

  // Fail the build in CI if test.only() was accidentally committed.
  forbidOnly: !!process.env.CI,

  // Retry once in CI for flaky network conditions; no retries locally.
  retries: process.env.CI ? 1 : 0,

  reporter: [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]],

  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'on-first-retry'
  },

  projects: [
    // Setup project: seeds test user + saves authenticated storage state.
    // All other projects depend on this completing first.
    {
      name: 'setup',
      testMatch: /global\.setup\.ts/
    },
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup']
    }
  ],

  // Start the production Next.js server before any test runs.
  // In CI, `npm run build` is a separate workflow step run before `playwright test`.
  webServer: {
    command: 'npm run start',
    url: 'http://localhost:3000',
    // Locally: reuse an already-running server. In CI: always start a fresh one.
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
    env: {
      PORT: '3000',
      API_URL: process.env.API_URL || 'http://localhost:8080/api',
      AUTH_SERVICE_URL: process.env.AUTH_SERVICE_URL || 'http://localhost:9090',
      WEBSOCKET_URL: process.env.WEBSOCKET_URL || 'ws://localhost:8080/ws',
      PASSWORD_LOGIN_ENABLED: 'true',
      NEXT_PUBLIC_BASE_PATH: ''
    }
  }
});
