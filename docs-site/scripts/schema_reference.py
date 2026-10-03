"""Generate exact reference pages from the built provider's Terraform schema."""
import argparse
import json
import os
import re
from pathlib import Path
import subprocess
import tempfile

SITE = Path(__file__).resolve().parents[1]
ROOT = SITE.parent
DESTINATION = SITE / 'docs/reference/generated'
CATALOG = SITE / 'product-catalog.json'
INVENTORY = SITE / 'docs/reference/index.md'


def provider_schema():
    binary = ROOT / 'terraform-provider-prisma-airs'
    if not binary.is_file():
        raise SystemExit('Build the provider with make build before generating schema documentation.')
    with tempfile.TemporaryDirectory(prefix='airs-docs-schema-') as directory:
        work = Path(directory)
        config = work / 'dev.tfrc'
        config.write_text('provider_installation {\n  dev_overrides {\n    "cdot65/prisma-airs" = ' + json.dumps(str(ROOT)) + '\n  }\n  direct {}\n}\n')
        (work / 'main.tf').write_text('terraform {\n  required_providers {\n    prisma-airs = { source = "cdot65/prisma-airs" }\n  }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(config), TF_IN_AUTOMATION='1')
        result = subprocess.run(['terraform', 'providers', 'schema', '-json'], cwd=work, env=env, capture_output=True, text=True, check=True)
        return json.loads(result.stdout)['provider_schemas']['registry.terraform.io/cdot65/prisma-airs']


def product_catalog(schema):
    result = subprocess.run(['go', 'run', './cmd/product-catalog'], cwd=ROOT,
                            env=dict(os.environ, GOPRIVATE='github.com/cdot65/*'),
                            capture_output=True, text=True, check=True)
    catalog = json.loads(result.stdout)
    owned = {'resource_schemas': {}, 'data_source_schemas': {}}
    ids = set()
    for product in catalog:
        if product['id'] in ids:
            raise SystemExit('Duplicate product ID: ' + product['id'])
        ids.add(product['id'])
        if not product['implemented'] and (product['resources'] or product['data_sources'] or product['endpoints']):
            raise SystemExit('Unimplemented products cannot register types or settings.')
        for key, group in [('resources', 'resource_schemas'), ('data_sources', 'data_source_schemas')]:
            for entry in product[key]:
                name = entry['name']
                if name in owned[group]:
                    raise SystemExit('Duplicate product ownership: ' + name)
                guide = SITE / 'docs' / (entry['guide'] + '.md')
                if not guide.is_file():
                    raise SystemExit('Missing lifecycle guide: ' + str(guide))
                owned[group][name] = (entry, product)
    for group, entries in owned.items():
        if set(entries) != set(schema[group]):
            raise SystemExit('Product catalog and built schema disagree: ' + group)
    return catalog, owned


def reference_name(name, kind, resource_names):
    # Resource and data-source type names may legally be identical. Preserve
    # existing resource URLs while giving the colliding read schema its own page.
    return 'data-source-' + name if kind == 'data source' and name in resource_names else name


def inventory(catalog):
    lines = ['---', 'title: Provider reference', 'slug: /reference', '---', '',
             'Product ownership is generated from the provider registrations. Lifecycle guides explain imports and updates; exact schemas list types, nested blocks, and sensitive fields.', '',
             'See [provider configuration](provider-configuration.md), [environment variables](environment-variables.md), and the [provider schema](generated/provider.md).', '']
    resource_names = {entry['name'] for product in catalog for entry in product['resources']}
    for product in catalog:
        lines += ['## ' + product['label'], '']
        if not product['implemented']:
            lines += ['Not yet implemented. There are no Gateway resources, data sources, or configuration settings in this release. See [AI Gateway status](../products/gateway.md).', '']
            continue
        lines += ['| Terraform type | Kind | Guide | Schema |', '| --- | --- | --- | --- |']
        for key, kind in [('resources', 'Resource'), ('data_sources', 'Data source')]:
            for entry in product[key]:
                lines.append(f"| `{entry['name']}` | {kind} | [Lifecycle](../{entry['guide']}.md) | [Attributes](generated/{reference_name(entry['name'], kind.lower(), resource_names)}.md) |")
        lines.append('')
    return '\n'.join(lines).rstrip() + '\n'


def type_name(value):
    if isinstance(value, str):
        return value
    kind, item = value
    if kind == 'object':
        return 'object'
    if kind == 'tuple':
        return 'tuple'
    return f'{kind}({type_name(item)})'


def description(text):
    return ' '.join(text.split()).replace('|', r'\|')


def attributes(rows, label, depth=2):
    lines = ['#' * depth + ' ' + label, '', '| Attribute | Type | Presence | Sensitive | Description |', '| --- | --- | --- | --- | --- |']
    for name, row in sorted(rows.items()):
        presence = ', '.join(flag for flag in ['required', 'optional', 'computed', 'write_only'] if row.get(flag))
        typ = type_name(row['type']) if 'type' in row else row['nested_type']['nesting_mode'] + '(object)'
        lines.append(f'| `{name}` | `{typ}` | {presence} | {"yes" if row.get("sensitive") else "—"} | {description(row.get("description", ""))} |')
    lines.append('')
    for name, row in sorted(rows.items()):
        if 'nested_type' in row:
            lines += attributes(row['nested_type']['attributes'], label + '.' + name, min(depth + 1, 6))
    return lines


def block_lines(block, label='Attributes', depth=2):
    lines = attributes(block.get('attributes', {}), label, depth) if block.get('attributes') else []
    for name, row in sorted(block.get('block_types', {}).items()):
        lines += ['#' * depth + ' ' + (name if label == 'Attributes' else label + '.' + name), '', f'Nesting: `{row["nesting_mode"]}`.']
        bounds = []
        if row.get('min_items'):
            bounds.append(f'Minimum: {row["min_items"]}.')
        if row.get('max_items'):
            bounds.append(f'Maximum: {row["max_items"]}.')
        lines += [' '.join(bounds), ''] if bounds else ['']
        lines += block_lines(row['block'], name if label == 'Attributes' else label + '.' + name, min(depth + 1, 6))
    return lines


def render(name, schema, kind):
    lines = [f'# {name} schema', '', f'Exact {kind} attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.', '', description(schema['block'].get('description', '')), '', 'Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.', '']
    lines += block_lines(schema['block'])
    return '\n'.join(lines).rstrip() + '\n'


def registry_render(name, schema, kind, entry, product):
    guide = SITE / 'docs' / (entry['guide'] + '.md')
    text = registry_content(guide)
    text = re.sub(r'^# .*\n', '', text, count=1)
    heading = f'# {name} {kind.title()}'
    lines = [f'---\npage_title: "{name} ({kind.title()})"\nsubcategory: "{product["label"]}"\n---', '', heading, '', text.strip(), '', '## Schema', '']
    lines += block_lines(schema['block'], depth=3)
    return '\n'.join(lines).rstrip() + '\n'


def registry_content(guide):
    text = guide.read_text()
    text = re.sub(r'^---\n.*?\n---\n', '', text, count=1, flags=re.DOTALL)
    text = re.sub(r'^:::[a-z]+(?:\[([^]\n]+)\])?\s*$', lambda m: '**' + (m.group(1) or 'Note') + '**', text, flags=re.MULTILINE)
    text = re.sub(r'^:::\s*$', '', text, flags=re.MULTILINE)
    return re.sub(r'(?<!!)\[([^]]+)\]\(([^)]+)\)', lambda m: registry_link(m, guide), text)


def registry_link(match, guide):
    label, link = match.groups()
    if link.startswith(('https://', 'http://', '#')):
        return match.group(0)
    relative, separator, anchor = link.partition('#')
    destination = (guide.parent / relative).resolve().relative_to((SITE / 'docs').resolve())
    route = str(destination).removesuffix('.md')
    if route.endswith('/index'):
        route = route.removesuffix('index')
    elif route == 'index':
        route = 'overview/'
    else:
        route += '/'
    return f'[{label}](https://cdot65.github.io/terraform-provider-prisma-airs/{route}' + (separator + anchor if separator else '') + ')'


def main():
    args = argparse.ArgumentParser()
    args.add_argument('--check', action='store_true')
    check = args.parse_args().check
    schema = provider_schema()
    catalog, owned = product_catalog(schema)
    derived = {CATALOG: json.dumps(catalog, indent=2) + '\n', INVENTORY: inventory(catalog)}
    expected = {'provider.md': render('Provider', schema['provider'], 'provider')}
    for group, kind in [('resource_schemas', 'resource'), ('data_source_schemas', 'data source')]:
        for name, row in sorted(schema[group].items()):
            expected[reference_name(name, kind, schema['resource_schemas']) + '.md'] = render(name, row, kind)
    registry = {}
    for group, kind, folder in [('resource_schemas', 'resource', 'resources'), ('data_source_schemas', 'data source', 'data-sources')]:
        for name, row in sorted(schema[group].items()):
            registry[ROOT / 'docs' / folder / (name.removeprefix('prisma-airs_') + '.md')] = registry_render(name, row, kind, *owned[group][name])
    for guide in sorted((SITE / 'docs/guides').glob('*.md')):
        title = next(line.removeprefix('# ') for line in guide.read_text().splitlines() if line.startswith('# '))
        registry[ROOT / 'docs/guides' / guide.name] = '---\npage_title: ' + json.dumps(title) + '\n---\n\n' + registry_content(guide).strip() + '\n'
    registry[ROOT / 'docs/index.md'] = '---\npage_title: "Prisma AIRS Provider"\n---\n\n' + registry_content(SITE / 'docs/index.md').strip() + '\n'
    if check:
        stale = [str(path.relative_to(ROOT)) for path, content in derived.items() if not path.is_file() or path.read_text() != content]
        if stale:
            raise SystemExit('Stale product catalog: ' + ', '.join(stale))
        actual_registry = set((ROOT / "docs").rglob("*.md"))
        if actual_registry != set(registry):
            raise SystemExit("Registry documentation has missing or extra pages. Run make generate and keep authored content in docs-site/docs/.")
        stale_registry = [str(path.relative_to(ROOT)) for path, content in registry.items() if not path.is_file() or path.read_text() != content]
        if stale_registry:
            raise SystemExit('Stale Registry pages: ' + ', '.join(stale_registry))
        print(f'Checked {len(registry)} Terraform Registry pages.')
        actual = {p.name for p in DESTINATION.glob('*.md')}
        if actual != set(expected):
            raise SystemExit('Generated schema catalog has missing or extra pages. Run npm run generate:reference.')
        stale = [name for name, text in expected.items() if (DESTINATION / name).read_text() != text]
        if stale:
            raise SystemExit('Stale schema pages: ' + ', '.join(stale))
        print(f'Checked {len(expected)} exact schema pages against the built provider.')
    else:
        for path, content in derived.items():
            path.write_text(content)
        for path in (ROOT / 'docs').rglob('*.md'):
            if path not in registry:
                path.unlink()
        for path, content in registry.items():
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content)
        print(f'Generated {len(registry)} Terraform Registry pages.')
        DESTINATION.mkdir(parents=True, exist_ok=True)
        for name, text in expected.items():
            (DESTINATION / name).write_text(text)
        for path in DESTINATION.glob('*.md'):
            if path.name not in expected:
                path.unlink()
        print(f'Generated {len(expected)} exact schema pages.')


if __name__ == '__main__':
    main()
