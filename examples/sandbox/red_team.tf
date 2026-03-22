# ---------------------------------------------------------------------------
# Red Team — Resources
# ---------------------------------------------------------------------------

# --- Custom Prompt Set ---
# Creates a custom prompt set for red team testing.
resource "prisma-airs_red_team_custom_prompt_set" "main" {
  name        = "sandbox-prompts"
  description = "Custom prompt set for adversarial testing"
}

# --- Target ---
# Registers an AI application as a red team testing target.
# Uses httpbin as a safe public endpoint for testing.
resource "prisma-airs_red_team_target" "main" {
  name            = "sandbox-target"
  description     = "Test target using httpbin echo service"
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
