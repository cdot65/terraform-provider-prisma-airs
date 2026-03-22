# prisma-airs_security_profile

Manages an AI security profile in Prisma AIRS Management API.

## Example Usage

### Basic — Prompt Injection Protection

```hcl
resource "prisma-airs_security_profile" "basic" {
  profile_name = "basic-protection"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }
  }
}
```

### Full — Multiple Protections with Data Leak Detection

```hcl
resource "prisma-airs_security_profile" "full" {
  profile_name = "full-protection"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 25
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "block"
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action           = "block"
        mask_data_inline = true

        member {
          text = "sensitive content"
        }
      }
    }
  }
}
```

### With Topic-Based Detection

```hcl
resource "prisma-airs_security_profile" "topics" {
  profile_name = "topic-detection"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "allow"

      topic_list {
        action = "allow"

        topic {
          topic_name = "Recipe Generation"
          topic_id   = prisma-airs_custom_topic.recipes.topic_id
          revision   = 1
        }
      }

      topic_list {
        action = "block"

        topic {
          topic_name = "Restricted Content"
          topic_id   = prisma-airs_custom_topic.restricted.topic_id
          revision   = 1
        }
      }
    }
  }
}
```

## Argument Reference

### Top-Level

- `profile_name` - (Required) Name of the security profile.

### `ai_security_profile` Block

- `model_type` - (Optional) Model type (e.g., `"default"`).
- `content_type` - (Optional) Content type.
- `mask_data_in_storage` - (Optional) Whether to mask data in storage.

### `latency` Block (inside `ai_security_profile`)

- `inline_timeout_action` - (Optional) Action on inline timeout: `"allow"` or `"block"`.
- `max_inline_latency` - (Optional) Maximum inline latency in seconds.

### `model_protection` Block (inside `ai_security_profile`)

- `name` - (Required) Protection name. Values: `"prompt-injection"`, `"toxic-content"`, `"contextual-grounding"`, `"topic-guardrails"`.
- `action` - (Required) Action to take: `"block"` or `"allow"`.

### `toxic_category` Block (inside `model_protection`)

Per-category overrides for toxic content detection.

- `category` - (Required) Category name: `"harassment"`, `"violence"`, `"hate-speech"`, `"sexual-content"`.
- `action` - (Required) Action for this category: `"block"` or `"allow"`.

### `topic_list` Block (inside `model_protection`)

- `action` - (Required) Action for matched topics: `"block"` or `"allow"`.

### `topic` Block (inside `topic_list`)

- `topic_name` - (Required) Topic name.
- `topic_id` - (Optional) Topic ID.
- `revision` - (Optional) Topic revision.

### `agent_protection` Block (inside `ai_security_profile`)

- `name` - (Required) Protection name (e.g., `"agent-security"`).
- `action` - (Required) Action to take: `"block"` or `"allow"`.

### `app_protection` Block (inside `ai_security_profile`)

- `allow_url_category` - (Optional) List of URL categories to allow.
- `block_url_category` - (Optional) List of URL categories to block.

### `data_protection` Block (inside `ai_security_profile`)

Contains a `data_leak_detection` sub-block.

### `data_leak_detection` Block (inside `data_protection`)

- `action` - (Optional) Action on detection: `"block"` or `"allow"`.
- `mask_data_inline` - (Optional) Whether to mask detected data inline.

### `member` Block (inside `data_leak_detection`)

- `text` - (Required) Member text identifier.
- `id` - (Optional) Member ID.
- `version` - (Optional) Member version.

### `dlp_data_profile` Block

- `name` - (Optional) Profile name.
- `uuid` - (Optional) Profile UUID.
- `profile_id` - (Optional) Profile ID.
- `version` - (Optional) Profile version.
- `log_severity` - (Optional) Log severity level.
- `non_file_based` - (Optional) Non-file-based detection action.
- `file_based` - (Optional) File-based detection action.

## Attribute Reference

- `id` - The profile ID.
- `profile_id` - The profile ID (same as `id`).
- `active` - Whether the profile is active.
- `created_at` - Timestamp when the profile was created.
- `updated_at` - Timestamp when the profile was last updated.

## Import

Security profiles can be imported by name:

```bash
terraform import prisma-airs_security_profile.example "profile-name"
```
