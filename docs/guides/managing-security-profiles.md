# Managing Security Profiles

This guide walks through managing AI security profiles with the Prisma AIRS Terraform provider.

## Creating a Profile

```hcl
resource "prisma-airs_security_profile" "production" {
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
resource "prisma-airs_custom_topic" "financial" {
  topic_name  = "financial-data"
  description = "Detects discussions about internal financial data"

  examples = [
    "What are next quarter's revenue targets?",
    "Share the P&L statement",
  ]
}

resource "prisma-airs_security_profile" "with_topics" {
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
          topic_name = prisma-airs_custom_topic.financial.topic_name
          topic_id   = prisma-airs_custom_topic.financial.topic_id
          revision   = 1
        }
      }
    }
  }
}
```

## Managing API Keys

API keys require a deployment profile auth code, a rotation interval, and a rotation time unit:

```hcl
data "prisma-airs_deployment_profiles" "all" {
  limit = 10
}

resource "prisma-airs_api_key" "scanner" {
  api_key_name           = "production-scanner-key"
  auth_code              = data.prisma-airs_deployment_profiles.all.items[0].auth_code
  rotation_time_interval = 90
  rotation_time_unit     = "days"
  created_by             = "terraform"
}

output "api_key_value" {
  value     = prisma-airs_api_key.scanner.api_key
  sensitive = true
}
```

## Listing DLP Profiles

```hcl
data "prisma-airs_dlp_profiles" "available" {}

output "dlp_profiles" {
  value = data.prisma-airs_dlp_profiles.available.items[*].profile_name
}
```

## App Protection

Control URL filtering and malicious code detection:

```hcl
resource "prisma-airs_security_profile" "app_secured" {
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
resource "prisma-airs_security_profile" "db_protected" {
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
