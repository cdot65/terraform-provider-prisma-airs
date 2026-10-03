# Groups: Create two independently named Hugging Face collections.
resource "prisma-airs_supply_chain_security_group" "hugging_face" {
  name        = "${var.group_prefix}hugging-face-models"
  description = "Security group for monitoring Hugging Face models"
  source_type = "HUGGING_FACE"
}

resource "prisma-airs_supply_chain_security_group" "custom_models" {
  name        = "${var.group_prefix}custom-trained-models"
  description = "Security group for internally-trained models"
  source_type = "HUGGING_FACE"
}
