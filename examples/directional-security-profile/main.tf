# Setup: Declare the provider required by this configuration.
terraform {
  required_providers {
    prisma-airs = {
      source  = "cdot65/prisma-airs"
      version = "~> 0.13.0"
    }
  }
}

# Authentication: Use Runtime Security OAuth environment variables.
provider "prisma-airs" {}

# Inputs: Use an existing DLP profile without modifying it.
variable "profile_name" {
  type    = string
  default = "terraform-directional-security"
}

variable "dlp_profile_name" {
  type        = string
  description = "Name of an existing Runtime DLP profile to inspect against."
}

variable "dlp_profile_version" {
  type        = string
  description = "Version of that existing DLP profile."
}

# Policy: Configure four independent directions and shared inspection settings.
resource "prisma-airs_runtime_security_profile" "directional" {
  profile_name = var.profile_name
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
              text    = var.dlp_profile_name
              version = var.dlp_profile_version
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
              text    = var.dlp_profile_name
              version = var.dlp_profile_version
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
              text    = var.dlp_profile_name
              version = var.dlp_profile_version
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
              text    = var.dlp_profile_name
              version = var.dlp_profile_version
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

# Receipt: Show the current revision UUID and revision number.
output "profile_id" {
  value = prisma-airs_runtime_security_profile.directional.profile_id
}

output "revision" {
  value = prisma-airs_runtime_security_profile.directional.revision
}
