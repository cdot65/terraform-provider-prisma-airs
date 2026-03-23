# ---------------------------------------------------------------------------
# Security Profiles
# ---------------------------------------------------------------------------
# Each profile maps to a real-world use case with appropriate protection
# levels. Profiles range from high-security (all protections enabled,
# strict blocking) to lightweight (minimal protections for low-risk apps).
# ---------------------------------------------------------------------------

# ── High Security Profile ─────────────────────────────────────────────
# Maximum protection: blocks prompt injection, toxic content, data leaks,
# agent abuse, and malicious URLs. Includes DLP with IP address detection
# and URL category filtering for high-risk categories.

resource "prisma-airs_security_profile" "high_security" {
  profile_name = "${var.profile_prefix}AI-Firewall-High-Security-Profile"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 5
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:block"
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action         = "block"
        mask_data_inline = true

        member {
          text = "sensitive content"
        }
      }
    }

    app_protection {
      allow_url_category = [
        "dynamic-dns",
        "grayware",
        "abused-drugs",
        "adult",
        "encrypted-dns",
        "high-risk",
        "phishing",
        "sports",
      ]
    }
  }

  dlp_data_profile {
    name         = "IP Addresses"
    profile_id   = "11995029"
    version      = "1"
    log_severity = "low"
  }
}

# ── Agent App Profile (Truffles) ──────────────────────────────────────
# Balanced profile for an AI agent: blocks prompt injection, allows moderate
# toxic content through, masks data inline, and uses topic guardrails to
# allow recipe generation while blocking ASCII art abuse.

resource "prisma-airs_security_profile" "truffles_agent" {
  profile_name = "${var.profile_prefix}Truffles Agent"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 25
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:allow"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "allow"

      topic_list {
        action = "allow"

        topic {
          topic_name = "Recipe Generation"
        }
      }

      topic_list {
        action = "block"

        topic {
          topic_name = "ASCII Art Generation"
        }
      }
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action         = "block"
        mask_data_inline = true

        member {
          text = "sensitive content"
        }
      }
    }
  }
}

# ── Recipe Extractor Profile ──────────────────────────────────────────
# Moderate profile for a recipe extraction agent on AWS. Allows moderate
# toxic content, uses topic guardrails to permit recipe discussions,
# and blocks prompt injection.

resource "prisma-airs_security_profile" "recipe_extractor" {
  profile_name = "${var.profile_prefix}Recipe Extractor AWS Agent"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "allow"
      max_inline_latency    = 5
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:block"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "allow"

      topic_list {
        action = "allow"

        topic {
          topic_name = "Recipe Recommendations and Creation"
        }
      }

      topic_list {
        action = "block"
      }
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }
  }
}

# ── IDE Integration Profile (Cursor) ─────────────────────────────────
# Strict profile for code-assistant IDE integrations. Blocks all model
# threats with strict toxic content filtering. Includes malicious code
# detection via app protection.

resource "prisma-airs_security_profile" "cursor_ide" {
  profile_name = "${var.profile_prefix}Cursor IDE - Hooks"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 25
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:block"
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action = "block"

        member {
          text = "sensitive content"
        }
      }
    }
  }
}

# ── Slack Moderation Profile ──────────────────────────────────────────
# Profile for the OpenClaw Slack bot. Blocks prompt injection, filters
# malicious URLs, and enables data protection. Latency timeout set to
# allow to avoid blocking chat responses.
#
# NOTE: The API's default-url-category and url-detected-action fields
# are not yet supported by the SDK/provider. Those must be set manually
# or added in a future SDK release.

resource "prisma-airs_security_profile" "slack_moderation" {
  profile_name = "${var.profile_prefix}OpenClaw - Slack Moderation"

  ai_security_profile {
    model_type        = "default"
    mask_data_in_storage = false

    latency {
      inline_timeout_action = "allow"
      max_inline_latency    = 5
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    app_protection {
      block_url_category = ["malicious"]
    }

    data_protection {
      data_leak_detection {
        action = ""
      }
    }
  }
}

# ── HIPAA Compliance Profile ──────────────────────────────────────────
# Healthcare-focused profile with HIPAA DLP data profile attached.
# Topic guardrails allow specific safe topics (Star Wars debates) while
# blocking everything else by default.

resource "prisma-airs_security_profile" "hipaa_compliance" {
  profile_name = "${var.profile_prefix}OpenClaw AI Agent Security Nightmare"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 5
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "toxic-content"
      action = "high:block, moderate:allow"
    }

    model_protection {
      name   = "topic-guardrails"
      action = "allow"

      topic_list {
        action = "allow"

        topic {
          topic_name = "Star Wars vs Star Trek Superiority Claims"
        }
      }

      topic_list {
        action = "block"
      }
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
    }

    data_protection {
      data_leak_detection {
        action = "block"

        member {
          text    = "HIPAA"
          id      = "11995010"
          version = "1"
        }
      }
    }
  }

  dlp_data_profile {
    name         = "HIPAA"
    profile_id   = "11995010"
    version      = "1"
    log_severity = "low"
  }
}
