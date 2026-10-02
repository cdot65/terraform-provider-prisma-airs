# Authentication

The provider manages AIRS configuration through OAuth client credentials. Use the same SCM service-account setup as the AIRS CLI and SDK management clients.

## 1. Obtain a service account

In your tenant's Strata Cloud Manager account, open the service-account/API-access area and create or select an OAuth service account. Record its client ID, client secret, and TSG ID in your secret manager. Console labels can vary by account.

Follow Palo Alto Networks’ [service-account creation guide](https://docs.paloaltonetworks.com/common-services/identity-and-access-access-management/manage-identity-and-access/add-service-accounts) for the console steps. Grant roles for the services this provider will manage.

## 2. Grant service access

| Domain | Access needed |
| --- | --- |
| Management | Profile, topic, deployment-profile, API-key, and customer-app actions used by your configuration |
| Model Security | Active Model Security entitlement and group/rule permissions |
| Red Team | Active Red Team entitlement and target/prompt-set permissions |

A valid OAuth token does not establish a service entitlement. A superuser role does not make an unsupported API route available. The customer-app resource uses the supported paginated list route.

## 3. Load credentials

Supply secret values through your existing secret-store integration. The required variable names are:

```bash
export PANW_MGMT_CLIENT_ID="your-client-id"
export PANW_MGMT_CLIENT_SECRET="your-client-secret"
export PANW_MGMT_TSG_ID="your-tsg-id"
```

```hcl
provider "prisma-airs" {}
```

The provider passes this resolved credential set to all three management clients. SDK-specific `PANW_MODEL_SEC_CLIENT_*` and `PANW_RED_TEAM_CLIENT_*` variables do not select separate identities through this provider. Use [provider aliases](configuration.md#separate-tenants) when a configuration needs different tenants or credentials.

## 4. Check access

Start with a read appropriate to the service you need, such as `prisma-airs_deployment_profiles` or `prisma-airs_model_security_rules`. `terraform validate` checks configuration without proving live authorization; `terraform plan` reads the configured data sources.

## Token and secret handling

The Go SDK acquires, caches, and refreshes management OAuth tokens. A Runtime scanning API key is an application credential and does not replace the provider's OAuth service account.

Provider attributes take precedence over their corresponding environment variables. Terraform does not load `.env` automatically. In repository examples, `scripts/terraform-env.sh` supplies the local helper; export variables for ordinary configurations and CI.

Mark secret input variables and outputs sensitive, protect the state backend, and keep `.env`, state, and saved plans out of source control. See [import and state](../guides/import-and-state.md).
