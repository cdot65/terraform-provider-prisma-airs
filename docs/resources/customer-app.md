# prisma-airs_customer_app

Manages an existing customer application in Prisma AIRS Management API.

!!! important
    Customer apps are created externally (via the AIRS console or when applications register themselves). This resource does **not** support `terraform apply` for new apps — use `terraform import` to bring an existing app under Terraform management, then update or delete it.

## Example Usage

### Import an existing app

```bash
terraform import prisma-airs_customer_app.chatbot customer-support-chatbot
```

### Manage the imported app

```hcl
resource "prisma-airs_customer_app" "chatbot" {
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
terraform import prisma-airs_customer_app.chatbot <app_name>
```
