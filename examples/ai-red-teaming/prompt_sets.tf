# Prompts: Create a container; populate attack prompts separately.
resource "prisma-airs_red_team_custom_prompt_set" "general_adversarial" {
  name        = "general-adversarial-prompts"
  description = "General adversarial testing prompts covering common attack vectors"
}

# Prompts: Create a container; populate attack prompts separately.
resource "prisma-airs_red_team_custom_prompt_set" "compliance" {
  name        = "compliance-testing-prompts"
  description = "Prompts targeting compliance and data protection bypasses"
}

# Prompts: Create a container; populate attack prompts separately.
resource "prisma-airs_red_team_custom_prompt_set" "agent_attacks" {
  name        = "agent-attack-prompts"
  description = "Prompts targeting AI agent-specific vulnerabilities"
}
