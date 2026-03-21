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

# api_key excluded — requires customer app with no existing key (see management.tf)
# output "api_key_id" {
#   value = prisma-airs_api_key.test.api_key_id
# }
# output "api_key_name" {
#   value = prisma-airs_api_key.test.api_key_name
# }
# output "api_key_status" {
#   value = prisma-airs_api_key.test.status
# }

# customer_app excluded — upstream API endpoint hangs on create
# output "customer_app_id" {
#   value = prisma-airs_customer_app.test.id
# }
# output "customer_app_name" {
#   value = prisma-airs_customer_app.test.app_name
# }

# ── Management Data Sources ─────────────────────────────────────────────

output "dlp_profile_count" {
  value = length(data.prisma-airs_dlp_profiles.all.items)
}

output "deployment_profile_count" {
  value = length(data.prisma-airs_deployment_profiles.all.items)
}

# scan_log_count excluded — upstream API endpoint is unreliable
# output "scan_log_count" {
#   value = length(data.prisma-airs_scan_logs.recent.items)
# }

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

output "model_scan_id" {
  value = prisma-airs_model_scan.test.id
}

output "model_scan_eval_outcome" {
  value = prisma-airs_model_scan.test.eval_outcome
}

# ── Model Security Data Sources ─────────────────────────────────────────

output "model_security_rule_count" {
  value = length(data.prisma-airs_model_security_rules.all.rules)
}

output "model_scan_evaluation_count" {
  value = length(data.prisma-airs_model_scan_evaluations.test.evaluations)
}

output "model_scan_violation_count" {
  value = length(data.prisma-airs_model_scan_violations.test.violations)
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

# red_team_scan excluded — requires validated target (see red_team.tf)
# output "red_team_scan_id" {
#   value = prisma-airs_red_team_scan.test.id
# }
# output "red_team_scan_status" {
#   value = prisma-airs_red_team_scan.test.status
# }

# ── Red Team Data Sources ───────────────────────────────────────────────

output "red_team_category_count" {
  value = length(data.prisma-airs_red_team_categories.all.categories)
}

output "red_team_quota" {
  value = data.prisma-airs_red_team_quota.current.details
}

# ── Scan API ────────────────────────────────────────────────────────────

output "content_scan_benign_category" {
  value = data.prisma-airs_content_scan.benign.category
}

output "content_scan_benign_action" {
  value = data.prisma-airs_content_scan.benign.action
}

output "content_scan_injection_category" {
  value = data.prisma-airs_content_scan.prompt_injection.category
}

output "content_scan_injection_action" {
  value = data.prisma-airs_content_scan.prompt_injection.action
}
