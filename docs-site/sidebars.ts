import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const resources = ['security-profile', 'custom-topic', 'api-key', 'customer-app', 'model-security-group', 'red-team-target', 'red-team-custom-prompt-set'];
const dataSources = ['dlp-profiles', 'deployment-profiles', 'model-security-rules'];
const generated = ['security_profile', 'custom_topic', 'api_key', 'customer_app', 'model_security_group', 'red_team_target', 'red_team_custom_prompt_set', 'dlp_profiles', 'deployment_profiles', 'model_security_rules'];

const sidebars: SidebarsConfig = {
  docs: [
    'index',
    {type: 'category', label: 'Getting started', collapsed: false, items: ['getting-started/index', 'getting-started/authentication', 'getting-started/installation', 'getting-started/configuration', 'getting-started/quick-start']},
    {type: 'category', label: 'Guides', items: ['guides/managing-security-profiles', 'guides/model-security-workflow', 'guides/red-team-testing', 'guides/import-and-state', 'guides/migration', 'guides/authentication', 'guides/troubleshooting']},
    {type: 'category', label: 'Examples', items: ['examples/index', 'examples/runtime-policy', 'examples/native-targets', 'examples/model-security', 'examples/repository-configurations']},
    {type: 'category', label: 'Resources', items: resources.map(name => ({type: 'ref' as const, id: `resources/${name}`}))},
    {type: 'category', label: 'Data sources', items: dataSources.map(name => ({type: 'ref' as const, id: `data-sources/${name}`}))},
    {type: 'category', label: 'About', items: ['about/release-notes', 'about/license']},
  ],
  reference: [
    'reference/index', 'reference/provider-configuration', 'reference/environment-variables',
    {type: 'category', label: 'Resource lifecycle', items: resources.map(name => `resources/${name}`)},
    {type: 'category', label: 'Data sources', items: dataSources.map(name => `data-sources/${name}`)},
    {type: 'category', label: 'Exact schemas', items: ['reference/generated/provider', ...generated.map(name => `reference/generated/prisma-airs_${name}`)]},
    'reference/error-handling',
  ],
  developers: ['development/architecture', 'development/documentation', 'development/design-parity', 'development/sdk-upgrade-verification'],
};
export default sidebars;
