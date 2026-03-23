# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

output "security_groups" {
  description = "Model security groups"
  value = {
    hugging_face = {
      id    = prisma-airs_model_security_group.hugging_face.uuid
      name  = prisma-airs_model_security_group.hugging_face.name
      state = prisma-airs_model_security_group.hugging_face.state
    }
    custom_models = {
      id    = prisma-airs_model_security_group.custom_models.uuid
      name  = prisma-airs_model_security_group.custom_models.name
      state = prisma-airs_model_security_group.custom_models.state
    }
  }
}

output "rule_count" {
  description = "Total model security rules"
  value       = length(data.prisma-airs_model_security_rules.all.rules)
}
