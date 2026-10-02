# Documentation development

The site uses Docusaurus 3.10.2 and Node 24+. Authored Markdown lives in `docs-site/docs/`; UI, build configuration, and browser checks live in `docs-site/`. Public URLs use the `/terraform-provider-prisma-airs/` base path.

## Run locally

```bash
npm ci --prefix docs-site
make docs-serve
```

## Change resource documentation

Build the provider and refresh the exact schema catalog after changing a schema:

```bash
make generate
make docs-check
```

`make generate` builds the provider before running the safe schema generator through `go generate`. It writes exact schema pages in `docs-site/docs/reference/generated/` and the Registry documentation in `docs/`; authored guides remain separate. Direct `go generate` requires an already-built current provider.

The schema generator uses Terraform's provider-schema protocol with a temporary development override. It needs Terraform and the built provider; it performs no live API calls. `--check` detects missing, extra, or stale generated pages.

Examples described as complete must include a provider requirement and all variable declarations. Add complete examples to the offline Terraform validation fixtures when extending the guide catalog. Live lifecycle evidence remains in [SDK upgrade verification](sdk-upgrade-verification.md).

## Maintain the shared design

The site uses the owner-provided Terraform logo and copies the Harness CSS, Prism colors, hero layout, and article layout. Product text and routes are explicit substitutions in `docs-site/design/harness/copy.json`. The source archive and hashes pin the reference; refresh it deliberately when the shared design changes.

`npm run check` checks design inputs, content, schema freshness, TypeScript, production build, browser navigation, and nine pixel comparisons. The independent reference uses its own locked Docusaurus 3.10.1 dependencies and the same Terraform copy. Screenshots compare homepage, getting-started, and migration pages at desktop, tablet, and mobile sizes with zero differing pixels. Both sites must load Inter and JetBrains Mono.

## Deployment

`.github/workflows/deploy-docs.yml` runs checks for documentation pull requests and main-branch changes, uploads browser evidence, and publishes `docs-site/build` through GitHub Pages. A post-deployment browser check verifies the public site. Local completion does not imply deployment; publication requires committed changes on the deploy branch.

Builds fail on broken links or anchors. Standard Docusaurus callouts and Mermaid diagrams replace the former site-specific Markdown extensions.

Registry resource and data-source pages use underscore filenames under `docs/resources/` and `docs/data-sources/`. The schema generator produces them from the lifecycle guides and exact schemas. These generated pages are separate from the authored Docusaurus guides. Release tags publish these pages to the Terraform Registry. GitHub Actions builds, checks, deploys, and tests the public Pages site after each relevant main-branch update.
