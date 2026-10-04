# prisma-airs_red_team_adapter

Read an existing adapter's script, channel reference, variable inventory, and status. This lookup does not execute the script or manage its lifecycle.

```hcl
# Discovery: Select an identity, then inspect its configuration.
data "prisma-airs_red_team_adapters" "available" {}

data "prisma-airs_red_team_adapter" "application" {
  id = data.prisma-airs_red_team_adapters.available.ids_by_name[var.adapter_name]
}
```

Secret values are null; scripts and variable maps are sensitive. An inaccessible or missing adapter produces a diagnostic rather than an empty object. Use the [adapter resource](../resources/red-team-adapter.md) and import to take ownership; merely reading it does not permit deletion by Terraform.

## Complete schema

See the [exact schema reference](../reference/generated/data-source-prisma-airs_red_team_adapter.md).
