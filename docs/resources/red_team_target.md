---
page_title: "prisma-airs_red_team_target (Resource)"
subcategory: "AI Red Teaming"
---

# prisma-airs_red_team_target Resource

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

Adapter targets require `api_endpoint_type = "NETWORK_BROKER"` and an existing channel. The adapter UUID may reference a managed [adapter resource](https://cdot65.github.io/terraform-provider-prisma-airs/resources/red-team-adapter/) or be resolved through [discovery](https://cdot65.github.io/terraform-provider-prisma-airs/data-sources/red-team-adapters/). Adapter overrides use a key map of `{ type = "VAR" or "SECRET", value = ... }`; missing imported override secrets permit no-op adoption but block writes until supplied.

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

See the [exact schema reference](https://cdot65.github.io/terraform-provider-prisma-airs/reference/generated/prisma-airs_red_team_target/) for all nested fields, types, and sensitivity flags.

## Schema

### Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint_type` | `string` | optional, computed | — | Endpoint accessibility. |
| `connection_type` | `string` | computed | — | Provider inferred from the selected connection block. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Target description. |
| `id` | `string` | computed | — | Terraform resource ID (same as uuid). |
| `name` | `string` | required | — | Target name. |
| `network_broker_channel_uuid` | `string` | optional | — | Preexisting Network Broker channel UUID. This resource never creates channels. |
| `response_mode` | `string` | computed | — | REST or STREAMING for endpoint connections; null for adapter-controlled transport. |
| `status` | `string` | computed | — | Target status. |
| `target_type` | `string` | optional, computed | — | Target category. |
| `unavailable_fields` | `list(string)` | computed | — | Payload/header paths explicitly redacted by the API. Writes require original values at these paths; no-op adoption does not. |
| `updated_at` | `string` | computed | — | Update timestamp. |
| `uuid` | `string` | computed | — | Target UUID. |

### adapter

Nesting: `single`.

#### adapter

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `uuid` | `string` | optional | — | Existing or Terraform-managed adapter UUID. |
| `variable_overrides` | `map(object)` | optional | yes | Complete target override key set. New overrides require values; redacted imported overrides cannot be written until supplied. |

##### adapter.variable_overrides

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `type` | `string` | required | — | VAR or SECRET. |
| `value` | `string` | optional | yes | Override value. Unavailable imported SECRET values remain null for read-only adoption. |

### basic_auth

Nesting: `single`.

#### basic_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `location` | `string` | optional, computed | — | Credential location. |
| `password` | `string` | optional | yes | Basic authentication password; unavailable on import. |
| `username` | `string` | optional | yes | Basic authentication username. |

### bedrock

Nesting: `single`.

#### bedrock

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `access_id` | `string` | optional | yes | AWS Bedrock access_id. |
| `access_secret` | `string` | optional | yes | AWS Bedrock access_secret. |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `model_id` | `string` | optional | — | AWS Bedrock model_id. |
| `region` | `string` | optional | — | AWS Bedrock region. |
| `request_body` | `dynamic` | optional | yes | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | yes | Desired provider request headers. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |
| `session_token` | `string` | optional | yes | Optional AWS session token. |

### custom

Nesting: `single`.

#### custom

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | yes | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | yes | Desired request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |

### databricks

Nesting: `single`.

#### databricks

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `access_token` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `client_id` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `model_name` | `string` | optional | — | Serving model name. |
| `request_body` | `dynamic` | optional | yes | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | yes | Desired provider request headers. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |
| `response_stop_key` | `string` | optional | — | Databricks streaming completion field. Imported empty values permit read-only adoption; writes require a nonempty value. |
| `response_stop_value` | `string` | optional | — | Databricks streaming completion value. Imported empty values permit read-only adoption; writes require a nonempty value. |
| `secret` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `workspace_url` | `string` | optional | — | Databricks workspace URL. |

### headers_auth

Nesting: `single`.

#### headers_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `headers` | `map(string)` | optional | yes | Desired secret header values; unavailable on import. |

### hugging_face

Nesting: `single`.

#### hugging_face

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `api_key` | `string` | optional | yes | Provider credential. Unavailable on import; supply the desired value. |
| `model_name` | `string` | optional | — | Provider model name. |
| `request_body` | `dynamic` | optional | yes | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | yes | Desired provider request headers. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |

### oauth2_auth

Nesting: `single`.

#### oauth2_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `body` | `dynamic` | optional | yes | Desired native HCL token request body. |
| `expiry_minutes` | `number` | optional, computed | — | Token validity in minutes; defaults to 60. |
| `headers` | `dynamic` | optional | yes | Desired native HCL token request headers. |
| `inject_header` | `map(string)` | optional | yes | Header templates containing {TOKEN}. |
| `response_key` | `string` | optional, computed | — | Token response path; defaults to access_token. |
| `token_url` | `string` | optional | — | Token endpoint URL. |

### openai

Nesting: `single`.

#### openai

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `api_key` | `string` | optional | yes | Provider credential. Unavailable on import; supply the desired value. |
| `model_name` | `string` | optional | — | Provider model name. |
| `request_body` | `dynamic` | optional | yes | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | yes | Desired provider request headers. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |

### rest

Nesting: `single`.

#### rest

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | yes | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | yes | Desired request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |

### streaming

Nesting: `single`.

#### streaming

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | yes | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | yes | Desired request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | yes | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |
| `response_stop_key` | `string` | optional | — | Streaming completion field. Imported empty values permit read-only adoption; writes require a nonempty value. |
| `response_stop_value` | `string` | optional | — | Streaming completion value. Imported empty values permit read-only adoption; writes require a nonempty value. |
