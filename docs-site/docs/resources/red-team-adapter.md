# prisma-airs_red_team_adapter

Own a Python adapter script and its complete variable inventory. An adapter translates Red Team prompts into calls to your application. See the [complete adapter walkthrough](../examples/red-team-adapters.md) for a runnable project.

## Save a draft first

```hcl
# Adapter: Save configuration without running the script.
resource "prisma-airs_red_team_adapter" "application" {
  name   = "terraform-application"
  script = var.adapter_script

  variables = {
    endpoint = {
      type  = "VAR"
      value = var.application_endpoint
    }
    api_key = {
      type  = "SECRET"
      value = var.application_api_key
    }
  }
}
```

`script` is plaintext; the provider encodes it for the API. Scripts and the entire variable map are sensitive because they can contain credentials. Protect state and saved plans.

`validate` defaults to `false`, explicitly overriding the API's execution default. Draft create/update does not run the script. To activate, supply an existing online `network_broker_channel_uuid`, set `validate = true`, and review the plan. Every subsequent create/update with that setting executes the script using `validation_prompt`; refresh, discovery, import, and no-op plans do not execute it. A channel reference does not install or upgrade a broker.

## Variables and updates

`variables` is the complete desired key set. An omitted key is deleted on update. Each entry has `type = "VAR"` or `"SECRET"` and a string `value`.

Imported secrets have `value = null`. Keep the key and its `SECRET` type: null retains its existing server value during updates. New keys, new secrets, and type changes need original values. Redacted masks are rejected. Plain variables and key/type changes are observable drift; secret-value drift cannot be established from redacted reads.

## Import an existing adapter

```bash
terraform import prisma-airs_red_team_adapter.application '<adapter-uuid>'
```

Recover the script, description, channel, variable names/types, visible values, and status from state. Keep each secret key with a null value if its original value is unavailable. Import records `validate = true` for ACTIVE adapters and false for drafts; match that setting to obtain a no-op plan. Future writes to an ACTIVE adapter with true execute the imported script, so review it first. The transient validation prompt cannot be recovered and uses the provider default.

Destroy dependent adapter targets before the adapter. Terraform infers this ordering when their UUID references this resource's `id`. An externally managed target can prevent deletion.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_red_team_adapter.md).
