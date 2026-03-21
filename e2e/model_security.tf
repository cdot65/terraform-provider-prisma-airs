# ---------------------------------------------------------------------------
# Model Security Resources
# ---------------------------------------------------------------------------

# --- Security Group ---
resource "prisma-airs_model_security_group" "test" {
  name        = "${local.prefix}-group"
  description = "E2E test security group"
  source_type = "HUGGING_FACE"
}

# --- Model Scan ---
# Scans a tiny HuggingFace model. The scan may remain PENDING — that's OK.
# Scans cannot be deleted via API; they are removed from state on destroy.
resource "prisma-airs_model_scan" "test" {
  model_uri           = "https://huggingface.co/prajjwal1/bert-tiny"
  scan_origin         = "HUGGING_FACE"
  security_group_uuid = prisma-airs_model_security_group.test.uuid

  labels = {
    test = "e2e"
    env  = "ci"
  }
}

# ---------------------------------------------------------------------------
# Model Security Data Sources
# ---------------------------------------------------------------------------

data "prisma-airs_model_security_rules" "all" {}

# Evaluations and violations for the scan we just created.
# May return empty lists if the scan hasn't completed yet.
data "prisma-airs_model_scan_evaluations" "test" {
  scan_uuid = prisma-airs_model_scan.test.uuid
}

data "prisma-airs_model_scan_violations" "test" {
  scan_uuid = prisma-airs_model_scan.test.uuid
}
