# AI Gateway

Create an upstream integration, bind it to an existing workspace, expose a provider, and save a single-model routing configuration. This focused project teaches how those resources depend on one another.

## Before you start

This example pins provider v0.10.0. You need an existing Gateway workspace, a provider-family UUID, and a model enabled for that connection.

Load `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` from your secret store into the environment. See the [authentication guide](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/authentication/).

Copy `terraform.tfvars.example` to `terraform.tfvars` and set `workspace_id`, `ai_provider_id`, and `model`. Load the real upstream credential through `TF_VAR_upstream_api_key`. Choose unused resource names in `main.tf`.

## Apply and inspect

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

After editing the configuration, review another saved plan before applying it. An unchanged `terraform plan -detailed-exitcode` should exit 0.

`config_id` identifies the routing configuration and `config_version_id` identifies its current version. Apply manages configuration; it does not send inference traffic. Protect state and saved plans, which can contain the upstream key.

Change `retry.attempts` to practice a routing update. The config ID stays stable while its version changes.

## Clean up

Use the same inputs, credentials, and state to remove the owned resources:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy removes the config, provider, and integration and disables their workspace binding. The existing workspace remains external. See the [Gateway workflow](https://cdot65.github.io/terraform-provider-prisma-airs/guides/gateway-workflow/) for secret and archive behavior.

## Continue with the platform example

Use the [expanded public Gateway project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-gateway) for AIRS guardrails, application credentials, four routing patterns, limits, and optional platform features. Its guide includes explicit inference requests and recorded live results.
