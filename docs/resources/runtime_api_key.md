---
page_title: "prisma-airs_runtime_api_key (Resource)"
subcategory: "AI Runtime Security"
---

# prisma-airs_runtime_api_key Resource

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

**Note**
The `api_key` attribute is only available after creation. Store it securely as it cannot be retrieved later.


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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_runtime_api_key/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_key` | `string` | computed | yes | The API key value. Only available at creation time. |
| `api_key_id` | `string` | computed | — | The unique identifier of the API key. |
| `api_key_name` | `string` | required | — | Name of the API key. |
| `auth_code` | `string` | required | yes | Deployment profile auth code for API key creation. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `created_by` | `string` | optional, computed | — | Identity of the user creating the key. |
| `cust_ai_agent_framework` | `string` | optional, computed | — | Customer AI agent framework. |
| `cust_app` | `string` | optional, computed | — | Customer application name to associate with the key. |
| `cust_cloud_provider` | `string` | optional, computed | — | Customer cloud provider used when creating the associated application. |
| `cust_env` | `string` | optional, computed | — | Customer environment used when creating the associated application. |
| `expires_at` | `string` | computed | — | Expiration timestamp. |
| `id` | `string` | computed | — | Terraform resource ID (same as api_key_id). |
| `revoked` | `bool` | computed | — | Whether the API key is revoked. |
| `rotation_time_interval` | `number` | required | — | Rotation interval value (e.g. 90 for 90 days). |
| `rotation_time_unit` | `string` | required | — | Rotation time unit (days, months). |
| `status` | `string` | computed | — | API key status. |
