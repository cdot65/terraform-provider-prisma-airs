# Directional security profiles

Configure prompt, response, tool-call, and tool-response inspection independently. Use provider v0.13.0 or later, built on Go SDK v0.9.0. Existing legacy protection blocks remain usable.

## Complete configuration

Save this configuration as `main.tf` in a separate directory. Load the [Runtime Security OAuth credentials](../getting-started/authentication.md), and select an unused profile name. Set `TF_VAR_dlp_profile_name` and `TF_VAR_dlp_profile_version` to an existing DLP profile's name and version; the example references that profile without changing it.

```hcl
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
```

Shared latency, storage masking, and full-conversation inspection belong directly in `ai_security_profile`. Direction blocks contain only protection settings. `tool_call` and `tool_response` map to the API's `tool-call` and `tool-response` keys. Mode and severity values are open-ended strings; the examples do not define exhaustive enums.

This example was applied to the live Runtime API on 2026-10-09 using a disposable profile and a real, read-only DLP reference. Create, refresh, import, response-only edits, shared edits, detector removal, remote drift detection, repair, and destroy passed. See [captured commands, outputs, and actual revision UUIDs](../development/directional-profile-verification.md).

The original sanitized response fixture also captured prompt database security as null, response without prompt injection, and tool-response without inline masking. The provider preserves these distinctions in its policy snapshot. The original create request was truncated; null/empty and extension edge cases remain supported by fixture and mock evidence.

## Read, update, and import

Change only the response toxicity confidence overrides, review `terraform plan`, and apply. Unrelated prompt and tool protections retain their settings and extension fields. Policy updates create a new service revision and UUID under the same Terraform resource address.

```bash
terraform import prisma-airs_runtime_security_profile.directional "terraform-directional-security"
terraform plan
```

Align all managed protection blocks with the imported policy. A profile may contain both legacy and directional protections; refresh preserves both, while configuration that writes both layouts in one profile entry receives a diagnostic. Resolve ownership before applying a layout change.

The computed `dlp_tenant_id` records optional response metadata. GET may omit it; this does not prevent profile management. A tenant seen in a write response remains recorded through subsequent omitted GET responses.

## Preservation and ownership

Terraform private state retains the service policy, including omitted fields, explicit false, null arrays, empty arrays/objects, unknown direction keys, detector options, DLP rule objects, and future additive fields. Updates compare the previous and planned typed projections, applying managed changes to that saved JSON. Policy numbers retain their precision. The snapshot is provider metadata, not an editable HCL attribute. Unknown top-level profile extensions are retained separately and carried into update requests; known response audit fields are excluded.

Entries are matched within their profile and direction by identity, never by list position. Duplicate identities that make a changed list ambiguous produce a diagnostic before writing. Removing an entire direction container with future direction keys, or an unrepresentable detector without a name/action, also requires resolving ownership before mutation. Observed toxic-content rules with explicit category settings retain empty actions. Removing a managed detector removes its subtree, including extensions on that detector. Unchanged opaque fields survive other edits. Optional computed severities retain observed service values when omitted from configuration; configuring them explicitly makes remote changes observable as drift.

The schema adds fields without changing existing paths or types. Old state remains readable and acquires preservation metadata on refresh. If that metadata is unavailable, update reports a diagnostic requiring refresh.

Destroy removes every revision under the managed name, including revisions predating import. See the [security profile lifecycle](../resources/security-profile.md#revision-ownership-and-changes).
