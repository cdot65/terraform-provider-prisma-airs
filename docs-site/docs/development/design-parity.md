# Shared AIRS design

The Terraform site follows the Prisma AIRS Harness and Go SDK design: identical logo artwork, Inter and JetBrains Mono fonts, obsidian/navy surfaces, cyan accents, two-column hero, numbered path cards, quick-links band, navbar, footer, and article width.

The source is [Harness commit 1885e40](https://github.com/cdot65/prisma-airs-harness/tree/1885e40eb1dc493ad5b47757694d077011afa431/docs-site). `docs-site/design/harness/source.json` records the full commit and hashes. The archive preserves independent source and dependency inputs. Terraform product text, links, and HCL syntax support are listed explicitly in `copy.json`.

## Verify rendering

```bash
make docs-check
```

Nine screenshots compare the homepage, getting-started guide, and migration guide at 1440px, 1024px, and 390px widths against an independent reference build. Both builds render the same Terraform content, so the zero-pixel-difference check measures design rather than wording differences between products. Evidence is saved to `docs-site/test-results/`.

The copied source files must match their pinned hashes. Mobile navigation and contents remain available; desktop articles use the same wide reading column as the Harness.

Reused Harness components retain their Apache-2.0 attribution in `docs-site/NOTICE` and `docs-site/THIRD_PARTY_NOTICES.md`. The provider remains MIT licensed.

## Scope and dependencies

The comparison verifies the pinned shared design with identical Terraform content. It does not compare unrelated product copy or live deployed sites. The reference navigation and guide text are shared inputs; the renderer, CSS, hero/article layout, and primary artwork come from the archived Harness source and its own dependency lockfile. Fonts load from Google Fonts as in the source design, so the rendering checks need network access to load both font families.
