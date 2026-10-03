---
page_title: "prisma-airs_runtime_security_profile (Resource)"
subcategory: "AI Runtime Security"
---

# prisma-airs_runtime_security_profile Resource

Manages the complete revision history of one named AI security profile in Prisma AIRS Management API.

## Example Usage

### Basic — Prompt Injection Protection

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "basic" {
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
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "full" {
  profile_name = "full-protection"

  ai_security_profile {
    model_type           = "default"
    mask_data_in_storage = false

    latency {
      inline_timeout_action = "allow"
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

    app_protection {
      block_url_category  = ["malicious"]
      url_detected_action = "block"

      default_url_category = ["malicious"]

      malicious_code_protection {
        name   = "malicious-code"
        action = "block"
      }
    }

    data_protection {
      data_leak_detection {
        action           = "block"
        mask_data_inline = true

        member {
          text = "sensitive content"
        }
      }

      database_security {
        name   = "database-security-create"
        action = "block"
      }

      database_security {
        name   = "database-security-read"
        action = "allow"
      }

      database_security {
        name   = "database-security-update"
        action = "block"
      }

      database_security {
        name   = "database-security-delete"
        action = "block"
      }
    }
  }
}
```

### Compound Toxic Content Action

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "toxic" {
  profile_name = "toxic-compound"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:allow"
    }
  }
}
```

### With Topic-Based Detection

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "topics" {
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
        }
      }

      topic_list {
        action = "block"

        topic {
          topic_name = "Restricted Content"
        }
      }
    }
  }
}
```

**Auto-Resolution**
Since v0.6.2, the provider automatically resolves `topic_name` to `topic_id` and `revision` via the Topics API. You no longer need to specify these manually.


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
- `action` - (Required) Action to take. For most protections: `"block"` or `"allow"`. For `toxic-content`, also accepts compound `ToxicContentAction` values:
    - `"high:block, moderate:allow"` — block high-severity, allow moderate
    - `"high:block, moderate:block"` — block both severity levels
    - `"high:allow, moderate:allow"` — allow both severity levels

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

- `name` - (Required) Protection name: `"agent-security"`.
- `action` - (Required) Action to take: `"block"`.

### `app_protection` Block (inside `ai_security_profile`)

- `alert_url_category` - (Optional) List of URL categories to alert on.
- `allow_url_category` - (Optional) List of URL categories to allow.
- `block_url_category` - (Optional) List of URL categories to block.
- `default_url_category` - (Optional, Computed) List of default URL categories (e.g., `["malicious"]`).
- `url_detected_action` - (Optional, Computed) Action when a URL is detected: `"block"` or `""` (disabled).

### `malicious_code_protection` Block (inside `app_protection`)

- `name` - (Optional, Computed) Protection name (value: `"malicious-code"`).
- `action` - (Optional, Computed) Action to take: `"block"`.

### `data_protection` Block (inside `ai_security_profile`)

Contains `data_leak_detection` and `database_security` sub-blocks.

### `data_leak_detection` Block (inside `data_protection`)

- `action` - (Optional) Action on detection: `"block"` or allow (empty string disables).
- `mask_data_inline` - (Optional) Whether to mask detected data inline.

### `member` Block (inside `data_leak_detection`)

- `text` - (Required) Member text identifier.
- `id` - (Optional) Member ID.
- `version` - (Optional) Member version.

### `database_security` Block (inside `data_protection`)

- `name` - (Required) Database operation name: `"database-security-create"`, `"database-security-read"`, `"database-security-update"`, `"database-security-delete"`.
- `action` - (Required) Action to take: `"block"` or `"allow"`.

### `dlp_data_profile` Block

- `name` - (Optional) Profile name.
- `uuid` - (Optional) Profile UUID.
- `profile_id` - (Optional) Profile ID.
- `version` - (Optional) Profile version.
- `log_severity` - (Required) Log severity level.
- `non_file_based` - (Optional) Non-file-based detection action.
- `file_based` - (Optional) File-based detection action.

## Attribute Reference

- `id` - The profile ID.
- `profile_id` - The profile ID (same as `id`).
- `revision` - Highest numeric revision under the managed name.
- `active` - Whether the profile is active.
- `created_at` - Timestamp of revision 1, or null when that history is unavailable.
- `updated_at` - Timestamp when the profile was last updated.

## Import

Security profiles can be imported by name:

```bash
terraform import prisma-airs_runtime_security_profile.example "profile-name"
```

## Revision ownership and changes

Import by profile name selects the highest numeric revision, independently of its `active` flag. Refresh follows externally created newer revisions and falls back to remaining history when the latest UUID disappears. A conflicting create requires explicit import.

Policy changes create a new service UUID and revision while Terraform plans an ordinary update at the same address. Unchanged policy fields remain known; `profile_id`, `id`, and `revision` become known after apply. `created_at` uses revision 1's timestamp, or null when that original revision is unavailable.

Renaming creates revision 1 of the new name and leaves the old named profile in AIRS. Terraform follows the new name. An occupied destination is rejected; import it separately if ownership is intended.

**Destroy deletes every revision under the currently managed name**, including history predating import and changes by other actors. After a rename, the previous name requires separate management or cleanup.

Omitted `model_type` defaults to `"default"`. Optional scalar policy fields that have server defaults are read as computed values. Explicit `mask_data_inline = false` remains false through apply and refresh. DLP entries require nonempty `log_severity`; the live API rejects its omission.

Computed DLP reference metadata is preserved only while every configured name, UUID, profile ID and version still matches. Changing or reordering references leaves derived metadata unknown until the new revision is read, rather than carrying it from the prior list entry.

## Complete schema

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_runtime_security_profile/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `active` | `bool` | computed | — | Whether the profile is active. |
| `created_at` | `string` | computed | — | Timestamp of revision 1, when that revision remains available; null if original creation time is unavailable. |
| `id` | `string` | computed | — | Terraform resource ID (same as profile_id). |
| `profile_id` | `string` | computed | — | The unique identifier of the security profile. |
| `profile_name` | `string` | required | — | Name of the security profile. |
| `revision` | `number` | computed | — | Highest numeric revision of the managed profile name; changes on policy updates. |
| `updated_at` | `string` | computed | — | Last update timestamp. |

### ai_security_profile

Nesting: `list`.

#### ai_security_profile

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `content_type` | `string` | optional, computed | — | Content type. |
| `mask_data_in_storage` | `bool` | optional, computed | — | Whether to mask data in storage. |
| `model_type` | `string` | optional, computed | — | Model type (e.g., 'default'). |

#### ai_security_profile.agent_protection

Nesting: `list`.

##### ai_security_profile.agent_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action to take: 'block'. |
| `name` | `string` | required | — | Protection name: 'agent-security'. |

#### ai_security_profile.app_protection

Nesting: `single`.

##### ai_security_profile.app_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_url_category` | `list(string)` | optional | — | URL categories to alert on. |
| `allow_url_category` | `list(string)` | optional | — | URL categories to allow. |
| `block_url_category` | `list(string)` | optional | — | URL categories to block. |
| `default_url_category` | `list(string)` | optional, computed | — | Default URL categories (e.g., 'malicious'). |
| `url_detected_action` | `string` | optional, computed | — | Action when a URL matches configured categories: 'block' or empty to disable. |

