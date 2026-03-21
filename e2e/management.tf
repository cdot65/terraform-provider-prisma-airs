# ---------------------------------------------------------------------------
# Management API Resources
# ---------------------------------------------------------------------------

# --- Security Profile ---
resource "prisma-airs_security_profile" "test" {
  profile_name = "${local.prefix}-profile"

  policy = jsonencode({
    "ai-security-profiles" = [
      {
        "model-type" = "default"
        "model-configuration" = {
          "latency" = {
            "inline-timeout-action" = "allow"
            "max-inline-latency"    = 30
          }
          "model-protection" = [
            {
              name   = "prompt-injection"
              action = "alert"
            },
            {
              name   = "url-filtering"
              action = "alert"
            },
            {
              name   = "toxic-content"
              action = "alert"
              "toxic-category-list" = [
                { category = "harassment", action = "alert" },
                { category = "violence", action = "alert" }
              ]
            }
          ]
        }
      }
    ]
  })
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
