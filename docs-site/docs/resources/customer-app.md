# prisma-airs_runtime_customer_app

Manages an existing customer application in Prisma AIRS Management API.

:::warning

Customer apps are created externally (via the AIRS console or when applications register themselves). This resource does **not** support `terraform apply` for new apps — use `terraform import` to bring an existing app under Terraform management, then update or delete it.

:::

## Example Usage

### Import an existing app

```bash
terraform import prisma-airs_runtime_customer_app.chatbot customer-support-chatbot
```

### Manage the imported app

```hcl
# Existing app: Import the externally created application before managing it.
resource "prisma-airs_runtime_customer_app" "chatbot" {
  app_name       = "customer-support-chatbot"
  model_name     = "gpt-4"
  cloud_provider = "aws"
  environment    = "production"
}
```

## Argument Reference

- `app_name` - (Required) Name of the customer application (used as the lookup key).
- `model_name` - (Optional) Model name associated with the app.
- `cloud_provider` - (Optional) Cloud provider for the app.
- `environment` - (Optional) Deployment environment.
- `updated_by` - (Optional) Identity of the user updating the app.

## Attribute Reference

- `id` - The application ID.
- `customer_app_id` - The application ID (same as `id`).
- `tsg_id` - Tenant service group ID.
- `status` - App status.
- `created_by` - Identity of the user who created the app.
- `agent_app` - Whether this is an agent application.
- `ai_agent_framework` - AI agent framework.
- `ai_sec_profile_name` - Associated AI security profile name.

## Import

Customer apps are imported by app name:

```bash
terraform import prisma-airs_runtime_customer_app.chatbot <app_name>
```

## Service compatibility

Read, import, and cleanup verification use the supported paginated customer-app list endpoint. The legacy single-app GET can return 403 even for a superuser account. A missing app is removed from state; the next plan requires creation, which this import-only resource cannot perform.

The live update API requires a deployment `auth_code`. SDK v0.6.1, pinned by this provider, resolves an unambiguous code through the supported list endpoint. Multiple distinct deployment codes produce an explicit ambiguity error rather than choosing a credential silently. Renaming an imported app is unsupported and returns a plan diagnostic; import a different application to manage a different name.

Explicit empty values for app name, model, cloud provider, or environment are rejected during validation because the SDK request omits them rather than clearing the existing value. Omitted optional fields retain observed metadata.

Renaming an imported application is rejected during planning. Establish the desired name outside Terraform, then import it at a separate address. `updated_by` is sent on updates and used for deletion.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_runtime_customer_app.md) for all nested fields, types, and sensitivity flags.
