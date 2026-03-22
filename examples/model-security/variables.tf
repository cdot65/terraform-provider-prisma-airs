# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
# Populate via terraform.tfvars (see terraform.tfvars.example)
# ---------------------------------------------------------------------------

variable "group_prefix" {
  type        = string
  default     = ""
  description = "Optional prefix for all group names (useful for testing)"
}
