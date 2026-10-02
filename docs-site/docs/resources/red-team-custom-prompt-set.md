# prisma-airs_red_team_custom_prompt_set

Manages a custom prompt set for red team testing in Prisma AIRS Red Team API.

## Example Usage

```hcl
resource "prisma-airs_red_team_custom_prompt_set" "injection_tests" {
  name        = "prompt-injection-tests"
  description = "Custom prompt injection test cases"
}
```

## Argument Reference

- `name` - (Required) Name of the prompt set.
- `description` - (Optional) Description of the prompt set.

## Attribute Reference

- `id` - The prompt set UUID.
- `uuid` - The prompt set UUID (same as `id`).
- `status` - Prompt set status.
- `active` - Whether the prompt set is active.
- `archive` - Whether the prompt set is archived.
- `created_at` - Timestamp when the prompt set was created.
- `updated_at` - Timestamp when the prompt set was last updated.

:::note

Deleting a custom prompt set archives it rather than permanently removing it, as the Red Team API does not support permanent deletion.

:::

## Import

```bash
terraform import prisma-airs_red_team_custom_prompt_set.injection_tests <uuid>
```

The legacy `properties` input is removed. SDK v0.6.1 does not provide the same operation; `property_names` is a separate service concept and is not translated automatically. Archived sets are removed from Terraform state on refresh and rejected during import.

Clearing a nonempty description plans replacement because the service retains the old value even when an update sends an explicit empty string. The previous set is archived, and a new set gets the empty description.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_red_team_custom_prompt_set.md) for all nested fields, types, and sensitivity flags.
