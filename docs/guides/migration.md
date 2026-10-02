---
page_title: "Migrate to the updated provider"
---

# Migrate to the updated provider

Provider v0.7.0 uses Go SDK v0.6.1 and changes the schemas from v0.6.3. Follow [installation](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/installation/), select `~> 0.7.0`, remove any development override, and run `terraform init -upgrade`. Review and refactor existing HCL and state before applying.

## Review existing ownership

Back up state in your protected backend and inventory the resources already managed at each Terraform address. Plan in a disposable environment before applying to a shared tenant. Existing HCL and state may require refactoring or explicit re-import.

| Area | Required action |
| --- | --- |
| Red Team target | Replace `connection_params` JSON input with exactly one native connection block and, where needed, an authentication block |
| Prompt set | Remove the unsupported `properties` input |
| Security profile | Treat the name as the managed history; changes generate UUIDs and revisions |
| DLP profile reference | Supply nonempty `log_severity` for every `dlp_data_profile` block |
| Deployment profiles | Treat `profile_id`, `auth_code`, and `details` as sensitive values |
| Customer app | Import existing apps; rename and create are unsupported |

## Write target inputs as native HCL

```hcl
variable "app_token" {
  type      = string
  sensitive = true
}

resource "prisma-airs_red_team_target" "app" {
  name = "production-app"

  rest {
    api_endpoint    = "https://app.example.com/chat"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { prompt = "{INPUT}", stream = false, temperature = 0 }
    response_body   = { answer = "{RESPONSE}" }
    response_key    = "answer"
  }

  headers_auth {
    headers = { Authorization = "Bearer ${var.app_token}" }
  }
}
```

Preserve the Terraform address where appropriate. Native blocks preserve false, zero, null, empty values, and nested objects/lists. For an existing CUSTOM/REST target, `rest/<uuid>` is the [import hint](https://cdot65.github.io/terraform-provider-prisma-airs/resources/red-team-target/#import-and-outputs); plain UUID import selects `custom`. Imported credentials and payloads that the API cannot recover remain null until configured.

## Understand replacement and update plans

Any setting change inside `openai`, `hugging_face`, `databricks`, or `bedrock` plans replacement. Custom transport mode/authentication method changes update in place; removing configured authentication or connection fields plans replacement. Clearing a Network Broker channel also plans target replacement. The channel remains externally managed.

Security profile policy edits update the Terraform resource while AIRS creates a new revision UUID. A rename creates version 1 under the new name and leaves the old history. Destroy owns all revisions of the currently managed name.

All API-key create inputs plan replacement. Imported keys have no recoverable key secret. Key deletion can also delete its associated customer app. Clearing a prompt-set description plans replacement and archives the old set.

## Verify the new state

```bash
terraform validate
terraform plan -out=migration.tfplan
terraform apply migration.tfplan
terraform plan -detailed-exitcode
```

Inspect every replacement and profile-name change before apply. A stable follow-up plan should exit 0. Protect the saved plan and state, including sensitive deployment-profile aliases. If outputting deployment details or API-key values, declare `sensitive = true`.
