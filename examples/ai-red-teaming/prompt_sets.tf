# ---------------------------------------------------------------------------
# Custom Prompt Sets
# ---------------------------------------------------------------------------
# Define custom prompt sets for different red team assessment scenarios.
# Each set can be uploaded with adversarial prompts via the AIRS console
# after creation.
# ---------------------------------------------------------------------------

# ── General Adversarial Testing ───────────────────────────────────────
# Broad prompt set covering common attack vectors: jailbreaks, prompt
# injection, social engineering, role-play exploits, etc.

resource "prisma-airs_red_team_custom_prompt_set" "general_adversarial" {
  name        = "general-adversarial-prompts"
  description = "General adversarial testing prompts covering common attack vectors"
}

# ── Compliance-Focused Testing ────────────────────────────────────────
# Prompts targeting compliance violations: PII extraction, content filter
# bypasses, regulated content generation, HIPAA/PCI data leaks.

resource "prisma-airs_red_team_custom_prompt_set" "compliance" {
  name        = "compliance-testing-prompts"
  description = "Prompts targeting compliance and data protection bypasses"
}

# ── Agent-Specific Testing ────────────────────────────────────────────
# Prompts for AI agent vulnerabilities: tool abuse, unauthorized actions,
# chain-of-thought manipulation, privilege escalation via tool calls.

resource "prisma-airs_red_team_custom_prompt_set" "agent_attacks" {
  name        = "agent-attack-prompts"
  description = "Prompts targeting AI agent-specific vulnerabilities"
}
