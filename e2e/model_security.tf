# ---------------------------------------------------------------------------
# Model Security Resources
# ---------------------------------------------------------------------------

# --- Security Group ---
resource "prisma-airs_model_security_group" "test" {
  name        = "${local.prefix}-group"
  description = "E2E test security group"
  source_type = "HUGGING_FACE"
}

# ---------------------------------------------------------------------------
# Model Security Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_model_security_rules" "all" {}
