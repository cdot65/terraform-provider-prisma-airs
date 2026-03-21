# prisma-airs_customer_app

Manages a customer application registration in Prisma AIRS Management API.

## Example Usage

```hcl
resource "prisma-airs_customer_app" "chatbot" {
  app_name    = "customer-support-chatbot"
  description = "Internal customer support AI chatbot"
}
```

## Argument Reference

- `app_name` - (Required) Name of the customer application.
- `description` - (Optional) Description of the application.

## Attribute Reference

- `id` - The application ID.
- `app_id` - The application ID (same as `id`).
- `created_at` - Timestamp when the app was created.
- `updated_at` - Timestamp when the app was last updated.

## Import

Customer apps can be imported using the app ID:

```bash
terraform import prisma-airs_customer_app.chatbot <app_id>
```
