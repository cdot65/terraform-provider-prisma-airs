# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

# ── Management ─────────────────────────────────────────────────────────────

output "security_profile" {
  description = "Security profile details"
  value = {
    id     = prisma-airs_security_profile.main.profile_id
    name   = prisma-airs_security_profile.main.profile_name
    active = prisma-airs_security_profile.main.active
  }
}

output "custom_topic" {
  description = "Custom topic details"
  value = {
    id   = prisma-airs_custom_topic.main.topic_id
    name = prisma-airs_custom_topic.main.topic_name
  }
}

# Customer app output — uncomment after importing an existing app:
# output "customer_app" {
#   description = "Customer app details"
#   value = {
#     id     = prisma-airs_customer_app.main.customer_app_id
#     name   = prisma-airs_customer_app.main.app_name
#     status = prisma-airs_customer_app.main.status
#   }
# }

output "dlp_profiles" {
  description = "Available DLP profiles"
  value       = [for p in data.prisma-airs_dlp_profiles.all.items : p.profile_name]
}

output "deployment_profiles" {
  description = "Available deployment profiles"
  value       = [for p in data.prisma-airs_deployment_profiles.all.items : p.profile_name]
}

# ── Model Security ─────────────────────────────────────────────────────────

output "model_security_group" {
  description = "Model security group details"
  value = {
    id    = prisma-airs_model_security_group.main.uuid
    name  = prisma-airs_model_security_group.main.name
    state = prisma-airs_model_security_group.main.state
  }
}

output "model_security_rule_count" {
  description = "Number of model security rules"
  value       = length(data.prisma-airs_model_security_rules.all.rules)
}

# ── Red Team ───────────────────────────────────────────────────────────────

output "red_team_prompt_set" {
  description = "Red team custom prompt set details"
  value = {
    id     = prisma-airs_red_team_custom_prompt_set.main.uuid
    name   = prisma-airs_red_team_custom_prompt_set.main.name
    status = prisma-airs_red_team_custom_prompt_set.main.status
  }
}

output "red_team_target" {
  description = "Red team target details"
  value = {
    id     = prisma-airs_red_team_target.main.uuid
    name   = prisma-airs_red_team_target.main.name
    status = prisma-airs_red_team_target.main.status
  }
}
