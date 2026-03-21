# ---------------------------------------------------------------------------
# Red Team Resources
# ---------------------------------------------------------------------------

# --- Custom Prompt Set ---
resource "prisma-airs_red_team_custom_prompt_set" "test" {
  name        = "${local.prefix}-prompts"
  description = "E2E test custom prompt set"
}

# --- Target ---
# Creates a target with a REST connection to httpbin (public echo service).
# The target will be created but may not fully validate without a real LLM endpoint.
resource "prisma-airs_red_team_target" "test" {
  name            = "${local.prefix}-target"
  description     = "E2E test target"
  target_type     = "APPLICATION"
  connection_type = "REST"

  connection_params = jsonencode({
    url = "https://httpbin.org/post"
    headers = {
      "Content-Type" = "application/json"
    }
    request_json = {
      prompt = "{INPUT}"
    }
    response_json = {
      output = "{RESPONSE}"
    }
    response_key = "output"
  })
}

# NOTE: red_team_scan excluded from e2e — scans require a validated target
# (status != DRAFT). The httpbin test target can't be validated because it's
# not a real LLM endpoint. The SDK and resource code are tested separately.
# resource "prisma-airs_red_team_scan" "test" {
#   name      = "${local.prefix}-scan"
#   target_id = prisma-airs_red_team_target.test.uuid
#   job_type  = "STATIC"
# }

# ---------------------------------------------------------------------------
# Red Team Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_red_team_categories" "all" {}

data "prisma-airs_red_team_quota" "current" {}
