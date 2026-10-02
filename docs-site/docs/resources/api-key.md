# prisma-airs_runtime_api_key

Manages an API key for AI Runtime Security scanning in Prisma AIRS.

## Example Usage

```hcl
data "prisma-airs_runtime_deployment_profiles" "all" {
  limit = 10
}

resource "prisma-airs_runtime_api_key" "scanner" {
  api_key_name           = "production-scanner"
  auth_code              = data.prisma-airs_runtime_deployment_profiles.all.items[0].auth_code
  rotation_time_interval = 90
  rotation_time_unit     = "days"
  created_by             = "terraform"
}
```

## Argument Reference

- `api_key_name` - (Required, ForceNew) Name for the API key.
- `auth_code` - (Required, ForceNew) Deployment profile auth code. Use the `prisma-airs_runtime_deployment_profiles` data source to discover available auth codes.
- `rotation_time_interval` - (Required, ForceNew) Rotation interval value (e.g., `90` for 90 days).
- `rotation_time_unit` - (Required, ForceNew) Rotation time unit (`days`, `months`).
- `cust_app` - (Optional, ForceNew) Customer application name to associate with the key.
- `cust_env` - (Optional, ForceNew) Customer environment.
- `cust_cloud_provider` - (Optional, ForceNew) Customer cloud provider.
- `cust_ai_agent_framework` - (Optional, ForceNew) Customer agent framework.
- `created_by` - (Optional, ForceNew) Identity of who created the key.

## Attribute Reference

- `id` - The API key ID.
- `api_key_id` - The API key ID (same as `id`).
- `api_key` - The generated API key value (sensitive).
- `status` - API key status.
- `revoked` - Whether the key is revoked.
- `created_at` - Timestamp when the key was created.
- `expires_at` - Expiration timestamp.

:::warning

The `api_key` attribute is only available after creation. Store it securely as it cannot be retrieved later.

:::

## Import

API keys can be imported using the key ID:

```bash
terraform import prisma-airs_runtime_api_key.scanner <api_key_id>
```

## Replacement and secret state

All create-time inputs plan replacement, including rotation settings and customer metadata. Additional optional replacement inputs are `cust_env`, `cust_cloud_provider`, and `cust_ai_agent_framework`. Regeneration is outside this resource's current scope.

Refresh preserves the creation-time `api_key`. An imported key has a null, unavailable key value; importing never regenerates a secret. `auth_code` and deployment-profile authentication details are sensitive. Terraform state still contains sensitive values: protect the state backend and avoid committing state or credentials.

The live service deletes the associated customer app when its key is deleted. Use disposable app associations in tests and account for this cascade when destroying managed keys.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_runtime_api_key.md) for all nested fields, types, and sensitivity flags.
