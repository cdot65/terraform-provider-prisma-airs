# ---------------------------------------------------------------------------
# Management API — Resources
# ---------------------------------------------------------------------------

# --- Security Profile ---
# Configures AI security detection with latency, prompt injection,
# toxic content detection, agent protection, and data leak detection.
resource "prisma-airs_security_profile" "main" {
  profile_name = "sandbox-profile"

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

# --- Custom Topic ---
# Defines a custom content detection topic with training examples.
resource "prisma-airs_custom_topic" "main" {
  topic_name  = "sandbox-topic"
  description = "Custom topic for detecting proprietary content"
  examples = [
    "Our Q4 revenue projections show significant growth in the APAC region.",
    "The internal product roadmap includes a new AI-powered feature launch in September.",
    "Employee stock option vesting schedules are outlined in section 4.2 of the handbook.",
  ]
}

# --- Customer App ---
# Customer apps are created externally (AIRS console / app registration).
# Use `terraform import` to bring an existing app under management:
#   terraform import prisma-airs_customer_app.main <app_name>
#
# resource "prisma-airs_customer_app" "main" {
#   app_name = "my-existing-app"
# }

# --- API Key ---
# Creates an API key for the registered app.
# Requires a valid auth_code from a deployment profile.
# Uncomment after you have a deployment profile auth code.
#
# resource "prisma-airs_api_key" "main" {
#   api_key_name           = "sandbox-key"
#   auth_code              = data.prisma-airs_deployment_profiles.all.items[0].auth_code
#   cust_app               = prisma-airs_customer_app.main.app_name
#   created_by             = "terraform-sandbox"
#   rotation_time_interval = 90
#   rotation_time_unit     = "days"
# }

# ---------------------------------------------------------------------------
# Management API — Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_dlp_profiles" "all" {
  limit = 50
}

data "prisma-airs_deployment_profiles" "all" {
  limit = 50
}
