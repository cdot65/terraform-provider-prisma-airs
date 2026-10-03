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

## Before you start

The profiles reference existing custom topics and DLP profiles by name. Inspect `profiles.tf` and select names already available in your tenant, or provision those dependencies separately. These files do not create them.

Load `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` from your secret store into the environment. See the [authentication guide](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/authentication/).

Copy `terraform.tfvars.example` to `terraform.tfvars` and set a distinct `profile_prefix`. Keep only the use cases you want to manage before the first apply. Removing profiles already in state proposes their destruction.

## Apply and inspect

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

After editing the configuration, review another saved plan before applying it. An unchanged `terraform plan -detailed-exitcode` should exit 0.

Outputs identify each owned profile. Edit a protection action to practice an update; Terraform keeps the resource address while the service creates a new profile UUID and revision.

## Clean up

Use the same inputs, credentials, and state to remove the owned resources:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Profile deletion covers all revisions owned by that profile. Existing topic and DLP dependencies remain external. For a self-contained profile, topic, application, and API-key workflow, use the [public Runtime Security project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-runtime-security).

## Files

| File | Purpose |
|------|---------|
| `main.tf` | Provider configuration |
| `variables.tf` | Variable declarations |
| `terraform.tfvars.example` | Example variable values (copy to `terraform.tfvars`) |
| `profiles.tf` | Security profile resources |
| `outputs.tf` | Profile IDs, names, and status |
