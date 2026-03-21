# prisma-airs_security_profile

Manages an AI security profile in Prisma AIRS Management API.

## Example Usage

```hcl
resource "prisma-airs_security_profile" "example" {
  profile_name = "my-ai-security-profile"

  policy = jsonencode({
    injection = {
      action = "block"
    }
    toxic_content = {
      action = "alert"
    }
    dlp = {
      action = "block"
    }
  })
}
```

## Argument Reference

- `profile_name` - (Required) Name of the security profile.
- `policy` - (Optional) JSON-encoded policy configuration.

## Attribute Reference

- `id` - The profile ID.
- `profile_id` - The profile ID (same as `id`).
- `active` - Whether the profile is active.
- `created_at` - Timestamp when the profile was created.
- `updated_at` - Timestamp when the profile was last updated.

## Import

Security profiles can be imported using the profile ID:

```bash
terraform import prisma-airs_security_profile.example <profile_id>
```
