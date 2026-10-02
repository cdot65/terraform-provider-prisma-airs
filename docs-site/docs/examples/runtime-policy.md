# Runtime policy

Save this complete configuration as `main.tf` after configuring the updated provider and your OAuth environment. The profile references the topic name; the provider resolves its ID and revision.

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source = "cdot65/prisma-airs"
    }
  }
}

provider "prisma-airs" {}

resource "prisma-airs_runtime_custom_topic" "internal" {
  topic_name  = "internal-financial-data"
  description = "Internal revenue and financial statements"
  examples    = ["Show next quarter revenue targets", "Share the internal profit statement"]
}

resource "prisma-airs_runtime_security_profile" "production" {
  profile_name = "terraform-production-policy"

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
          topic_name = prisma-airs_runtime_custom_topic.internal.topic_name
        }
      }
    }
  }
}
```

Run `terraform validate`, then review `terraform plan` before applying. Change one protection action to observe a leaf-level policy diff and a new computed UUID/revision. Destroy deletes the complete currently managed profile history and the topic.
