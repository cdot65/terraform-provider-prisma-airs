# prisma-airs_customer_app schema

Exact resource attributes for the updated provider. See [Provider reference](../index.md) for lifecycle guides and imports.

Manages an existing customer application in Prisma AIRS. Customer apps are created externally (via the AIRS console or when apps register); use `terraform import` to bring them under management.

Presence flags describe the schema. Lifecycle guides specify validation rules, defaults, update behavior, and replacement conditions.

## Attributes

| Attribute | Type | Presence | Sensitive | Description |
| --- | --- | --- | --- | --- |
| `agent_app` | `bool` | computed | — | Whether this is an agent application. |
| `ai_agent_framework` | `string` | computed | — | AI agent framework. |
| `ai_sec_profile_name` | `string` | computed | — | Associated AI security profile name. |
| `app_name` | `string` | required | — | Name of the customer application. Explicit empty values are unsupported; omit optional fields to retain observed metadata. |
| `cloud_provider` | `string` | optional, computed | — | Cloud provider for the app. Explicit empty values are unsupported; omit optional fields to retain observed metadata. |
| `created_by` | `string` | computed | — | Identity of the user who created the app. |
| `customer_app_id` | `string` | computed | — | The unique identifier of the customer app. |
| `environment` | `string` | optional, computed | — | Deployment environment. Explicit empty values are unsupported; omit optional fields to retain observed metadata. |
| `id` | `string` | computed | — | Terraform resource ID (same as customer_app_id). |
| `model_name` | `string` | optional, computed | — | Model name associated with the app. Explicit empty values are unsupported; omit optional fields to retain observed metadata. |
| `status` | `string` | computed | — | App status. |
| `tsg_id` | `string` | computed | — | Tenant service group ID. |
| `updated_by` | `string` | optional | — | Identity of the user updating the app. |
