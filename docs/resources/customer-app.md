# prisma-airs_customer_app

Manages a customer application registration in Prisma AIRS Management API.

## Example Usage

```hcl
resource "prisma-airs_customer_app" "chatbot" {
  app_name      = "customer-support-chatbot"
  tsg_id        = "1234567890"
  cloud_provider = "aws"
  environment   = "production"
}
```

## Argument Reference

- `app_name` - (Required) Name of the customer application.
- `tsg_id` - (Optional, ForceNew) Tenant service group ID.
- `model_name` - (Optional) Model name associated with the app.
- `cloud_provider` - (Optional) Cloud provider for the app.
- `environment` - (Optional) Deployment environment.
- `updated_by` - (Optional) Identity of the user updating the app.

## Attribute Reference

- `id` - The application ID.
- `customer_app_id` - The application ID (same as `id`).
- `status` - App status.
- `created_by` - Identity of the user who created the app.
- `ai_agent_framework` - AI agent framework.

## Import

Customer apps can be imported using the app name:

```bash
terraform import prisma-airs_customer_app.chatbot <app_name>
```
