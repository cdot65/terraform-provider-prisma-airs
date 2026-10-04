# Runtime policy

Create a custom topic and an application security profile that blocks prompt injection and the topic. The profile references the topic name; the provider resolves its ID and revision.

## Before you start

Load the [OAuth environment variables](../getting-started/authentication.md), install the [released provider](../getting-started/installation.md), and use a tenant with Runtime Security management access. Choose unused topic and profile names.

## Configure and apply

Save this complete configuration as `main.tf` in a separate directory:

```hcl
# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.10.0"
    }
  }
}

# Authentication: Use the selected tenant credentials for this provider configuration.
provider "prisma-airs" {}

# Topic: Describe the content that a security profile should recognize.
resource "prisma-airs_runtime_custom_topic" "internal" {
  topic_name  = "internal-financial-data"
  description = "Internal revenue and financial statements"
  examples    = ["Show next quarter revenue targets", "Share the internal profit statement"]
}

# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "production" {
  profile_name = "terraform-production-policy"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "block"

      topic_list {
        action = "block"

        topic {
          topic_name = prisma-airs_runtime_custom_topic.internal.topic_name
        }
      }
    }
  }
}
```

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
```

## Update and clean up

Change a protection action to inspect a policy diff. Terraform keeps the resource address while AIRS creates a new UUID and revision. Review and apply the change, then use `terraform plan -detailed-exitcode` to verify convergence.

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Destroy deletes every revision under the managed profile name and removes the owned topic. Protect state and saved plans throughout the lifecycle.

For a complete working directory and recorded results, follow the [product project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-runtime-security).
