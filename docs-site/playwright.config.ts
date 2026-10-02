import {defineConfig} from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  testMatch: 'navigation.spec.ts',
  use: {
    baseURL: process.env.DOCS_URL || 'http://127.0.0.1:4173/terraform-provider-prisma-airs/',
    launchOptions: {executablePath: process.env.CHROMIUM_PATH || undefined},
  },
  webServer: process.env.DOCS_URL ? undefined : {
    command: 'npm run serve -- --host 127.0.0.1 --port 4173',
    url: 'http://127.0.0.1:4173/terraform-provider-prisma-airs/',
    reuseExistingServer: !process.env.CI,
  },
});
