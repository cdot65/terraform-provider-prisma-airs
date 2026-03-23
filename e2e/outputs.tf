# ---------------------------------------------------------------------------
# Outputs — used by run.sh to validate that everything was created
# ---------------------------------------------------------------------------

# ── Management Resources ────────────────────────────────────────────────

output "security_profile_id" {
  value = prisma-airs_security_profile.test.id
}

output "security_profile_name" {
  value = prisma-airs_security_profile.test.profile_name
}

output "custom_topic_id" {
  value = prisma-airs_custom_topic.test.id
}

output "custom_topic_name" {
  value = prisma-airs_custom_topic.test.topic_name
}

# ── Management Data Sources ─────────────────────────────────────────────

output "dlp_profile_count" {
  value = length(data.prisma-airs_dlp_profiles.all.items)
}

output "deployment_profile_count" {
  value = length(data.prisma-airs_deployment_profiles.all.items)
}

# ── Model Security Resources ───────────────────────────────────────────

output "model_security_group_id" {
  value = prisma-airs_model_security_group.test.id
}

output "model_security_group_name" {
  value = prisma-airs_model_security_group.test.name
}

output "model_security_group_state" {
  value = prisma-airs_model_security_group.test.state
}

# ── Model Security Data Sources ─────────────────────────────────────────

output "model_security_rule_count" {
  value = length(data.prisma-airs_model_security_rules.all.rules)
}

# ── Red Team Resources ──────────────────────────────────────────────────

output "red_team_prompt_set_id" {
  value = prisma-airs_red_team_custom_prompt_set.test.id
}

output "red_team_prompt_set_status" {
  value = prisma-airs_red_team_custom_prompt_set.test.status
}

output "red_team_target_id" {
  value = prisma-airs_red_team_target.test.id
}

output "red_team_target_status" {
  value = prisma-airs_red_team_target.test.status
}
