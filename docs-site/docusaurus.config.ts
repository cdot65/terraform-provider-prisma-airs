import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import airsTheme from './src/css/prism-airs';

const config: Config = {
  title: 'Prisma AIRS Terraform',
  tagline: 'Infrastructure as code for Prisma AIRS security',
  favicon: 'img/brand-logo.png',
  url: 'https://cdot65.github.io',
  baseUrl: '/terraform-provider-prisma-airs/',
  organizationName: 'cdot65',
  projectName: 'terraform-provider-prisma-airs',
  future: {v4: true},
  trailingSlash: true,
  onBrokenLinks: 'throw',
  onBrokenAnchors: 'throw',
  markdown: {format: 'detect', mermaid: true, hooks: {onBrokenMarkdownLinks: 'throw'}},
  themes: ['@docusaurus/theme-mermaid'],
  i18n: {defaultLocale: 'en', locales: ['en']},
  presets: [['classic', {
    docs: {path: './docs', sidebarPath: './sidebars.ts', routeBasePath: '/'},
    blog: false,
    theme: {customCss: './src/css/custom.css'},
  } satisfies Preset.Options]],
  themeConfig: {
    mermaid: {theme: {light: 'dark', dark: 'dark'}, options: {themeVariables: {
      background: '#030609', primaryColor: '#061b29', primaryTextColor: '#f5f8fa',
      primaryBorderColor: '#00ddf2', lineColor: '#8999a6', secondaryColor: '#0b293b', tertiaryColor: '#061b29',
    }}},
    docs: {sidebar: {hideable: true}},
    colorMode: {defaultMode: 'dark', disableSwitch: true, respectPrefersColorScheme: false},
    navbar: {
      title: 'Prisma AIRS Terraform',
      logo: {alt: 'Prisma AIRS Terraform', src: 'img/brand-logo.png'},
      items: [
        {type: 'docSidebar', sidebarId: 'docs', label: 'Docs', position: 'left'},
        {type: 'docSidebar', sidebarId: 'reference', label: 'Provider Reference', position: 'left'},
        {type: 'docSidebar', sidebarId: 'developers', label: 'Developers', position: 'left'},
        {href: 'https://github.com/cdot65/terraform-provider-prisma-airs', label: 'GitHub', position: 'right'},
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {title: 'Terraform Provider', items: [
          {label: 'Getting Started', to: '/getting-started/'},
          {label: 'Provider Reference', to: '/reference/'},
          {label: 'Releases', to: '/about/release-notes/'},
        ]},
        {title: 'Prisma AIRS', items: [
          {label: 'CLI', href: 'https://cdot65.github.io/prisma-airs-cli/'},
          {label: 'TypeScript SDK', href: 'https://cdot65.github.io/prisma-airs-sdk/'},
          {label: 'Harness', href: 'https://cdot65.github.io/prisma-airs-harness/'},
        ]},
        {title: 'Source', items: [
          {label: 'GitHub · issues and contributions', href: 'https://github.com/cdot65/terraform-provider-prisma-airs/issues'},
          {label: 'Go SDK', href: 'https://cdot65.github.io/prisma-airs-go/'},
        ]},
      ],
      copyright: `Copyright © ${new Date().getFullYear()} cdot65. Terraform Provider: MIT. Built with Docusaurus.`,
    },
    prism: {theme: airsTheme, darkTheme: airsTheme,
      additionalLanguages: ['hcl', 'go', 'bash', 'json', 'yaml', 'python', 'powershell', 'toml', 'diff', 'rust']},
  } satisfies Preset.ThemeConfig,
};
export default config;
