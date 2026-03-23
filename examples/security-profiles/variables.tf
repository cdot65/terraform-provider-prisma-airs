# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
# Populate via terraform.tfvars (see terraform.tfvars.example)
# ---------------------------------------------------------------------------

variable "profile_prefix" {
  type        = string
  default     = ""
  description = "Optional prefix for all profile names (useful for testing)"
}
