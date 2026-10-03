# prisma-airs_supply_chain_skill_scanning_instance schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Skill Scanning tenant instance. Existing instances require import; destroy deletes the tenant instance, not individual scans.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `auth_code` | `string` | optional, write_only | yes | Deployment authorization code, never stored in plan or state. Requires Terraform 1.11+. Increment auth_code_version to change or clear it. |
| `auth_code_version` | `number` | optional | — | Nonsecret trigger for authorization-code changes. Increment to send the configured code, or explicit null when clearing it. |
| `created_by` | `string` | required | — | Creator identity required by the service. |
| `iam_controlled` | `bool` | optional | — | Request-only IAM setting, persisted in state. Explicit false is sent; removal sends null on update. GET cannot detect drift. |
| `id` | `string` | computed | — | Terraform identity. |
| `registration_details` | `dynamic` | required | yes | Complete desired native registration object for SDK provisioning metadata (region, license_name, entitlements, tsg_instances, extra, and extensions). PUT owns this inventory and can deactivate removed profiles. Identity/authentication fields belong to their dedicated attributes. Never supply a JSON string. This desired object is retained because GET cannot recover the complete registration request. |
| `support_account_id` | `string` | required | — | Support account ID. |
| `support_account_name` | `string` | optional | — | Support account name; removing a configured value sends null on update. |
| `tenant_id` | `string` | required | — | Skill Scanning tenant ID. Changing it replaces the instance. |
| `tsg_id` | `string` | computed | — | Tenant Service Group ID. |
