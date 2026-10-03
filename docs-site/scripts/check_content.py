"""Check Markdown conventions and validate complete HCL examples without API access."""
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

SITE = Path(__file__).resolve().parents[1]
ROOT = SITE.parent
COMPLETE = ['getting-started/index.md', 'examples/runtime-policy.md', 'examples/native-targets.md', 'examples/model-security.md', 'examples/gateway.md']


def main():
    errors = []
    for path in sorted((SITE / 'docs').rglob('*.md')):
        text = path.read_text()
        if re.search(r'^!!!|^===|^:::[a-z]+ ', text, re.MULTILINE):
            errors.append(f'{path.relative_to(ROOT)}: legacy Markdown extension')
    if errors:
        raise SystemExit('\n'.join(errors))
    env = {key: value for key, value in os.environ.items() if not key.startswith('PANW_')}
    with tempfile.TemporaryDirectory(prefix='airs-docs-examples-') as directory:
        base = Path(directory)
        config = base / 'dev.tfrc'
        config.write_text('provider_installation {\n  dev_overrides {\n    "cdot65/prisma-airs" = ' + json.dumps(str(ROOT)) + '\n  }\n  direct {}\n}\n')
        env.update(TF_CLI_CONFIG_FILE=str(config), TF_IN_AUTOMATION='1')
        validated = 0
        skipped = []
        for path in sorted((SITE / 'docs').rglob('*.md')):
            relative = str(path.relative_to(SITE / 'docs'))
            blocks = re.findall(r'```hcl\n(.*?)\n```', path.read_text(), re.DOTALL)
            if relative in COMPLETE and len(blocks) != 1:
                errors.append(f'{relative}: expected one complete HCL configuration')
            for index, block in enumerate(blocks):
                # CLI configuration and explicitly partial nested blocks are not Terraform roots.
                if block.lstrip().startswith('provider_installation') or (relative == 'guides/managing-security-profiles.md' and block.lstrip().startswith('app_protection')):
                    skipped.append(f'{relative}#{index}')
                    continue
                text = block
                if 'required_providers' not in text:
                    text = 'terraform {\n  required_providers {\n    prisma-airs = { source = "cdot65/prisma-airs" }\n  }\n}\n' + text
                references = set(re.findall(r'\bvar\.([a-zA-Z_][a-zA-Z_0-9]*)', text))
                declarations = set(re.findall(r'variable\s+"([^"]+)"', text))
                for name in sorted(references - declarations):
                    text += '\nvariable "' + name + '" {\n  type = string\n  sensitive = true\n}\n'
                work = base / str(validated)
                work.mkdir()
                (work / 'main.tf').write_text(text + '\n')
                result = subprocess.run(['terraform', 'validate', '-json'], cwd=work, env=env, capture_output=True, text=True, check=False)
                report = json.loads(result.stdout)
                if result.returncode or not report['valid']:
                    errors.append(f'{relative}#{index}: ' + json.dumps(report['diagnostics']))
                validated += 1
    if errors:
        raise SystemExit('\n'.join(errors))
    print(f'Checked Markdown conventions and validated {validated} HCL examples, including {len(COMPLETE)} complete configurations, with no live credentials. Explicit non-root fragments: {len(skipped)}.')


if __name__ == '__main__':
    main()
