# ---------------------------------------------------------------------------
# Data Sources
# ---------------------------------------------------------------------------
# Read existing model security rules to understand the current
# security posture and integrate with other tooling.
# ---------------------------------------------------------------------------

data "prisma-airs_model_security_rules" "all" {}
