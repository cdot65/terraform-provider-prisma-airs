# prisma-airs_runtime_security_profile schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an AI Runtime Security profile.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `active` | `bool` | computed | — | Whether the profile is active. |
| `created_at` | `string` | computed | — | Timestamp of revision 1, when that revision remains available; null if original creation time is unavailable. |
| `id` | `string` | computed | — | Terraform resource ID (same as profile_id). |
| `profile_id` | `string` | computed | — | The unique identifier of the security profile. |
| `profile_name` | `string` | required | — | Name of the security profile. |
| `revision` | `number` | computed | — | Highest numeric revision of the managed profile name; changes on policy updates. |
| `updated_at` | `string` | computed | — | Last update timestamp. |

## ai_security_profile

Nesting: `list`.

### ai_security_profile

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `content_type` | `string` | optional, computed | — | Content type. |
| `mask_data_in_storage` | `bool` | optional, computed | — | Whether to mask data in storage. |
| `model_type` | `string` | optional, computed | — | Model type (e.g., 'default'). |

### ai_security_profile.agent_protection

Nesting: `list`.

#### ai_security_profile.agent_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action to take: 'block'. |
| `name` | `string` | required | — | Protection name: 'agent-security'. |

### ai_security_profile.app_protection

Nesting: `single`.

#### ai_security_profile.app_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `alert_url_category` | `list(string)` | optional | — | URL categories to alert on. |
| `allow_url_category` | `list(string)` | optional | — | URL categories to allow. |
| `block_url_category` | `list(string)` | optional | — | URL categories to block. |
| `default_url_category` | `list(string)` | optional, computed | — | Default URL categories (e.g., 'malicious'). |
| `url_detected_action` | `string` | optional, computed | — | Action when a URL matches configured categories: 'block' or empty to disable. |

#### ai_security_profile.app_protection.malicious_code_protection

Nesting: `single`.

##### ai_security_profile.app_protection.malicious_code_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | optional, computed | — | Action to take: 'block' or 'allow'. |
| `name` | `string` | optional, computed | — | Protection name (e.g., 'malicious-code'). |

### ai_security_profile.data_protection

Nesting: `single`.

#### ai_security_profile.data_protection.data_leak_detection

Nesting: `single`.

##### ai_security_profile.data_protection.data_leak_detection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | optional, computed | — | Action on detection: 'block' or 'allow'. |
| `mask_data_inline` | `bool` | optional, computed | — | Whether to mask detected data inline. |

##### ai_security_profile.data_protection.data_leak_detection.member

Nesting: `list`.

###### ai_security_profile.data_protection.data_leak_detection.member

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `id` | `string` | optional, computed | — | Member ID. |
| `text` | `string` | required | — | Member text identifier. |
| `version` | `string` | optional, computed | — | Member version. |

#### ai_security_profile.data_protection.database_security

Nesting: `list`.

##### ai_security_profile.data_protection.database_security

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action: 'block' or 'allow'. |
| `name` | `string` | required | — | Database operation name (e.g., 'database-security-create'). |

### ai_security_profile.latency

Nesting: `single`.

#### ai_security_profile.latency

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `inline_timeout_action` | `string` | optional, computed | — | Action on inline timeout ('allow' or 'block'). |
| `max_inline_latency` | `number` | optional, computed | — | Maximum inline latency in seconds. |

### ai_security_profile.model_protection

Nesting: `list`.

#### ai_security_profile.model_protection

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action to take: 'block', 'allow', or compound toxic-content values like 'high:block, moderate:allow'. |
| `name` | `string` | required | — | Protection name: 'prompt-injection', 'toxic-content', 'contextual-grounding', or 'topic-guardrails'. |

#### ai_security_profile.model_protection.topic_list

Nesting: `list`.

##### ai_security_profile.model_protection.topic_list

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action for matched topics. |

##### ai_security_profile.model_protection.topic_list.topic

Nesting: `list`.

###### ai_security_profile.model_protection.topic_list.topic

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `revision` | `number` | optional, computed | — | Topic revision. |
| `topic_id` | `string` | optional, computed | — | Topic ID. |
| `topic_name` | `string` | required | — | Topic name. |

#### ai_security_profile.model_protection.toxic_category

Nesting: `list`.

##### ai_security_profile.model_protection.toxic_category

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `action` | `string` | required | — | Action for this category. |
| `category` | `string` | required | — | Category name (e.g., 'harassment', 'violence', 'hate-speech', 'sexual-content'). |

## dlp_data_profile

Nesting: `list`.

### dlp_data_profile

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `file_based` | `string` | optional, computed | — | File-based detection action. |
| `log_severity` | `string` | required | — | Log severity level. |
| `name` | `string` | optional, computed | — | Profile name. |
| `non_file_based` | `string` | optional, computed | — | Non-file-based detection action. |
| `profile_id` | `string` | optional, computed | — | Profile ID. |
| `uuid` | `string` | optional, computed | — | Profile UUID. |
| `version` | `string` | optional, computed | — | Profile version. |
