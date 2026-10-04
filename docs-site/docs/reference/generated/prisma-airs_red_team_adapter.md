# prisma-airs_red_team_adapter schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Red Team adapter script and its complete variable key set. Defaults to draft writes without execution. Explicit validate=true executes through an existing Network Broker channel.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Adapter description. |
| `id` | `string` | computed | — | Adapter UUID. |
| `name` | `string` | required | — | Adapter name. |
| `network_broker_channel_uuid` | `string` | optional | — | Existing Network Broker channel; required to execute validation. Terraform does not install or manage the broker. |
| `script` | `string` | required | yes | Plaintext Python adapter script. Base64 encoding is internal; scripts may contain credentials. |
| `status` | `string` | computed | — | DRAFT or ACTIVE. |
| `updated_at` | `string` | computed | — | Update timestamp. |
| `validate` | `bool` | optional, computed | — | Explicit execution opt-in on create/update. False saves DRAFT; true validates through Network Broker. Import sets true only for ACTIVE adapters without executing them. |
| `validation_prompt` | `string` | optional, computed | yes | Transient prompt sent when validate=true; not recovered on import. |
| `variables` | `map(object)` | optional, computed | yes | Complete desired key set. Omitted keys are deleted on update. Null preserves an existing SECRET with the same key/type. |

### Attributes.variables

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `type` | `string` | required | — | VAR or SECRET. |
| `value` | `string` | optional | yes | Desired value, or null to retain an imported SECRET. New variables need values. |
