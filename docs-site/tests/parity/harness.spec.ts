import {test, expect} from '@playwright/test';
import {mkdirSync, writeFileSync} from 'node:fs';
import path from 'node:path';

const target = process.env.DOCS_URL || 'http://127.0.0.1:4182/terraform-provider-prisma-airs/';
const reference = 'http://127.0.0.1:4181/terraform-provider-prisma-airs/';
const viewports = [
  {name: 'desktop', width: 1440, height: 1000},
  {name: 'tablet', width: 1024, height: 768},
  {name: 'mobile', width: 390, height: 844},
];

for (const viewport of viewports) {
  for (const route of ['', 'getting-started/', 'guides/migration/']) {
    test(`${viewport.name} ${route || 'home'} matches harness design pixels`, async ({browser}, info) => {
      const context = await browser.newContext({viewport: {width: viewport.width, height: viewport.height}});
      const baseline = await context.newPage();
      const actual = await context.newPage();
      for (const [page, base] of [[baseline, reference], [actual, target]] as const) {
        await page.goto(new URL(route, base).href);
        await page.evaluate(() => document.fonts.ready);
        const families = await page.evaluate(() => [...document.fonts].filter(font => font.status === 'loaded').map(font => font.family));
        expect(families).toEqual(expect.arrayContaining(['Inter', 'JetBrains Mono']));
        await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
      }
      const options = {fullPage: true, animations: 'disabled' as const, caret: 'hide' as const};
      const expectedImage = await baseline.screenshot(options);
      const actualImage = await actual.screenshot(options);
      const name = `${viewport.name}-${route.replaceAll('/', '-') || 'home'}.png`;
      const expectedPath = info.snapshotPath(name);
      mkdirSync(path.dirname(expectedPath), {recursive: true});
      writeFileSync(expectedPath, expectedImage);
      await info.attach('harness-reference', {body: expectedImage, contentType: 'image/png'});
      await info.attach('terraform-site', {body: actualImage, contentType: 'image/png'});
      expect(actualImage).toMatchSnapshot(name, {maxDiffPixels: 0, threshold: 0});
      await context.close();
    });
  }
}
