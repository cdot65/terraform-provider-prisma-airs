# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

output "profiles" {
  description = "All managed security profiles"
  value = {
    high_security = {
      id     = prisma-airs_security_profile.high_security.profile_id
      name   = prisma-airs_security_profile.high_security.profile_name
      active = prisma-airs_security_profile.high_security.active
    }
    truffles_agent = {
      id     = prisma-airs_security_profile.truffles_agent.profile_id
      name   = prisma-airs_security_profile.truffles_agent.profile_name
      active = prisma-airs_security_profile.truffles_agent.active
    }
    recipe_extractor = {
      id     = prisma-airs_security_profile.recipe_extractor.profile_id
      name   = prisma-airs_security_profile.recipe_extractor.profile_name
      active = prisma-airs_security_profile.recipe_extractor.active
    }
    cursor_ide = {
      id     = prisma-airs_security_profile.cursor_ide.profile_id
      name   = prisma-airs_security_profile.cursor_ide.profile_name
      active = prisma-airs_security_profile.cursor_ide.active
    }
    slack_moderation = {
      id     = prisma-airs_security_profile.slack_moderation.profile_id
      name   = prisma-airs_security_profile.slack_moderation.profile_name
      active = prisma-airs_security_profile.slack_moderation.active
    }
    hipaa_compliance = {
      id     = prisma-airs_security_profile.hipaa_compliance.profile_id
      name   = prisma-airs_security_profile.hipaa_compliance.profile_name
      active = prisma-airs_security_profile.hipaa_compliance.active
    }
  }
}
