---
page_title: "Red Team Testing"
---

# Red Team Testing

Register an application or model endpoint and a custom prompt-set record through the AIRS Red Team management API. Terraform manages configuration; it does not start assessment jobs or execute inference.

## Before you start

Use the [updated provider from source](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/installation/#use-the-updated-provider-from-source), a Red Team licensed tenant, and management OAuth permissions. Follow [Authentication](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/authentication/). Prepare an endpoint you own and supply its credentials through sensitive variables.

Choose exactly one connection family: `openai`, `hugging_face`, `databricks`, `bedrock`, `custom`, `rest`, or `streaming`. The [complete native-target example](https://cdot65.github.io/terraform-provider-prisma-airs/examples/native-targets/) illustrates all seven. Native providers contain their credentials; custom transports may use `headers_auth`, `basic_auth`, or `oauth2_auth`.

## Step 1: Create a Target

```hcl
variable "chatbot_api_key" {
  type      = string
  sensitive = true
}

resource "prisma-airs_red_team_target" "chatbot" {
  name            = "customer-chatbot"
  target_type     = "APPLICATION"
  description     = "Customer-facing chatbot application"
  rest {
    api_endpoint = "https://chatbot.example.com/api/chat"
    request_headers = {
      "Content-Type"  = "application/json"
    }
    request_body = {
      prompt = "{INPUT}"
    }
    response_body = {
      output = "{RESPONSE}"
    }
    response_key = "output"
  }
  headers_auth {
    headers = { Authorization = "Bearer ${var.chatbot_api_key}" }
  }
}
```

## Step 2: Create Custom Prompt Sets

Create custom prompt sets for targeted testing:

```hcl
resource "prisma-airs_red_team_custom_prompt_set" "injection_tests" {
  name        = "custom-injection-tests"
  description = "Custom prompt injection test cases for our domain"
}
```

## Step 3: Review and apply

```bash
terraform validate
terraform plan -out=targets.tfplan
terraform apply targets.tfplan
terraform plan -detailed-exitcode
```

Review the saved plan before applying. The follow-up plan should exit 0. Create/update use management operations without connection validation; successful registration alone does not prove endpoint inference or a scan verdict.

## Step 4: Update or adopt a target

Any setting change inside a native model-provider block plans replacement. Custom transport mode and supported authentication-method switches can update in place. Removing authentication or configured connection fields, changing category, or clearing a broker channel plans replacement. Ordinary payload diffs remain visible; credentials remain sensitive. See [target authentication and state](https://cdot65.github.io/terraform-provider-prisma-airs/resources/red-team-target/#authentication-and-state).

Import a CUSTOM/REST target with a `rest/<uuid>` hint to retain the REST family; plain UUID import chooses `custom`. API-masked credentials and unrecoverable payloads remain null on import. Configure desired values before connection edits.

## Step 5: Use an existing channel or clean up

A `NETWORK_BROKER` target requires a preexisting `network_broker_channel_uuid`. Target operations never create or remove the channel. Setting it back to a public connection may require target replacement; inspect the plan.

Destroy removes target configuration and archives prompt sets. Archived sets are removed from Terraform state and cannot be imported as active resources. Clearing a nonempty prompt-set description also replaces the set and archives its predecessor. Protect state and saved plans because sensitive markings do not remove stored credentials.
