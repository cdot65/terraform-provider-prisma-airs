# Policy: Combine strict blocking with DLP and URL controls.
resource "prisma-airs_runtime_security_profile" "high_security" {
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

# Policy: Limit agent topics and mask detected data leaks.
resource "prisma-airs_runtime_security_profile" "truffles_agent" {
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

# Policy: Scope recipe extraction topics and moderate toxic content.
resource "prisma-airs_runtime_security_profile" "recipe_extractor" {
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

# Policy: Apply strict content and data-leak controls to a code assistant.
resource "prisma-airs_runtime_security_profile" "cursor_ide" {
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

# Policy: Keep Slack moderation focused on injection and malicious URLs.
resource "prisma-airs_runtime_security_profile" "slack_moderation" {
  profile_name = "${var.profile_prefix}OpenClaw - Slack Moderation - Moderate"

  ai_security_profile {
    model_type           = "default"
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

# Policy: Reference an existing HIPAA DLP profile and topic controls.
resource "prisma-airs_runtime_security_profile" "hipaa_compliance" {
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
