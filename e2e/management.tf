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
          "model-protection" = [
            {
              name   = "prompt-injection"
              action = "alert"
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

# NOTE: api_key excluded from e2e — creating an API key requires a customer
# app with no existing key association. The upstream customer app creation
# endpoint hangs indefinitely, so we can't provision a fresh app for this test.
# resource "prisma-airs_api_key" "test" {
#   api_key_name           = "${local.prefix}-key"
#   auth_code              = data.prisma-airs_deployment_profiles.all.items[0].auth_code
#   cust_app               = "${local.prefix}-app"
#   created_by             = "terraform-e2e"
#   rotation_time_interval = 90
#   rotation_time_unit     = "days"
# }

# NOTE: customer_app resource excluded from e2e — the upstream API endpoint
# hangs indefinitely on POST /v1/mgmt/customerapp. The SDK has no
# integration test for CustomerApps.Create.
# resource "prisma-airs_customer_app" "test" {
#   app_name       = "${local.prefix}-app"
#   cloud_provider = "aws"
#   environment    = "development"
# }

# ---------------------------------------------------------------------------
# Management API Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_dlp_profiles" "all" {
  limit = 10
}

data "prisma-airs_deployment_profiles" "all" {
  limit = 10
}

# NOTE: scan_logs data source excluded from e2e — the upstream API endpoint
# returns "invalid time duration specified" consistently. The SDK's own
# integration test also skips this endpoint (t.Skip).
# data "prisma-airs_scan_logs" "recent" {
#   limit         = 5
#   time_interval = 24
#   time_unit     = "hour"
#   filter        = "all"
# }
