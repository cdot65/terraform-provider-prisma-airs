---
title: Getting started
slug: /getting-started
---

Manage your first Prisma AIRS security profile with Terraform. This walkthrough follows the same setup path as the AIRS CLI and SDK sites: obtain tenant access, configure credentials, install the tool, and verify a real operation.

## Before you start

You need Terraform 1.0 or later, an AIRS tenant with Runtime Security management access, and an OAuth service account with permission to manage profiles. Model Security and Red Team resources require their own service entitlements.

:::info[Updated provider]

These guides describe provider v0.10.0, built on Go SDK v0.8.1. v0.7.0 and earlier use different resource names or configuration schemas. Follow [installation](installation.md) and review [migration](../guides/migration.md) before upgrading existing state.

:::

## 1. Configure your tenant

Obtain a client ID, client secret, and tenant service group (TSG) ID using the [authentication guide](authentication.md). Set the three `PANW_MGMT_*` variables through your secret manager. The provider uses this credential set across AI Runtime Security, AI Supply Chain Security, and AI Red Teaming.

## 2. Install the provider

Follow [installation](installation.md) to install provider v0.10.0 from the Terraform Registry. Work in a separate directory for the configuration below.

## 3. Write your first profile

Save as `main.tf`:

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

# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "first" {
  profile_name = "terraform-first-profile"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }
  }
}

# Outputs: Expose results for the next configuration or application step.
output "profile_revision" {
  value = prisma-airs_runtime_security_profile.first.revision
}
```

Use a unique profile name. If that name already exists, [import it](../resources/security-profile.md#import) instead of creating a second owner.

## 4. Review and apply

Initialize the working directory, then review a saved plan before applying:

```bash
terraform init
terraform fmt
terraform validate
terraform plan -out=first.tfplan
terraform apply first.tfplan
terraform plan -detailed-exitcode
```

Review the saved plan before applying it. A successful apply records the service UUID and revision. The final plan should report no changes and exit with code 0; exit code 2 means changes remain, while code 1 indicates an error.

## 5. Change a policy

Change the protection action to `"allow"` and run `terraform plan`. Terraform plans an update at the same resource address. AIRS creates a new UUID and revision; unchanged configuration stays known in the plan. See [revision ownership](../resources/security-profile.md#revision-ownership-and-changes).

## 6. Clean up deliberately

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

:::warning[Profile history ownership]

Destroy deletes every revision under the currently managed profile name, including revisions created before import or by another actor. Renaming follows a new name and leaves the old history in AIRS.

:::

## Choose your next task

| Task | Guide |
| --- | --- |
| Configure credentials and endpoints | [Configuration](configuration.md) |
| Manage topics and revisioned profiles | [Security profiles](../guides/managing-security-profiles.md) |
| Register a model or custom endpoint | [Red Team targets](../guides/red-team-testing.md) |
| Manage licensed model groups | [Model Security](../guides/model-security-workflow.md) |
| Adopt existing infrastructure | [Import and state](../guides/import-and-state.md) |
| Update existing HCL and state | [Migration](../guides/migration.md) |
