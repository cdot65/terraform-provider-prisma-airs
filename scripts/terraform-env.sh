#!/usr/bin/env bash
# Loads .env file and runs terraform with those environment variables.
# Usage: ./scripts/terraform-env.sh [terraform args...]
#
# Looks for .env in the current directory first, then the repo root.

set -euo pipefail

load_env() {
    local envfile="$1"
    if [[ -f "$envfile" ]]; then
        # Export each non-comment, non-empty line
        set -a
        # shellcheck disable=SC1090
        source "$envfile"
        set +a
        echo "Loaded env from $envfile" >&2
        return 0
    fi
    return 1
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Try local .env first, then repo root
load_env ".env" || load_env "$REPO_ROOT/.env" || {
    echo "Warning: no .env file found. Using existing environment." >&2
}

exec terraform "$@"
