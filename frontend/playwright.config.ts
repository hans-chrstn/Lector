import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  workers: 1,
  projects: [
    { name: 'development', use: { baseURL: 'http://127.0.0.1:4173' } },
    { name: 'production', use: { baseURL: 'http://127.0.0.1:4300' } },
  ],
  use: {
    baseURL: 'http://127.0.0.1:4173',
    trace: 'retain-on-failure',
    launchOptions: {
      executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH,
      args: ['--no-sandbox'],
    },
  },
  webServer: [
    {
      command: 'bash ../scripts/e2e-server.sh',
      url: 'http://127.0.0.1:4300/api/health',
      reuseExistingServer: false,
      timeout: 120000,
    },
    {
      command: 'LECTOR_API_TARGET=http://127.0.0.1:4300 npm run dev -- --port 4173 --strictPort',
      url: 'http://127.0.0.1:4173',
      reuseExistingServer: false,
    },
  ],
});
