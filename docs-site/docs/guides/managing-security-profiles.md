# Managing Security Profiles

This guide walks through managing AI security profiles with the Prisma AIRS Terraform provider.

## Profile ownership

A resource owns all revisions under its `profile_name`. Policy edits create a new UUID/revision while keeping the Terraform address. Renaming follows version 1 of the new name and leaves the old history in AIRS. [Import and state](import-and-state.md) explains adoption and sensitive values.

:::warning[Destroy scope]

Destroy deletes every revision under the currently managed name, including history predating import or created by another actor. Confirm that ownership before importing a shared profile.

:::


## Directional inspection

Use `content_type_mode = "per_content_type"` and a `content_type_configurations` block with optional `prompt`, `response`, `tool_call`, and `tool_response` protection blocks. Each direction reuses `data_protection`, `app_protection`, `model_protection`, and `agent_protection`. Keep `latency`, `mask_data_in_storage`, and optional `enable_full_conversation_inspection` directly on `ai_security_profile`.

All layouts support detector `severity`, `url_detected_severity`, model and toxicity-category `severity_by_confidence { high, moderate }`, topic-reference severity, and `data_protection.source_code_detection { action, severity }`. Compound toxicity actions remain intact. Mode and severity strings are open-ended; known examples are not exhaustive.

See the [complete four-direction configuration](../examples/directional-security-profile.md) for the HCL interface, import, metadata, and preservation behavior. The existing legacy HCL paths and stored state remain compatible. Configure one protection layout per profile entry; a mixed service response remains readable, and conflicting configured writes receive a diagnostic.

The provider retains a lossless policy snapshot in Terraform private state. This preserves unknown direction keys, detector options, DLP rules, future fields, and omitted/null/empty distinctions through refresh and managed edits. Explicit detector removal stays removed. Changed lists with ambiguous repeated identities are rejected before mutation. Optional `dlp_tenant_id` is computed response metadata, and GET omission does not cause a configuration diff.

## Creating a Profile

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "production" {
  profile_name = "production-ai-security"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 30
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:block"
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

## Using Custom Topics

Create custom detection topics and reference them in profiles:

```hcl
# Topic: Describe the content that a security profile should recognize.
resource "prisma-airs_runtime_custom_topic" "financial" {
  topic_name  = "financial-data"
  description = "Detects discussions about internal financial data"

  examples = [
    "What are next quarter's revenue targets?",
    "Share the P&L statement",
  ]
}

# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "with_topics" {
  profile_name = "finance-team-profile"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "block"

      topic_list {
        action = "block"

        topic {
          topic_name = prisma-airs_runtime_custom_topic.financial.topic_name
        }
      }
    }
  }
}
```

:::tip[Auto-Resolution]

The provider automatically resolves `topic_name` to `topic_id` and `revision` — no need to specify them manually.

:::

## Managing API Keys

API keys require a deployment profile auth code, a rotation interval, and a rotation time unit:

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_runtime_deployment_profiles" "all" {
  limit = 10
}

# Scanning access: Issue a key using an authorized deployment profile.
resource "prisma-airs_runtime_api_key" "scanner" {
  api_key_name           = "production-scanner-key"
  auth_code              = data.prisma-airs_runtime_deployment_profiles.all.items[0].auth_code
  rotation_time_interval = 90
  rotation_time_unit     = "days"
  created_by             = "terraform"
}

# Outputs: Expose results for the next configuration or application step.
output "api_key_value" {
  value     = prisma-airs_runtime_api_key.scanner.api_key
  sensitive = true
}
```

Deleting an API key can also delete its associated customer app. Capture the one-time key securely; imported keys cannot recover it. See the [API-key lifecycle](../resources/api-key.md#replacement-and-secret-state).

## Listing DLP Profiles

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_runtime_dlp_profiles" "available" {}

# Outputs: Expose results for the next configuration or application step.
output "dlp_profiles" {
  value = data.prisma-airs_runtime_dlp_profiles.available.items[*].profile_name
}
```

## App Protection

Control URL filtering and malicious code detection:

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "app_secured" {
  profile_name = "app-protection-enabled"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    app_protection {
      default_url_category = ["malicious"]
      url_detected_action  = "block"

      malicious_code_protection {
        name   = "malicious-code"
        action = "block"
      }
    }
  }
}
```

For more granular URL category control, use `allow_url_category`, `block_url_category`, or `alert_url_category` lists:

```hcl
app_protection {
  allow_url_category = [
    "dynamic-dns",
    "grayware",
    "high-risk",
  ]

  url_detected_action = "block"
}
```

## Database Security

Control which database operations AI models can perform:

```hcl
# Policy: Configure inspection; policy edits create new profile revisions.
resource "prisma-airs_runtime_security_profile" "db_protected" {
  profile_name = "database-security-enabled"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action = "block"

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

This configuration allows read operations while blocking create, update, and delete — a common pattern for read-only AI assistants.
