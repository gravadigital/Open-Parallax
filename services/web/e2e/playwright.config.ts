import { defineConfig, devices } from '@playwright/test';

// Runs against the product already up (`make test-e2e` at the root brings it up with
// the test overrides). WEB_URL is the web as the browser sees it.
export default defineConfig({
  testDir: './tests',
  timeout: 60_000,
  fullyParallel: false,
  forbidOnly: true,
  retries: 0,
  reporter: [
    ['list'],
    ['junit', { outputFile: process.env.PLAYWRIGHT_JUNIT_OUTPUT_FILE ?? '../../../reports/junit/web-e2e.xml' }],
  ],
  use: {
    baseURL: process.env.WEB_URL ?? 'http://localhost:3000',
    locale: 'es-AR',
    trace: 'retain-on-failure',
  },
  outputDir: '/tmp/playwright-results',
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
