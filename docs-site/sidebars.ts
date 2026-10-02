import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

import catalog from './product-catalog.json';

type ProductItem = string | {type: 'ref'; id: string};
const productItems = (reference: boolean) => catalog.map(product => ({
  type: 'category' as const,
  label: product.label,
  items: product.implemented
    ? [...product.resources, ...product.data_sources].flatMap<ProductItem>(entry => reference
      ? [entry.guide, `reference/generated/${entry.name}`]
      : [{type: 'ref' as const, id: entry.guide}])
    : [reference ? 'products/gateway' : {type: 'ref' as const, id: 'products/gateway'}],
}));

const sidebars: SidebarsConfig = {
  docs: [
    'index',
    {type: 'category', label: 'Getting started', collapsed: false, items: ['getting-started/index', 'getting-started/authentication', 'getting-started/installation', 'getting-started/configuration', 'getting-started/quick-start']},
    {type: 'category', label: 'Guides', items: ['guides/managing-security-profiles', 'guides/model-security-workflow', 'guides/red-team-testing', 'guides/import-and-state', 'guides/migration', 'guides/authentication', 'guides/troubleshooting']},
    {type: 'category', label: 'Examples', items: ['examples/index', 'examples/runtime-policy', 'examples/native-targets', 'examples/model-security', 'examples/repository-configurations']},
    ...productItems(false),
    {type: 'category', label: 'About', items: ['about/release-notes', 'about/license']},
  ],
  reference: [
    'reference/index', 'reference/provider-configuration', 'reference/environment-variables',
    'reference/generated/provider',
    ...productItems(true),
    'reference/error-handling',
  ],
  developers: ['development/architecture', 'development/documentation', 'development/design-parity', 'development/sdk-upgrade-verification', 'development/product-refactor-verification'],
};
export default sidebars;
