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

```hcl
resource "prisma-airs_api_key" "scanner" {
  api_key_name = "production-scanner-key"
  created_by   = "terraform"
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
