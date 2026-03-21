#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# Prisma AIRS Provider — End-to-End Test Runner
#
# Usage:
#   ./e2e/run.sh              # full cycle: build → apply → validate → destroy
#   ./e2e/run.sh --no-destroy # skip destroy (leave resources for inspection)
#   ./e2e/run.sh --destroy    # destroy only (clean up a previous run)
# ---------------------------------------------------------------------------
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
E2E_DIR="$(cd "$(dirname "$0")" && pwd)"
NO_DESTROY=false
DESTROY_ONLY=false

for arg in "$@"; do
  case "$arg" in
    --no-destroy) NO_DESTROY=true ;;
    --destroy)    DESTROY_ONLY=true ;;
  esac
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

pass() { echo -e "  ${GREEN}PASS${NC}  $1"; }
fail() { echo -e "  ${RED}FAIL${NC}  $1"; FAILURES=$((FAILURES + 1)); }
info() { echo -e "${CYAN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}WARN${NC}  $1"; }

FAILURES=0

# ── Source credentials ──────────────────────────────────────────────────
if [[ -f "$REPO_ROOT/.env" ]]; then
  info "Loading credentials from .env"
  while IFS= read -r line || [[ -n "$line" ]]; do
    # skip comments, empty lines, and lines without =
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# || ! "$line" =~ = ]] && continue
    export "$line"
  done < "$REPO_ROOT/.env"
else
  echo -e "${RED}ERROR${NC}: $REPO_ROOT/.env not found"
  echo "Create .env with: PANW_AI_SEC_API_KEY, PANW_MGMT_CLIENT_ID, PANW_MGMT_CLIENT_SECRET, PANW_MGMT_TSG_ID"
  exit 1
fi

# ── Build & install provider ───────────────────────────────────────────
info "Building provider binary"
cd "$REPO_ROOT"
make build 2>&1 | tail -1

# Point Terraform at the local binary via dev_overrides
export TF_CLI_CONFIG_FILE="$E2E_DIR/.dev.tfrc"
cat > "$TF_CLI_CONFIG_FILE" <<EOF
provider_installation {
  dev_overrides {
    "cdot65/prisma-airs" = "$REPO_ROOT"
  }
  direct {}
}
EOF

cd "$E2E_DIR"

# ── Destroy-only mode ──────────────────────────────────────────────────
if $DESTROY_ONLY; then
  info "Destroying resources"
  terraform destroy -auto-approve
  rm -f "$TF_CLI_CONFIG_FILE"
  info "Done"
  exit 0
fi

# ── Plan ───────────────────────────────────────────────────────────────
info "Running terraform plan"
terraform plan -out=e2e.tfplan

# ── Apply ──────────────────────────────────────────────────────────────
info "Running terraform apply"
terraform apply e2e.tfplan
rm -f e2e.tfplan

# ── Validate Outputs ───────────────────────────────────────────────────
info "Validating outputs"
OUTPUTS=$(terraform output -json)

check_output() {
  local key="$1"
  local description="$2"
  local value
  value=$(echo "$OUTPUTS" | jq -r ".[\"$key\"].value // empty")

  if [[ -z "$value" || "$value" == "null" ]]; then
    fail "$description ($key): empty or missing"
  else
    pass "$description ($key) = $value"
  fi
}

check_output_gte() {
  local key="$1"
  local min="$2"
  local description="$3"
  local value
  value=$(echo "$OUTPUTS" | jq -r ".[\"$key\"].value // 0")

  if [[ "$value" -ge "$min" ]]; then
    pass "$description ($key) = $value (>= $min)"
  else
    fail "$description ($key) = $value (expected >= $min)"
  fi
}

echo ""
echo "─── Management Resources ────────────────────────────────────────"
check_output "security_profile_id"   "Security Profile ID"
check_output "security_profile_name" "Security Profile Name"
check_output "custom_topic_id"       "Custom Topic ID"
check_output "custom_topic_name"     "Custom Topic Name"
# api_key excluded — requires customer app with no existing key
# check_output "api_key_id"            "API Key ID"
# check_output "api_key_name"          "API Key Name"
# check_output "api_key_status"        "API Key Status"
# customer_app excluded — upstream API hangs on create
# check_output "customer_app_id"       "Customer App ID"
# check_output "customer_app_name"     "Customer App Name"

echo ""
echo "─── Management Data Sources ─────────────────────────────────────"
check_output_gte "dlp_profile_count"        0 "DLP Profiles"
check_output_gte "deployment_profile_count" 0 "Deployment Profiles"
# scan_log_count excluded — upstream API endpoint is unreliable
# check_output_gte "scan_log_count"           0 "Scan Logs"

echo ""
echo "─── Model Security Resources ────────────────────────────────────"
check_output "model_security_group_id"    "Security Group ID"
check_output "model_security_group_name"  "Security Group Name"
check_output "model_security_group_state" "Security Group State"
check_output "model_scan_id"              "Model Scan ID"
check_output "model_scan_eval_outcome"    "Model Scan Eval Outcome"

echo ""
echo "─── Model Security Data Sources ─────────────────────────────────"
check_output_gte "model_security_rule_count"    0 "Security Rules"
check_output_gte "model_scan_evaluation_count"  0 "Scan Evaluations"
check_output_gte "model_scan_violation_count"   0 "Scan Violations"

echo ""
echo "─── Red Team Resources ──────────────────────────────────────────"
check_output "red_team_prompt_set_id"     "Custom Prompt Set ID"
check_output "red_team_prompt_set_status" "Custom Prompt Set Status"
check_output "red_team_target_id"         "Target ID"
check_output "red_team_target_status"     "Target Status"
# red_team_scan excluded — requires validated target (httpbin can't validate)
# check_output "red_team_scan_id"           "Scan Job ID"
# check_output "red_team_scan_status"       "Scan Job Status"

echo ""
echo "─── Red Team Data Sources ───────────────────────────────────────"
check_output_gte "red_team_category_count" 1 "Attack Categories"
check_output     "red_team_quota"            "Quota Details"

echo ""
echo "─── Scan API ────────────────────────────────────────────────────"
check_output "content_scan_benign_category"    "Benign Scan Category"
check_output "content_scan_benign_action"      "Benign Scan Action"
check_output "content_scan_injection_category" "Injection Scan Category"
check_output "content_scan_injection_action"   "Injection Scan Action"

# ── Summary ────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ $FAILURES -eq 0 ]]; then
  echo -e "${GREEN}ALL CHECKS PASSED${NC}"
else
  echo -e "${RED}$FAILURES CHECK(S) FAILED${NC}"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ── Destroy ────────────────────────────────────────────────────────────
if ! $NO_DESTROY; then
  echo ""
  info "Destroying resources"
  terraform destroy -auto-approve
fi

# ── Cleanup ────────────────────────────────────────────────────────────
rm -f "$TF_CLI_CONFIG_FILE"

if [[ $FAILURES -gt 0 ]]; then
  exit 1
fi
