# prisma-airs_red_team_target

Manages target configuration through the Red Team management API. Create and update do not execute inference or connection validation.

## Example usage

```hcl
# Target: Register the endpoint contract; apply does not run an assessment.
resource "prisma-airs_red_team_target" "chatbot" {
  name = "production-chatbot"

  rest {
    api_endpoint = "https://chatbot.example.com/chat"

    request_headers = {
      "Content-Type" = "application/json"
    }

    request_body = {
      prompt      = "{INPUT}"
      temperature = 0
      stream      = false
    }

    response_body = {
      output = "{RESPONSE}"
    }

    response_key = "output"
  }

  headers_auth {
    headers = {
      Authorization = "Bearer ${var.app_token}"
    }
  }
}
```

## Connection blocks

Configure exactly one block. Payloads are native HCL objects or maps, with nested lists and explicit null, empty, false, and zero values preserved. JSON serialization happens inside the provider.

| Block | Provider settings | Inferred response mode |
| --- | --- | --- |
| `openai` | `api_key`, `model_name` | REST |
| `hugging_face` | `api_key`, `model_name` | REST |
| `bedrock` | `access_id`, `access_secret`, `region`, `model_id`; optional `session_token` | REST |
| `databricks` | `workspace_url`, `model_name`; `access_token` **or** `client_id` and `secret` | STREAMING |
| `custom`, `rest` | `api_endpoint` | REST |
| `streaming` | `api_endpoint` | STREAMING |
| `adapter` | `uuid`; optional `variable_overrides` | null (adapter-controlled) |

All endpoint blocks accept `request_body`, `response_body`, `response_key`, and optional nonsecret `request_headers`. Native provider blocks accept optional `api_endpoint` overrides. All endpoint blocks require request/response payloads for writes; every block except OpenAI requires `response_key`. Streaming and Databricks require `response_stop_key` and `response_stop_value`.

`target_type` defaults to `MODEL` for native providers and `APPLICATION` for custom transports. Native providers require `MODEL`. `api_endpoint_type` defaults to `PUBLIC`; `NETWORK_BROKER` requires a preexisting `network_broker_channel_uuid`. This resource never creates or removes channels.

`rest` and `streaming` use the service's `CUSTOM` discriminator with the corresponding response mode. `connection_type` and `response_mode` are computed outputs. Legacy JSON `connection_params` input is removed.

Adapter targets require `api_endpoint_type = "NETWORK_BROKER"` and an existing channel. The adapter UUID may reference a managed [adapter resource](red-team-adapter.md) or be resolved through [discovery](../data-sources/red-team-adapters.md). Adapter overrides use a key map of `{ type = "VAR" or "SECRET", value = ... }`; missing imported override secrets permit no-op adoption but block writes until supplied.

## Authentication and state

Custom transports accept at most one of `headers_auth` (`headers`), `basic_auth` (`username`, `password`, optional `location`), or `oauth2_auth` (`token_url`, `inject_header`, optional `headers`, `body`, `expiry_minutes`, `response_key`). OAuth defaults are `expiry_minutes = 60` and `response_key = "access_token"`; expiry zero is allowed. Native blocks contain their own credentials and reject separate authentication blocks. Removing authentication plans target replacement because the service ignores null authentication in updates; changing between supported methods updates the existing target. Changing any setting within a native provider block (OpenAI, Hugging Face, Databricks or Bedrock), switching native provider families, changing target category, removing configured custom connection/auth fields, or clearing a Network Broker channel plans replacement. Native credential updates cannot be independently confirmed through masked reads; replacement applies the desired credentials through creation. Top-level name/description edits still update in place. Custom transport mode changes update the existing target. Switching between adapter and endpoint families plans replacement so the API cannot retain an old adapter reference.

Credentials, authentication headers, and authentication bodies are sensitive. Request/response templates and request headers are also sensitive because imported payloads can contain credentials. Sensitive values are still stored in state, so protect the backend.

Desired payloads, request headers, and credentials retain their configured values across refresh. Masked responses never become desired credentials. Readable fixed fields such as endpoint and model names detect drift. Drift inside arbitrary payloads cannot reliably be detected. Native provider `response_key` is required by some writes but omitted by read-back, and remains a desired setting.

## Import and outputs

```bash
terraform import prisma-airs_red_team_target.chatbot rest/<uuid>
```

Plain UUID import chooses `custom` for a CUSTOM/REST target. The optional `<family>/<uuid>` hint distinguishes `rest` from `custom` and must match the observed family. Imports recover usable request/response payloads, request headers, fixed fields, and OAuth injection templates returned by the service. Missing or masked inputs stay null. Match observable settings to get a read-only no-op plan even without original credentials. Create/update requires complete inputs; `unavailable_fields` identifies explicitly redacted payload/header paths that must be supplied before updating. Never copy a mask into configuration. Omitted native response keys remain null until supplied.

Outputs include `id`, `uuid`, `connection_type`, `response_mode`, `status`, `created_at`, and `updated_at`. Name and description are configurable; an omitted description becomes an empty string.

## Complete schema

See the [exact schema reference](../reference/generated/prisma-airs_red_team_target.md) for all nested fields, types, and sensitivity flags.
