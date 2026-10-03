# Provider Configuration

OAuth credentials are shared defaults. Endpoint overrides belong to the product that uses them. Known explicit values take precedence over the mapped environment variable, then the SDK default. An unknown value at Configure time falls back to the environment, matching prior releases; use known endpoint values when planning against a specific service.

## Shared credentials

| Attribute | Environment variable |
| --- | --- |
| `client_id` | `PANW_MGMT_CLIENT_ID` |
| `client_secret` (sensitive) | `PANW_MGMT_CLIENT_SECRET` |
| `tsg_id` | `PANW_MGMT_TSG_ID` |
| `token_endpoint` | `PANW_MGMT_TOKEN_ENDPOINT` |

The default token endpoint is `https://auth.apps.paloaltonetworks.com/oauth2/access_token`. Configure secrets through a secret manager; see [authentication](../getting-started/authentication.md).

## Product settings

Each block is optional. Omitting it still allows endpoint environment overrides.

| Product | Block attribute | Environment variable | SDK default |
| --- | --- | --- | --- |
| AI Runtime Security | `runtime.mgmt_endpoint` | `PANW_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aisec` |
| AI Red Teaming | `red_team.data_endpoint` | `PANW_RED_TEAM_DATA_ENDPOINT` | `https://api.sase.paloaltonetworks.com/ai-red-teaming/data-plane` |
| AI Red Teaming | `red_team.mgmt_endpoint` | `PANW_RED_TEAM_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/ai-red-teaming/mgmt-plane` |
| AI Gateway | `gateway.data_endpoint` | `PANW_AI_GW_DATA_ENDPOINT` | `https://api.apps.paloaltonetworks.com/ai_gw/v2` |
| AI Gateway | `gateway.admin_endpoint` | `PANW_AI_GW_ADMIN_ENDPOINT` | `https://api.apps.paloaltonetworks.com/ai_gw/admin/v2` |
| AI Supply Chain Security | `supply_chain.data_endpoint` | `PANW_MODEL_SEC_DATA_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aims/data` |
| AI Supply Chain Security | `supply_chain.mgmt_endpoint` | `PANW_MODEL_SEC_MGMT_ENDPOINT` | `https://api.sase.paloaltonetworks.com/aims/mgmt` |

The Supply Chain Security module currently manages Model Security groups and rules; it does not claim coverage of every product capability.

## Example

```hcl
provider "prisma-airs" {
  client_id     = var.panw_client_id
  client_secret = var.panw_client_secret
  tsg_id        = var.panw_tsg_id

  runtime {
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/aisec"
  }

  red_team {
    data_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/data-plane"
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/ai-red-teaming/mgmt-plane"
  }

  gateway {
    data_endpoint = "https://api.apps.paloaltonetworks.com/ai_gw/v2"
    admin_endpoint = "https://api.apps.paloaltonetworks.com/ai_gw/admin/v2"
  }

  supply_chain {
    data_endpoint = "https://api.sase.paloaltonetworks.com/aims/data"
    mgmt_endpoint = "https://api.sase.paloaltonetworks.com/aims/mgmt"
  }
}
```

For separate tenant identities, use [aliased provider configurations](../getting-started/configuration.md#separate-tenants). Client construction performs no API operations; entitlement and permissions are checked when a resource uses its product. Missing credentials produce a diagnostic for the product being used.

See the [exact provider schema](generated/provider.md) and [migration guide](../guides/migration.md).
