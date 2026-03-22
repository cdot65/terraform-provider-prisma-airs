# ---------------------------------------------------------------------------
# Model Security — Resources
# ---------------------------------------------------------------------------

# --- Security Group ---
# Creates a model security group for monitoring AI models.
resource "prisma-airs_model_security_group" "main" {
  name        = "sandbox-group"
  description = "Security group for monitoring Hugging Face models"
  source_type = "HUGGING_FACE"
}

# ---------------------------------------------------------------------------
# Model Security — Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_model_security_rules" "all" {}
