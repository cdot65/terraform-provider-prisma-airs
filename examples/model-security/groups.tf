# ---------------------------------------------------------------------------
# Model Security Groups
# ---------------------------------------------------------------------------
# Security groups organize AI models by source for monitoring and
# threat detection. Each group tracks models from a specific provider.
# ---------------------------------------------------------------------------

# ── Hugging Face Models ───────────────────────────────────────────────
# Monitor models sourced from Hugging Face Hub for supply chain attacks,
# malicious weights, and model poisoning.

resource "prisma-airs_model_security_group" "hugging_face" {
  name        = "${var.group_prefix}hugging-face-models"
  description = "Security group for monitoring Hugging Face models"
  source_type = "HUGGING_FACE"
}

# ── Custom Models ─────────────────────────────────────────────────────
# Monitor internally-trained or fine-tuned models uploaded from
# custom sources.

resource "prisma-airs_model_security_group" "custom_models" {
  name        = "${var.group_prefix}custom-trained-models"
  description = "Security group for internally-trained models"
  source_type = "HUGGING_FACE"
}
