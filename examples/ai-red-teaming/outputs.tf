# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

output "targets" {
  description = "Red team targets"
  value = {
    litellm_multiturn = {
      id     = prisma-airs_red_team_target.litellm_multiturn.uuid
      name   = prisma-airs_red_team_target.litellm_multiturn.name
      status = prisma-airs_red_team_target.litellm_multiturn.status
    }
    litellm_singleturn = {
      id     = prisma-airs_red_team_target.litellm_singleturn.uuid
      name   = prisma-airs_red_team_target.litellm_singleturn.name
      status = prisma-airs_red_team_target.litellm_singleturn.status
    }
    worf = {
      id     = prisma-airs_red_team_target.worf.uuid
      name   = prisma-airs_red_team_target.worf.name
      status = prisma-airs_red_team_target.worf.status
    }
    bedrock_claude = {
      id     = prisma-airs_red_team_target.bedrock_claude.uuid
      name   = prisma-airs_red_team_target.bedrock_claude.name
      status = prisma-airs_red_team_target.bedrock_claude.status
    }
    qwen_completions = {
      id     = prisma-airs_red_team_target.qwen_completions.uuid
      name   = prisma-airs_red_team_target.qwen_completions.name
      status = prisma-airs_red_team_target.qwen_completions.status
    }
    talkdesk = {
      id     = prisma-airs_red_team_target.talkdesk.uuid
      name   = prisma-airs_red_team_target.talkdesk.name
      status = prisma-airs_red_team_target.talkdesk.status
    }
  }
}

output "prompt_sets" {
  description = "Red team custom prompt sets"
  value = {
    general_adversarial = {
      id     = prisma-airs_red_team_custom_prompt_set.general_adversarial.uuid
      name   = prisma-airs_red_team_custom_prompt_set.general_adversarial.name
      status = prisma-airs_red_team_custom_prompt_set.general_adversarial.status
    }
    compliance = {
      id     = prisma-airs_red_team_custom_prompt_set.compliance.uuid
      name   = prisma-airs_red_team_custom_prompt_set.compliance.name
      status = prisma-airs_red_team_custom_prompt_set.compliance.status
    }
    agent_attacks = {
      id     = prisma-airs_red_team_custom_prompt_set.agent_attacks.uuid
      name   = prisma-airs_red_team_custom_prompt_set.agent_attacks.name
      status = prisma-airs_red_team_custom_prompt_set.agent_attacks.status
    }
  }
}
