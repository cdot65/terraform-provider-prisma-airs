import {test, expect} from '@playwright/test';

const resources = ['security-profile', 'custom-topic', 'api-key', 'customer-app', 'model-security-group', 'red-team-target', 'red-team-custom-prompt-set'];
const dataSources = ['dlp-profiles', 'deployment-profiles', 'model-security-rules'];
const routes = [
  'overview/', 'getting-started/', ...['installation', 'configuration', 'quick-start', 'authentication'].map(name => `getting-started/${name}/`),
  ...resources.map(name => `resources/${name}/`), ...dataSources.map(name => `data-sources/${name}/`),
  ...['authentication', 'managing-security-profiles', 'model-security-workflow', 'red-team-testing', 'migration', 'import-and-state', 'troubleshooting'].map(name => `guides/${name}/`),
  'examples/', ...['runtime-policy', 'native-targets', 'model-security', 'repository-configurations'].map(name => `examples/${name}/`),
  'reference/', ...['provider-configuration', 'environment-variables', 'error-handling'].map(name => `reference/${name}/`),
  'reference/generated/provider/', ...[...resources, ...dataSources].map(name => `reference/generated/prisma-airs_${name.replaceAll('-', '_')}/`),
  ...['architecture', 'documentation', 'design-parity', 'sdk-upgrade-verification'].map(name => `development/${name}/`),
  'about/release-notes/', 'about/license/',
];

test('all old and new guide routes have correct canonical URLs', async ({request, baseURL}) => {
  for (const route of routes) {
    const response = await request.get(new URL(route, baseURL).href);
    expect(response.status(), route).toBe(200);
    const html = await response.text();
    expect(html, route).toContain('<h1');
    const canonical = html.match(/<link\b[^>]*rel=["']?canonical["']?[^>]*href=["']?([^"' >]+)/)?.[1];
    expect(canonical, route).toBe(`https://cdot65.github.io/terraform-provider-prisma-airs/${route}`);
  }
});

test('hero and all homepage paths reach provider-specific guides', async ({page}) => {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  await page.goto('./');
  await expect(page.locator('#hero-title')).toHaveText('Policy as code.Security under control.');
  await expect(page.locator('main > section').nth(1).locator('a')).toHaveCount(4);
  await expect(page.locator('.theme-doc-sidebar-container')).toHaveCount(0);
  await expect(page.getByRole('img', {name: 'Prisma AIRS shield and prism spectrum'})).toBeVisible();
  await page.getByRole('link', {name: 'Get started →', exact: true}).click();
  await expect(page.getByRole('heading', {name: 'Getting started', exact: true})).toBeVisible();
  await expect(page.locator('pre.language-hcl').first()).toContainText('prisma-airs_security_profile');
  await expect(page.locator('.theme-admonition').filter({hasText: 'Updated provider'})).toBeVisible();
  expect(errors).toEqual([]);
});

test('HCL highlights, callouts, architecture diagram, and authentication guide render', async ({page, request}) => {
  await page.goto('examples/native-targets/');
  await expect(page.locator('pre.language-hcl')).toContainText('response_stop_value');
  expect(await page.locator('pre.language-hcl .token').count()).toBeGreaterThan(10);
  await expect(page.getByRole('button', {name: /Copy code to clipboard/})).toBeAttached();
  await page.goto('resources/api-key/');
  await expect(page.locator('.theme-admonition').filter({hasText: 'only available after creation'})).toBeVisible();
  await page.goto('development/architecture/');
  await expect(page.locator('.docusaurus-mermaid-container svg')).toBeVisible();
  await page.goto('getting-started/authentication/');
  await expect(page.getByRole('link', {name: 'service-account creation guide'})).toBeVisible();
  await expect(page.locator('article')).toContainText('PANW_MGMT_CLIENT_SECRET');
  const notice = await request.get('licenses/NOTICE.txt');
  expect(notice.status()).toBe(200);
  expect(await notice.text()).toContain('Prisma AIRS Terraform documentation adaptation');
});

test('desktop articles use the shared reading width and mobile-only contents', async ({page}) => {
  await page.setViewportSize({width: 1440, height: 1000});
  await page.goto('getting-started/');
  await expect(page.locator('.theme-doc-toc-desktop')).toHaveCount(0);
  await expect(page.locator('.theme-doc-toc-mobile')).toBeHidden();
  expect(await page.locator('article').evaluate(element => element.getBoundingClientRect().width)).toBeGreaterThan(800);
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
});

test('reference catalog exposes every native family and sensitive fields', async ({page}) => {
  await page.goto('resources/red-team-target/');
  await expect(page.locator('.navbar__link--active')).toHaveText('Provider Reference');
  await page.goto('reference/generated/prisma-airs_red_team_target/');
  for (const family of ['openai', 'hugging_face', 'databricks', 'bedrock', 'custom', 'rest', 'streaming']) {
    await expect(page.locator(`h2#${family}`)).toBeVisible();
  }
  await page.goto('reference/generated/prisma-airs_deployment_profiles/');
  for (const field of ['profile_id', 'auth_code', 'details']) {
    await expect(page.locator('tr').filter({has: page.locator(`td:first-child code:text-is("${field}")`)})).toContainText('yes');
  }
});

for (const viewport of [{width: 1440, height: 1000}, {width: 1024, height: 768}, {width: 390, height: 844}]) {
  test(`homepage links and responsive pages work at ${viewport.width}px`, async ({page}) => {
    await page.setViewportSize(viewport);
    await page.goto('./');
    const links = await page.locator('main a').evaluateAll(elements => elements.map(element => (element as HTMLAnchorElement).href));
    for (const href of links) {
      const response = await page.goto(href);
      expect(response?.status()).toBe(200);
      await expect(page.locator('article h1')).toBeVisible();
    }
    await page.goto('reference/generated/prisma-airs_red_team_target/');
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(viewport.width);
    if (viewport.width === 390) {
      await expect(page.locator('.theme-doc-toc-mobile')).toBeVisible();
      await page.getByRole('button', {name: 'Toggle navigation bar'}).click();
      await expect(page.locator('.navbar-sidebar')).toBeVisible();
    }
  });
}
