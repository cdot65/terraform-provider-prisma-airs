# prisma-airs_gateway_workspace schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Coordinates Gateway workspace creation and archival with dedicated IAM scope ownership. External scopes receive no writes. Membership and role assignments are separately managed.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `defaults` | `single(object)` | optional | — | Owns the complete default config selection and metadata object. Removing a previously configured object clears those fields. Unreadable allow_config_override is not supported. |
| `description` | `string` | optional | — | Workspace description. The API does not clear blank descriptions; removal stops managing this field. |
| `icon` | `string` | optional | — | Workspace icon. Removing a previously configured value clears it. |
| `id` | `string` | computed | — | Workspace UUID; a pending recovery identity may be present after a failed create. |
| `last_updated_at` | `string` | computed | — | Last update timestamp. |
| `name` | `string` | required | — | Workspace display label. Renaming preserves its UUID and unique slug. |
| `provisioning_stage` | `string` | computed | — | Last completed provisioning or cleanup stage; retained after partial failures. |
| `rate_limits` | `list(object)` | optional | — | Owns the complete workspace rate-policy collection. Empty or removed collections clear the previously managed policies. |
| `scope_binding_ready` | `bool` | computed | — | Whether the scope is bound to this workspace slug. This does not grant service-account access. |
| `scope_management` | `string` | required | — | managed creates and cleans a dedicated scope; external never writes IAM. External owners maintain bindings and role grants. |
| `scope_name` | `string` | required | — | Stable IAM scope name, rather than its composite display ID. |
| `scope_owned` | `bool` | computed | — | Whether Terraform acknowledged ownership of the dedicated IAM scope. |
| `scope_ownership_token` | `string` | computed | — | Correlation token saved before scope creation; never a credential or permission grant. |
| `slug` | `string` | computed | — | Server-assigned unique workspace slug. |
| `status` | `string` | computed | — | Remote workspace lifecycle status. |
| `usage_limits` | `list(object)` | optional | — | Owns the complete workspace usage-policy collection (the service permits one policy). Empty or removed collections clear its associations; hidden backing-record deletion is not asserted. |

### Attributes.defaults

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `config_id` | `string` | optional | — | Existing default config UUID. Use references rather than embedded credentials. |
| `metadata` | `dynamic` | optional | — | Native HCL metadata object. Embedded credential fields are rejected. |

### Attributes.rate_limits

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `type` | `string` | required | — | Measurement type, for example requests. |
| `unit` | `string` | required | — | Rate unit, for example rpm. |
| `value` | `number` | required | — | Nonnegative integer rate. |

### Attributes.usage_limits

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_threshold` | `number` | optional | — | Optional nonnegative alert threshold; zero is explicit. |
| `credit_limit` | `number` | required | — | Nonnegative usage allowance. |
| `next_usage_reset_at` | `string` | optional | — | Optional next reset timestamp in RFC 3339 format. |
| `periodic_reset` | `string` | optional | — | Optional reset cadence, such as monthly or weekly. |
| `periodic_reset_days` | `number` | optional | — | Optional positive custom reset interval in days. |
| `type` | `string` | required | — | Usage measurement type, for example tokens. |
