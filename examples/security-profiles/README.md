# Security Profiles at Scale

Manages multiple AI security profiles with varying protection levels — from high-security (all protections, strict blocking, DLP) to lightweight (prompt injection only).

## Profiles

| Resource | Use Case | Protections |
|----------|----------|-------------|
| `high_security` | Enterprise AI firewall | All protections + DLP + URL filtering |
| `truffles_agent` | Recipe AI agent | Topic guardrails + data leak masking |
| `recipe_extractor` | AWS recipe extraction agent | Topic guardrails + moderate toxic filtering |
| `cursor_ide` | Code assistant IDE integration | Strict toxic content + data leak detection |
| `slack_moderation` | Internal Slack bot | Prompt injection only (lightweight) |
| `hipaa_compliance` | Healthcare agent | HIPAA DLP profile + topic guardrails |

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars if needed

# For local development builds:
export TF_CLI_CONFIG_FILE=/path/to/.dev.tfrc

# Set provider credentials
export PANW_MGMT_CLIENT_ID="..."
export PANW_MGMT_CLIENT_SECRET="..."
export PANW_MGMT_TSG_ID="..."

terraform init
terraform plan
terraform apply
```

## Files

| File | Purpose |
|------|---------|
| `main.tf` | Provider configuration |
| `variables.tf` | Variable declarations |
| `terraform.tfvars.example` | Example variable values (copy to `terraform.tfvars`) |
| `profiles.tf` | Security profile resources |
| `outputs.tf` | Profile IDs, names, and status |
