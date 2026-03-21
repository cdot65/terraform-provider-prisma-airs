# Managing Security Profiles

This guide walks through managing AI security profiles with the Prisma AIRS Terraform provider.

## Creating a Profile

```hcl
resource "prisma-airs_security_profile" "production" {
  profile_name = "production-ai-security"

  policy = jsonencode({
    injection = {
      action = "block"
    }
    toxic_content = {
      action = "block"
    }
    dlp = {
      action = "alert"
    }
    url_cats = {
      action = "alert"
    }
  })
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

  policy = jsonencode({
    topic_violation = {
      action = "block"
      topics = [prisma-airs_custom_topic.financial.topic_id]
    }
  })
}
```

## Managing API Keys

```hcl
resource "prisma-airs_api_key" "scanner" {
  api_key_name = "production-scanner-key"
  updated_by   = "terraform"
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
