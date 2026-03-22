#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# Sandbox setup — build provider and configure dev_overrides
# ---------------------------------------------------------------------------
# Run this once before using terraform plan/apply/destroy:
#   source ./setup.sh
# ---------------------------------------------------------------------------
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SANDBOX_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Building provider..."
cd "$REPO_ROOT"
make build

# Create dev_overrides config
cat > "$SANDBOX_DIR/.dev.tfrc" <<EOF
provider_installation {
  dev_overrides {
    "cdot65/prisma-airs" = "$REPO_ROOT"
  }
  direct {}
}
EOF

export TF_CLI_CONFIG_FILE="$SANDBOX_DIR/.dev.tfrc"

# Load credentials from .env if available
if [[ -f "$REPO_ROOT/.env" ]]; then
  echo "Loading credentials from .env"
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# || ! "$line" =~ = ]] && continue
    export "$line"
  done < "$REPO_ROOT/.env"
fi

cd "$SANDBOX_DIR"
echo ""
echo "Ready! You can now run:"
echo "  terraform plan"
echo "  terraform apply"
echo "  terraform destroy"
echo ""
echo "TF_CLI_CONFIG_FILE=$TF_CLI_CONFIG_FILE"
