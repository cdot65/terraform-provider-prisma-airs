import {defineConfig} from '@playwright/test';

const reference = process.env.PARITY_REFERENCE_DIR;
if (!reference) throw new Error('Run npm run test:parity to build the independent reference first');

export default defineConfig({
  testDir: './tests/parity',
  snapshotPathTemplate: '{testDir}/../../test-results/reference/{testFilePath}/{arg}{ext}',
  use: {launchOptions: {executablePath: process.env.CHROMIUM_PATH || undefined}},
  webServer: [
    {command: 'node node_modules/@docusaurus/core/bin/docusaurus.mjs serve --host 127.0.0.1 --port 4181',
      cwd: reference, url: 'http://127.0.0.1:4181/terraform-provider-prisma-airs/', reuseExistingServer: false},
    ...(process.env.DOCS_URL ? [] : [{command: 'npm run serve -- --host 127.0.0.1 --port 4182',
      url: 'http://127.0.0.1:4182/terraform-provider-prisma-airs/', reuseExistingServer: false}]),
  ],
});
