# prisma-airs_api_key

Manages an API key for AI Runtime Security scanning in Prisma AIRS.

## Example Usage

```hcl
data "prisma-airs_deployment_profiles" "all" {
  limit = 10
}

resource "prisma-airs_api_key" "scanner" {
  api_key_name           = "production-scanner"
  auth_code              = data.prisma-airs_deployment_profiles.all.items[0].auth_code
  rotation_time_interval = 90
  rotation_time_unit     = "days"
  created_by             = "terraform"
}
```

## Argument Reference

- `api_key_name` - (Required, ForceNew) Name for the API key.
- `auth_code` - (Required, ForceNew) Deployment profile auth code. Use the `prisma-airs_deployment_profiles` data source to discover available auth codes.
- `rotation_time_interval` - (Required) Rotation interval value (e.g., `90` for 90 days).
- `rotation_time_unit` - (Required) Rotation time unit (`days`, `months`).
- `cust_app` - (Optional) Customer application name to associate with the key.
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
