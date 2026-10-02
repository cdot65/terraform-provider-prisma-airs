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
        presence = ', '.join(flag for flag in ['required', 'optional', 'computed'] if row.get(flag))
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


def registry_render(name, schema, kind):
    short = name.removeprefix('prisma-airs_')
    folder = 'resources' if kind == 'resource' else 'data-sources'
    guide = SITE / 'docs' / folder / (short.replace('_', '-') + '.md')
    text = registry_content(guide)
    text = re.sub(r'^# .*\n', '', text, count=1)
    heading = f'# {name} {kind.title()}'
    lines = [f'---\npage_title: "{name} ({kind.title()})"\n---', '', heading, '', text.strip(), '', '## Schema', '']
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
    expected = {'provider.md': render('Provider', schema['provider'], 'provider')}
    for group, kind in [('resource_schemas', 'resource'), ('data_source_schemas', 'data source')]:
        for name, row in sorted(schema[group].items()):
            expected[name + '.md'] = render(name, row, kind)
    registry = {}
    for group, kind, folder in [('resource_schemas', 'resource', 'resources'), ('data_source_schemas', 'data source', 'data-sources')]:
        for name, row in sorted(schema[group].items()):
            registry[ROOT / 'docs' / folder / (name.removeprefix('prisma-airs_') + '.md')] = registry_render(name, row, kind)
    for guide in sorted((SITE / 'docs/guides').glob('*.md')):
        title = next(line.removeprefix('# ') for line in guide.read_text().splitlines() if line.startswith('# '))
        registry[ROOT / 'docs/guides' / guide.name] = '---\npage_title: ' + json.dumps(title) + '\n---\n\n' + registry_content(guide).strip() + '\n'
    registry[ROOT / 'docs/index.md'] = '---\npage_title: "Prisma AIRS Provider"\n---\n\n' + registry_content(SITE / 'docs/index.md').strip() + '\n'
    if check:
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
