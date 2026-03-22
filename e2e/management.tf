# ---------------------------------------------------------------------------
# Management API Resources
# ---------------------------------------------------------------------------

# --- Security Profile ---
resource "prisma-airs_security_profile" "test" {
  profile_name = "${local.prefix}-profile"

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
      action = "block"
    }
  }
}

# --- Custom Topic ---
resource "prisma-airs_custom_topic" "test" {
  topic_name  = "${local.prefix}-topic"
  description = "E2E test custom topic for content detection"
  examples = [
    "This is a test example for the custom topic.",
    "Another example sentence to train detection.",
  ]
}

# ---------------------------------------------------------------------------
# Management API Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_dlp_profiles" "all" {
  limit = 10
}

data "prisma-airs_deployment_profiles" "all" {
  limit = 10
}
