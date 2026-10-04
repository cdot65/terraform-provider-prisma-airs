# AI Red Teaming

Manages red team targets and custom prompt sets for adversarial testing of AI applications. Demonstrates multiple connection types and target configurations.

## Targets

| Resource | Model | Connection | Description |
|----------|-------|------------|-------------|
| `litellm_multiturn` | Mistral-7b | REST / Public | LiteLLM proxy (legacy resource address) |
| `litellm_singleturn` | Mistral-7b | REST / Public | Single-turn baseline testing |
| `worf` | Qwen3-14B-AWQ | REST / Public | Local vLLM with AIRS guardrails |
| `bedrock_claude` | Claude Opus 4.6 | BEDROCK | AWS Bedrock REST (MODEL type) |
| `qwen_completions` | Qwen2.5-7B-Instruct | REST / Public | Completions API (not chat) |
| `talkdesk` | Virtual Agent | REST / Public | Talkdesk with custom request format |

## Prompt Sets

| Resource | Purpose |
|----------|---------|
| `general_adversarial` | Common attack vectors (jailbreaks, prompt injection) |
| `compliance` | Data protection and compliance bypasses |
| `agent_attacks` | AI agent-specific vulnerabilities |

## Before you start

This example pins provider v0.10.0; use endpoints you control. The endpoints in `targets.tf` illustrate different request formats; replace them and keep only the targets and variables for services you intend to configure **before the first apply**. Removing a target from existing state proposes its destruction.

Load `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` from your secret store into the environment. See the [authentication guide](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/authentication/).

Load each retained target's credential through its corresponding `TF_VAR_*` variable. `terraform.tfvars.example` lists these names; it contains no credential assignments. Targets use native connection blocks and separate sensitive authentication blocks. For private connectivity, supply a preexisting channel and `api_endpoint_type = "NETWORK_BROKER"`. Multi-turn configuration is not exposed by these blocks.

## Apply and inspect

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

After editing the configuration, review another saved plan before applying it. An unchanged `terraform plan -detailed-exitcode` should exit 0.

Apply registers targets and prompt sets; it does not launch an adversarial test. Protect state and saved plans, which can contain target credentials.

## Clean up

Use the same inputs, credentials, and state to remove the owned resources:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

For a cohesive target-and-prompt-set walkthrough, use the [public Red Teaming project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-red-teaming).

## Files

| File | Purpose |
|------|---------|
| `main.tf` | Provider configuration |
| `variables.tf` | Sensitive variable declarations (API keys, credentials) |
| `terraform.tfvars.example` | Environment variable names for target credentials |
| `targets.tf` | Red team target resources |
| `prompt_sets.tf` | Custom prompt set resources |
| `outputs.tf` | Target and prompt set IDs, names, and status |