##### ai_security_profile.app_protection.malicious_code_protection

Nesting: `single`.

###### ai_security_profile.app_protection.malicious_code_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | optional, computed | — | Action to take: 'block' or 'allow'. |
| `name` | `string` | optional, computed | — | Protection name (e.g., 'malicious-code'). |

#### ai_security_profile.data_protection

Nesting: `single`.

##### ai_security_profile.data_protection.data_leak_detection

Nesting: `single`.

###### ai_security_profile.data_protection.data_leak_detection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | optional, computed | — | Action on detection: 'block' or 'allow'. |
| `mask_data_inline` | `bool` | optional, computed | — | Whether to mask detected data inline. |

###### ai_security_profile.data_protection.data_leak_detection.member

Nesting: `list`.

###### ai_security_profile.data_protection.data_leak_detection.member

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | optional, computed | — | Member ID. |
| `text` | `string` | required | — | Member text identifier. |
| `version` | `string` | optional, computed | — | Member version. |

##### ai_security_profile.data_protection.database_security

Nesting: `list`.

###### ai_security_profile.data_protection.database_security

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action: 'block' or 'allow'. |
| `name` | `string` | required | — | Database operation name (e.g., 'database-security-create'). |

#### ai_security_profile.latency

Nesting: `single`.

##### ai_security_profile.latency

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `inline_timeout_action` | `string` | optional, computed | — | Action on inline timeout ('allow' or 'block'). |
| `max_inline_latency` | `number` | optional, computed | — | Maximum inline latency in seconds. |

#### ai_security_profile.model_protection

Nesting: `list`.

##### ai_security_profile.model_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action to take: 'block', 'allow', or compound toxic-content values like 'high:block, moderate:allow'. |
| `name` | `string` | required | — | Protection name: 'prompt-injection', 'toxic-content', 'contextual-grounding', or 'topic-guardrails'. |

##### ai_security_profile.model_protection.topic_list

Nesting: `list`.

###### ai_security_profile.model_protection.topic_list

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action for matched topics. |

###### ai_security_profile.model_protection.topic_list.topic

Nesting: `list`.

###### ai_security_profile.model_protection.topic_list.topic

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `revision` | `number` | optional, computed | — | Topic revision. |
| `topic_id` | `string` | optional, computed | — | Topic ID. |
| `topic_name` | `string` | required | — | Topic name. |

##### ai_security_profile.model_protection.toxic_category

Nesting: `list`.

###### ai_security_profile.model_protection.toxic_category

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action for this category. |
| `category` | `string` | required | — | Category name (e.g., 'harassment', 'violence', 'hate-speech', 'sexual-content'). |

### dlp_data_profile

Nesting: `list`.

#### dlp_data_profile

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `file_based` | `string` | optional, computed | — | File-based detection action. |
| `log_severity` | `string` | required | — | Log severity level. |
| `name` | `string` | optional, computed | — | Profile name. |
| `non_file_based` | `string` | optional, computed | — | Non-file-based detection action. |
| `profile_id` | `string` | optional, computed | — | Profile ID. |
| `uuid` | `string` | optional, computed | — | Profile UUID. |
| `version` | `string` | optional, computed | — | Profile version. |
