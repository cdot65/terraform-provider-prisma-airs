resource "prisma-airs_runtime_security_profile" "directional" {
  profile_name = "Directionality Test"
  ai_security_profile {
    model_type                          = "default"
    content_type_mode                   = "per_content_type"
    mask_data_in_storage                = false
    enable_full_conversation_inspection = false
    latency {
      inline_timeout_action = "block"
      max_inline_latency    = 5
    }
    content_type_configurations {
      prompt {
        app_protection {
          default_url_category = ["malicious"]
          malicious_code_protection {
            action   = "block"
            name     = "malicious-code"
            severity = "high"
          }
          url_detected_action   = "block"
          url_detected_severity = "low"
        }
        data_protection {
          data_leak_detection {
            action           = "block"
            mask_data_inline = true
            member {
              id      = ""
              text    = "sensitive content"
              version = "2"
            }
          }
        }
        model_protection {
          action   = "block"
          name     = "prompt-injection"
          severity = "medium"
        }
        model_protection {
          action = "high:block, moderate:allow"
          name   = "toxic-content"
          severity_by_confidence {
            high     = "medium"
            moderate = "low"
          }
        }
      }
      response {
        app_protection {
          default_url_category = ["malicious"]
          malicious_code_protection {
            action   = "block"
            name     = "malicious-code"
            severity = "high"
          }
          url_detected_action   = "block"
          url_detected_severity = "low"
        }
        data_protection {
          data_leak_detection {
            action           = "block"
            mask_data_inline = true
            member {
              id      = ""
              text    = "sensitive content"
              version = "2"
            }
          }
          database_security {
            name     = "database-security-create"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-read"
            action   = "block"
            severity = "low"
          }
          database_security {
            name     = "database-security-update"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-delete"
            action   = "block"
            severity = "high"
          }
        }
        model_protection {
          action = "high:block, moderate:allow"
          name   = "toxic-content"
          severity_by_confidence {
            high     = "medium"
            moderate = "low"
          }
        }
      }
      tool_call {
        app_protection {
          default_url_category = ["malicious"]
          malicious_code_protection {
            action   = "block"
            name     = "malicious-code"
            severity = "high"
          }
          url_detected_action   = "block"
          url_detected_severity = "low"
        }
        data_protection {
          data_leak_detection {
            action           = "block"
            mask_data_inline = true
            member {
              id      = ""
              text    = "sensitive content"
              version = "2"
            }
          }
          database_security {
            name     = "database-security-create"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-read"
            action   = "block"
            severity = "low"
          }
          database_security {
            name     = "database-security-update"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-delete"
            action   = "block"
            severity = "high"
          }
        }
        model_protection {
          action   = "block"
          name     = "prompt-injection"
          severity = "medium"
        }
        model_protection {
          action = "high:block, moderate:allow"
          name   = "toxic-content"
          severity_by_confidence {
            high     = "medium"
            moderate = "low"
          }
        }
      }
      tool_response {
        app_protection {
          default_url_category = ["malicious"]
          malicious_code_protection {
            action   = "block"
            name     = "malicious-code"
            severity = "high"
          }
          url_detected_action   = "block"
          url_detected_severity = "low"
        }
        data_protection {
          data_leak_detection {
            action = "block"
            member {
              id      = ""
              text    = "sensitive content"
              version = "2"
            }
          }
          database_security {
            name     = "database-security-create"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-read"
            action   = "block"
            severity = "low"
          }
          database_security {
            name     = "database-security-update"
            action   = "block"
            severity = "medium"
          }
          database_security {
            name     = "database-security-delete"
            action   = "block"
            severity = "high"
          }
        }
        model_protection {
          action   = "block"
          name     = "prompt-injection"
          severity = "medium"
        }
        model_protection {
          action = "high:block, moderate:allow"
          name   = "toxic-content"
          severity_by_confidence {
            high     = "medium"
            moderate = "low"
          }
        }
      }
    }
  }
}
