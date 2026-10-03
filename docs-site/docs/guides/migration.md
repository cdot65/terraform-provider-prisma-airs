# Migrate to the updated provider

Provider v0.8.0 groups existing functionality by product and uses Go SDK v0.6.1. The provider address remains `cdot65/prisma-airs`. Upgrade HCL and state together before applying; the provider does not register aliases for old type names.

## Migrate v0.7.0 names and settings

| Old type | New type |
| --- | --- |
| `prisma-airs_security_profile` | `prisma-airs_runtime_security_profile` |
| `prisma-airs_custom_topic` | `prisma-airs_runtime_custom_topic` |
| `prisma-airs_api_key` | `prisma-airs_runtime_api_key` |
| `prisma-airs_customer_app` | `prisma-airs_runtime_customer_app` |
| `prisma-airs_model_security_group` | `prisma-airs_supply_chain_security_group` |
| `prisma-airs_dlp_profiles` | `prisma-airs_runtime_dlp_profiles` |
| `prisma-airs_deployment_profiles` | `prisma-airs_runtime_deployment_profiles` |
| `prisma-airs_model_security_rules` | `prisma-airs_supply_chain_security_rules` |

Red Teaming type names remain unchanged. Update references and outputs as well as declarations. Move `mgmt_endpoint` into `runtime`; move `red_team_data_endpoint` / `red_team_mgmt_endpoint` into `red_team` as `data_endpoint` / `mgmt_endpoint`; move `model_sec_data_endpoint` / `model_sec_mgmt_endpoint` into `supply_chain` with the same nested attribute names. Endpoint environment variables remain unchanged; shared OAuth attributes stay at the top level.

Pause concurrent Terraform runs and back up state in protected storage. With v0.7.0 still installed, record each resource's import identifier and configured secrets before changing HCL. Follow the individual [lifecycle guides](../reference/index.md) for supported IDs; security profiles import by `profile_name`, customer apps by `app_name`, and keys/topics/groups by their documented identifiers. Record the profile name, not its current UUID. Red Teaming types are unchanged and need no state removal or re-import for this product refactor.

Before switching provider versions, use the installed v0.7.0 provider to remove **all** renamed managed resource and data-source addresses from state in one step. Use `terraform state list` to inventory the full addresses, including module prefixes and `for_each`/`count` indices. Shell-quote indexed addresses, for example `'module.m.prisma-airs_custom_topic.t["a"]'`. Keep Red Teaming addresses in state. Do not plan, apply, or import while any renamed old-type address remains: the new provider has no schema for it.

After updating HCL and selecting `~> 0.8.0`, run `terraform init -upgrade` and import each recorded remote object into its new address. For a profile, topic, and DLP catalog, the sequence is:

```bash
terraform state pull > protected-backup.tfstate
# With v0.7.0 still installed; include EVERY renamed address from your inventory.
terraform state rm 'prisma-airs_security_profile.first' 'prisma-airs_custom_topic.topic' 'data.prisma-airs_dlp_profiles.catalog'
# Update declarations/references and select ~> 0.8.0, then initialize.
terraform init -upgrade
terraform import 'prisma-airs_runtime_security_profile.first' '<recorded-profile-name>'
terraform import 'prisma-airs_runtime_custom_topic.topic' '<recorded-topic-id>'
# Imports and the final plan re-read the renamed data sources.
terraform plan
```

`state rm` forgets ownership without deleting the remote object. Do not apply between removal and completion of all imports. Leaving other old-type entries in state can make import/plan fail while decoding unsupported schemas. This provider does not implement cross-type state moves; do not rely on `moved` blocks or `state mv` to convert types. Re-import cannot recover the one-time API-key secret. The `api_key` output is computed-only, so it cannot be restored by setting a configuration argument. After removal/import, references to that attribute will plan to null, including outputs or secret-store resources. Preserve an existing key value in your protected external secret store before migration and update downstream consumers to use it, rather than the imported resource's null secret. If the value is unavailable, deliberate key replacement/rotation is needed to obtain a new secret; assess application impact first. Configured key creation inputs remain in HCL. Red Teaming state and its configured credentials need no re-import for this refactor. Importing profiles adopts all revisions under the resolved name; no rename is needed for this migration. Rename data-source declarations and remove their old state entries with the managed resources; they are read again during import/planning.

If a migration fails, stop without applying and recover using your backend's protected state version. Inspect the final plan: renamed resources should remain imported objects, not replacements. Review any actual remote drift separately.

## Upgrading v0.6.3 or earlier

The v0.7.0 native HCL and lifecycle changes below also apply when coming from an older release.

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
# Inputs: Supply environment-specific values; load sensitive values from your secret store.
variable "app_token" {
  type      = string
  sensitive = true
}

# Target: Register the endpoint contract; apply does not run an assessment.
resource "prisma-airs_red_team_target" "app" {
  name = "production-app"

  rest {
    api_endpoint = "https://app.example.com/chat"

    request_headers = {
      "Content-Type" = "application/json"
    }

    request_body = {
      prompt      = "{INPUT}"
      stream      = false
      temperature = 0
    }

    response_body = {
      answer = "{RESPONSE}"
    }

    response_key = "answer"
  }

  headers_auth {
    headers = {
      Authorization = "Bearer ${var.app_token}"
    }
  }
}
```

Preserve the Terraform address where appropriate. Native blocks preserve false, zero, null, empty values, and nested objects/lists. For an existing CUSTOM/REST target, `rest/<uuid>` is the [import hint](../resources/red-team-target.md#import-and-outputs); plain UUID import selects `custom`. Imported credentials and payloads that the API cannot recover remain null until configured.

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
