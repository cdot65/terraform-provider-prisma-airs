# prisma-airs_api_key

Manages an API key for AI Runtime Security scanning in Prisma AIRS.

## Example Usage

```hcl
resource "prisma-airs_api_key" "scanner" {
  api_key_name = "production-scanner"
  updated_by   = "terraform"
}
```

## Argument Reference

- `api_key_name` - (Required) Name for the API key.
- `updated_by` - (Optional) Identifier of who created/updated the key.

## Attribute Reference

- `id` - The API key ID.
- `api_key_id` - The API key ID (same as `id`).
- `api_key` - The generated API key value (sensitive).
- `active` - Whether the key is active.
- `created_at` - Timestamp when the key was created.
- `expires_at` - Timestamp when the key expires.

!!! warning
    The `api_key` attribute is only available after creation or regeneration. Store it securely as it cannot be retrieved later.

## Import

API keys can be imported using the key ID:

```bash
terraform import prisma-airs_api_key.scanner <api_key_id>
```
