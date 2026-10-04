---
page_title: "prisma-airs_red_team_adapter (Data Source)"
subcategory: "AI Red Teaming"
---

# prisma-airs_red_team_adapter Data Source

Read an existing adapter's script, channel reference, variable inventory, and status. This lookup does not execute the script or manage its lifecycle.

```hcl
# Discovery: Select an identity, then inspect its configuration.
data "prisma-airs_red_team_adapters" "available" {}

data "prisma-airs_red_team_adapter" "application" {
  id = data.prisma-airs_red_team_adapters.available.ids_by_name[var.adapter_name]
}
```

Secret values are null; scripts and variable maps are sensitive. An inaccessible or missing adapter produces a diagnostic rather than an empty object. Use the [adapter resource](https://cdot65.github.io/terraform-provider-prisma-airs/resources/red-team-adapter/) and import to take ownership; merely reading it does not permit deletion by Terraform.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/data-source-prisma-airs_red_team_adapter/).

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `description` | `string` | computed | — | Adapter description. |
| `id` | `string` | required | — | Adapter UUID, typically selected from red_team_adapters.ids_by_name. |
| `name` | `string` | computed | — | Adapter name. |
| `network_broker_channel_uuid` | `string` | computed | — | Existing channel reference, if present. |
| `script` | `string` | computed | yes | Decoded Python script; never executed by this data source. |
| `status` | `string` | computed | — | DRAFT or ACTIVE. |
| `variables` | `map(object)` | computed | yes | Variable key/type inventory; redacted secrets have null values. |

#### Attributes.variables

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `type` | `string` | computed | — | VAR or SECRET. |
| `value` | `string` | computed | yes | Visible value or null if redacted. |
