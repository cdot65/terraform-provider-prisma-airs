# AI Red Teaming

Manages red team targets and custom prompt sets for adversarial testing of AI applications. Demonstrates multiple connection types and target configurations.

## Targets

| Resource | Model | Connection | Description |
|----------|-------|------------|-------------|
| `litellm_multiturn` | Mistral-7b | CUSTOM / Network Broker | Multi-turn via LiteLLM proxy |
| `litellm_singleturn` | Mistral-7b | CUSTOM / Network Broker | Single-turn baseline testing |
| `worf` | Qwen3-14B-AWQ | CUSTOM / Network Broker | Local vLLM with AIRS guardrails |
| `bedrock_claude` | Claude Opus 4.6 | BEDROCK | AWS Bedrock streaming (MODEL type) |
| `qwen_completions` | Qwen2.5-7B-Instruct | CUSTOM / Public | Completions API (not chat) |
| `talkdesk` | Virtual Agent | CUSTOM / Network Broker | Talkdesk with custom request format |

## Prompt Sets

| Resource | Purpose |
|----------|---------|
| `general_adversarial` | Common attack vectors (jailbreaks, prompt injection) |
| `compliance` | Data protection and compliance bypasses |
| `agent_attacks` | AI agent-specific vulnerabilities |

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# Fill in API keys and credentials in terraform.tfvars

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
| `variables.tf` | Sensitive variable declarations (API keys, credentials) |
| `terraform.tfvars.example` | Example variable values (copy to `terraform.tfvars`) |
| `targets.tf` | Red team target resources |
| `prompt_sets.tf` | Custom prompt set resources |
| `outputs.tf` | Target and prompt set IDs, names, and status |
