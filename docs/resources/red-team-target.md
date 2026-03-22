# prisma-airs_red_team_target

Manages a red team target in Prisma AIRS Red Team API.

## Example Usage

```hcl
resource "prisma-airs_red_team_target" "my_app" {
  name            = "production-chatbot"
  target_type     = "APPLICATION"
  description     = "Production customer-facing chatbot"
  connection_type = "REST"

  connection_params = jsonencode({
    url = "https://my-app.example.com/api/chat"
    headers = {
      "Authorization" = "Bearer ${var.app_token}"
      "Content-Type"  = "application/json"
    }
    request_json = {
      prompt = "{INPUT}"
    }
    response_json = {
      output = "{RESPONSE}"
    }
    response_key = "output"
  })
}
```

## Argument Reference

- `name` - (Required) Name of the target.
- `target_type` - (Optional) Type of target: `APPLICATION`, `AGENT`, `MODEL`.
- `description` - (Optional) Description of the target.
- `connection_type` - (Optional) Connection type: `REST`, `STREAMING`, `OPENAI`, `BEDROCK`, `DATABRICKS`, `HUGGING_FACE`, `CUSTOM`.
- `connection_params` - (Optional, Sensitive) JSON-encoded connection parameters.

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
