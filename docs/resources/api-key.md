# prisma-airs_api_key

Manages an API key for AI Runtime Security scanning in Prisma AIRS.

## Example Usage

```hcl
resource "prisma-airs_api_key" "scanner" {
  api_key_name = "production-scanner"
  created_by   = "terraform"
}
```

## Argument Reference

- `api_key_name` - (Required, ForceNew) Name for the API key.
- `created_by` - (Optional) Identity of who created the key.

## Attribute Reference

- `id` - The API key ID.
- `api_key_id` - The API key ID (same as `id`).
- `api_key` - The generated API key value (sensitive).
- `status` - API key status.
- `revoked` - Whether the key is revoked.
- `created_at` - Timestamp when the key was created.
- `expires_at` - Expiration timestamp.

!!! warning
    The `api_key` attribute is only available after creation. Store it securely as it cannot be retrieved later.

## Import

API keys can be imported using the key ID:

```bash
terraform import prisma-airs_api_key.scanner <api_key_id>
```
