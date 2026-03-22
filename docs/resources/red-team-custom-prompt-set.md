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
- `properties` - (Optional) Properties as a JSON-encoded string.

## Attribute Reference

- `id` - The prompt set UUID.
- `uuid` - The prompt set UUID (same as `id`).
- `status` - Prompt set status.
- `active` - Whether the prompt set is active.
- `archive` - Whether the prompt set is archived.
- `created_at` - Timestamp when the prompt set was created.
- `updated_at` - Timestamp when the prompt set was last updated.

!!! note
    Deleting a custom prompt set archives it rather than permanently removing it, as the Red Team API does not support permanent deletion.

## Import

```bash
terraform import prisma-airs_red_team_custom_prompt_set.injection_tests <uuid>
```
