import { defineConfig, devices } from '@playwright/test';

const isCI = !!process.env.CI;
const isNightly = !!process.env.NIGHTLY;
const baseURL = process.env.PLAYWRIGHT_TEST_BASE_URL || 'https://odyssey-beta-nine.vercel.app';
const isLocal = baseURL.includes('localhost');

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers: 1,
  reporter: isCI ? [['html'], ['github']] : 'html',
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    ...(isNightly
      ? [
          {
            name: 'firefox',
            use: { ...devices['Desktop Firefox'] },
          },
          {
            name: 'webkit',
            use: { ...devices['Desktop Safari'] },
          },
        ]
      : []),
  ],
  ...(isLocal
    ? {
        webServer: {
          command: 'npm run dev',
          url: 'http://localhost:5173',
          reuseExistingServer: !isCI,
          timeout: 120 * 1000,
        },
      }
    : {}),
});

