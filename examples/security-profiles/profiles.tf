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
  profile_name = "${var.profile_prefix}InfoSec - AI Firewall - Strict"

  ai_security_profile {
    model_type = "default"

    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 1
    }

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }

    model_protection {
      name   = "contextual-grounding"
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
      }

      topic_list {
        action = "block"

        topic {
          topic_name = "Deletion and Destruction of Cloud Infrastructure"
        }

        topic {
          topic_name = "Tax Evasion Techniques"
        }

        topic {
          topic_name = "Illegal Weapons Manufacturing and Procurement"
        }

        topic {
          topic_name = "Home Manufacturing of Illegal Drugs"
        }

        topic {
          topic_name = "Star Wars vs Star Trek Superiority Claims"
        }

        topic {
          topic_name = "ASCII Art Generation"
        }

        topic {
          topic_name = "Weapons Manufacturing and Procurement"
        }

        topic {
          topic_name = "Building Explosives"
        }

        topic {
          topic_name = "Tax Guidance and Recommendations"
        }

        topic {
          topic_name = "Explosives and Bomb-Making Discussions"
        }

        topic {
          topic_name = "Offensive Military Operation Planning Against Iran"
        }

        topic {
          topic_name = "Retail Black Friday Sale"
        }
      }
    }

    agent_protection {
      name   = "agent-security"
      action = "block"
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
      url_detected_action = "block"

      malicious_code_protection {
        name   = "malicious-code"
        action = "block"
      }
    }

    data_protection {
      data_leak_detection {
        action = "block"

        member {
          text    = "IP Addresses"
          id      = "11995029"
          version = "1"
        }
      }

      database_security {
        name   = "database-security-create"
        action = "block"
      }

      database_security {
        name   = "database-security-read"
        action = "block"
      }

      database_security {
        name   = "database-security-update"
        action = "block"
      }

      database_security {
        name   = "database-security-delete"
        action = "block"
      }
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
  profile_name = "${var.profile_prefix}Truffles - Agent Security - Moderate"

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

    app_protection {
      default_url_category = ["malicious"]
      url_detected_action  = "block"
    }

    data_protection {
      data_leak_detection {
        action           = "block"
        mask_data_inline = true

        member {
          text    = "sensitive content"
          version = "2"
        }
      }

      database_security {
        name   = "database-security-create"
        action = "allow"
      }

      database_security {
        name   = "database-security-read"
        action = "allow"
      }

      database_security {
        name   = "database-security-update"
        action = "allow"
      }

      database_security {
        name   = "database-security-delete"
        action = "block"
      }
    }
  }
}

# ── Recipe Extractor Profile ──────────────────────────────────────────
# Moderate profile for a recipe extraction agent on AWS. Allows moderate
# toxic content, uses topic guardrails to permit recipe discussions,
# and blocks prompt injection.

resource "prisma-airs_security_profile" "recipe_extractor" {
  profile_name = "${var.profile_prefix}Truffles - Recipe Extractor - Moderate"

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
      name   = "contextual-grounding"
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

    app_protection {
      default_url_category = ["malicious"]
      url_detected_action  = "block"
    }

    data_protection {
      data_leak_detection {
        action = ""
      }
    }
  }
}

# ── IDE Integration Profile (Cursor) ─────────────────────────────────
# Strict profile for code-assistant IDE integrations. Blocks all model
# threats with strict toxic content filtering. Includes malicious code
# detection via app protection.

resource "prisma-airs_security_profile" "cursor_ide" {
  profile_name = "${var.profile_prefix}InfoSec - Code Assistant - Strict"

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

    app_protection {
      default_url_category = ["malicious"]
      url_detected_action  = "block"

      malicious_code_protection {
        name   = "malicious-code"
        action = "block"
      }
    }

    data_protection {
      data_leak_detection {
        action = "block"

        member {
          text    = "sensitive content"
          version = "2"
        }
      }

      database_security {
        name   = "database-security-create"
        action = "allow"
      }

      database_security {
        name   = "database-security-read"
        action = "allow"
      }

      database_security {
        name   = "database-security-update"
        action = "block"
      }

      database_security {
        name   = "database-security-delete"
        action = "block"
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
  profile_name = "${var.profile_prefix}OpenClaw - Slack Moderation - Moderate"

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
      default_url_category = ["malicious"]
      url_detected_action  = "block"
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
  profile_name = "${var.profile_prefix}OpenClaw - HIPAA Compliance - Strict"

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

    app_protection {
      default_url_category = ["malicious"]
      url_detected_action  = "block"
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

      database_security {
        name   = "database-security-create"
        action = "block"
      }

      database_security {
        name   = "database-security-read"
        action = "allow"
      }

      database_security {
        name   = "database-security-update"
        action = "block"
      }

      database_security {
        name   = "database-security-delete"
        action = "block"
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
