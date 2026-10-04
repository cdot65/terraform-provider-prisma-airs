# Red Team adapters

Save an adapter draft, activate it deliberately, and register an adapter-backed target. This walkthrough uses a deterministic text response so you can learn the dependency graph without an upstream model key. It is a connectivity exercise, not an adversarial assessment of a model.

## Before you start

Load the [management environment variables](../getting-started/authentication.md). You need an existing online Network Broker channel with text adapter support. Terraform manages the adapter and target; broker installation and upgrades remain prerequisites. Use provider v0.12.0 or later for adapter ownership and discovery.

Save this as `adapter.py`:

```python
# Adapter: Return a deterministic response through the broker.
def call_target(context, inference_input):
    return CallTargetResult(output=context.vars["message"])
```

Save the following complete configuration as `main.tf`. The draft path creates just the adapter; `activate = true` runs its script and creates the target.

```hcl
# Setup: Install the provider used by this example.
terraform {
  required_version = ">= 1.11.0, < 2.0.0"
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.12.0"
    }
  }
}

# Authentication: Read management credentials from the environment.
provider "prisma-airs" {}

# Inputs: Activation is explicit because it executes the script.
variable "activate" {
  type    = bool
  default = false
}

variable "network_broker_channel_uuid" {
  type    = string
  default = null
}

# Adapter: Encode the Python file internally and save a draft by default.
resource "prisma-airs_red_team_adapter" "example" {
  name                        = "terraform-adapter-example"
  script                      = file("${path.module}/adapter.py")
  validate                    = var.activate
  network_broker_channel_uuid = var.network_broker_channel_uuid
  validation_prompt           = "Text-only connectivity exercise"

  variables = {
    message = {
      type  = "VAR"
      value = "Terraform adapter is reachable"
    }
  }
}

# Target: Register only after deliberate adapter activation succeeds.
resource "prisma-airs_red_team_target" "example" {
  count                       = var.activate ? 1 : 0
  name                        = "terraform-adapter-target"
  target_type                 = "APPLICATION"
  api_endpoint_type           = "NETWORK_BROKER"
  network_broker_channel_uuid = var.network_broker_channel_uuid

  adapter {
    uuid = prisma-airs_red_team_adapter.example.id
  }
}

# Outputs: Show managed identities without exposing scripts or variables.
output "adapter_id" {
  value = prisma-airs_red_team_adapter.example.id
}

output "target_id" {
  value = one(prisma-airs_red_team_target.example[*].id)
}
```

## Apply in two steps

```bash
terraform init
terraform plan -out=draft.tfplan
terraform apply draft.tfplan
```

Set `activate = true` and your channel UUID in `terraform.tfvars`. Review the activation plan before applying it:

```bash
terraform plan -out=activate.tfplan
terraform apply activate.tfplan
terraform output
```

Activation executes the adapter through the broker. Target creation registers configuration without running an assessment. If execution fails, inspect the broker connection and adapter support before retrying; an ONLINE channel alone does not prove a compatible broker client. Do not toggle activation off as a troubleshooting step without reviewing the proposed target destruction.

For an existing adapter, use [discovery](../data-sources/red-team-adapters.md) to reference it, or [import](../resources/red-team-adapter.md#import-an-existing-adapter) to manage its script and variables. The [repository project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-red-teaming) also demonstrates secret-preserving adoption and includes sanitized live-run evidence.

## Clean up

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Terraform removes the dependent target first and then the adapter. It does not remove the channel or broker installation.
