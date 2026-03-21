# prisma-airs_red_team_target

Manages a red team target in Prisma AIRS Red Team API.

## Example Usage

```hcl
resource "prisma-airs_red_team_target" "my_app" {
  name        = "production-chatbot"
  target_type = "APPLICATION"
  description = "Production customer-facing chatbot"

  connection = jsonencode({
    type     = "REST"
    endpoint = "https://my-app.example.com/api/chat"
    headers = {
      "Authorization" = "Bearer ${var.app_token}"
    }
  })
}
```

## Argument Reference

- `name` - (Required) Name of the target.
- `target_type` - (Required) Type of target. Valid values: `APPLICATION`, `AGENT`, `MODEL`.
- `description` - (Optional) Description of the target.
- `connection` - (Required) JSON-encoded connection configuration.
- `validate` - (Optional) Whether to validate the target connection on create/update. Default: `true`.

## Attribute Reference

- `id` - The target UUID.
- `uuid` - The target UUID (same as `id`).
- `status` - Target status (`DRAFT`, `VALIDATING`, `VALIDATED`, `ACTIVE`, `INACTIVE`, `FAILED`, `PENDING_AUTH`).
- `created_at` - Timestamp when the target was created.
- `updated_at` - Timestamp when the target was last updated.

## Import

```bash
terraform import prisma-airs_red_team_target.my_app <uuid>
```
