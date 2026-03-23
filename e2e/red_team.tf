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
