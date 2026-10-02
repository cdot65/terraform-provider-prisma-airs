# prisma-airs_red_team_target schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages a Red Team target with exactly one native connection block. Desired payloads and secrets are preserved through masked reads; arbitrary payload drift cannot be detected reliably.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint_type` | `string` | optional, computed | — | Endpoint accessibility. |
| `connection_type` | `string` | computed | — | Provider inferred from the selected connection block. |
| `created_at` | `string` | computed | — | Creation timestamp. |
| `description` | `string` | optional, computed | — | Target description. |
| `id` | `string` | computed | — | Terraform resource ID (same as uuid). |
| `name` | `string` | required | — | Target name. |
| `network_broker_channel_uuid` | `string` | optional | — | Preexisting Network Broker channel UUID. This resource never creates channels. |
| `response_mode` | `string` | computed | — | REST or STREAMING, inferred from the selected connection block. |
| `status` | `string` | computed | — | Target status. |
| `target_type` | `string` | optional, computed | — | Target category. |
| `updated_at` | `string` | computed | — | Update timestamp. |
| `uuid` | `string` | computed | — | Target UUID. |

## basic_auth

Nesting: `single`.

### basic_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `location` | `string` | optional, computed | — | Credential location. |
| `password` | `string` | optional | yes | Basic authentication password; unavailable on import. |
| `username` | `string` | optional | yes | Basic authentication username. |

## bedrock

Nesting: `single`.

### bedrock

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `access_id` | `string` | optional | yes | AWS Bedrock access_id. |
| `access_secret` | `string` | optional | yes | AWS Bedrock access_secret. |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `model_id` | `string` | optional | — | AWS Bedrock model_id. |
| `region` | `string` | optional | — | AWS Bedrock region. |
| `request_body` | `dynamic` | optional | — | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | — | Nonsecret provider request headers. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |
| `session_token` | `string` | optional | yes | Optional AWS session token. |

## custom

Nesting: `single`.

### custom

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | — | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | — | Nonsecret request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |

## databricks

Nesting: `single`.

### databricks

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `access_token` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `client_id` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `model_name` | `string` | optional | — | Serving model name. |
| `request_body` | `dynamic` | optional | — | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | — | Nonsecret provider request headers. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |
| `response_stop_key` | `string` | optional | — | Databricks streaming completion field. |
| `response_stop_value` | `string` | optional | — | Databricks streaming completion value. |
| `secret` | `string` | optional | yes | Use access_token, or client_id and secret for OAuth; never combine them. |
| `workspace_url` | `string` | optional | — | Databricks workspace URL. |

## headers_auth

Nesting: `single`.

### headers_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `headers` | `map(string)` | optional | yes | Desired secret header values; unavailable on import. |

## hugging_face

Nesting: `single`.

### hugging_face

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `api_key` | `string` | optional | yes | Provider credential. Unavailable on import; supply the desired value. |
| `model_name` | `string` | optional | — | Provider model name. |
| `request_body` | `dynamic` | optional | — | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | — | Nonsecret provider request headers. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |

## oauth2_auth

Nesting: `single`.

### oauth2_auth

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `body` | `dynamic` | optional | yes | Desired native HCL token request body. |
| `expiry_minutes` | `number` | optional, computed | — | Token validity in minutes; defaults to 60. |
| `headers` | `dynamic` | optional | yes | Desired native HCL token request headers. |
| `inject_header` | `map(string)` | optional | yes | Header templates containing {TOKEN}. |
| `response_key` | `string` | optional, computed | — | Token response path; defaults to access_token. |
| `token_url` | `string` | optional | — | Token endpoint URL. |

## openai

Nesting: `single`.

### openai

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Optional provider endpoint override. |
| `api_key` | `string` | optional | yes | Provider credential. Unavailable on import; supply the desired value. |
| `model_name` | `string` | optional | — | Provider model name. |
| `request_body` | `dynamic` | optional | — | Native HCL request object containing {INPUT}; required for text targets. |
| `request_headers` | `map(string)` | optional | — | Nonsecret provider request headers. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. |
| `response_key` | `string` | optional | — | Desired response path. Native provider read-back may omit this value; unavailable on import. |

## rest

Nesting: `single`.

### rest

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | — | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | — | Nonsecret request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |

## streaming

Nesting: `single`.

### streaming

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `api_endpoint` | `string` | optional | — | Target API URL; management does not execute inference. |
| `request_body` | `dynamic` | optional | — | Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal. |
| `request_headers` | `map(string)` | optional | — | Nonsecret request headers. Put credentials in an authentication block. |
| `response_body` | `dynamic` | optional | — | Desired native HCL response object. Read-back cannot reliably detect changes inside this payload. |
| `response_key` | `string` | optional | — | Path to the model response. |
| `response_stop_key` | `string` | optional | — | Streaming completion field. |
| `response_stop_value` | `string` | optional | — | Streaming completion value. |
