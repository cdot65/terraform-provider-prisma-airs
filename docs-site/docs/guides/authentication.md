# Authentication guide

Follow [Authentication](../getting-started/authentication.md) for the shared SCM service-account walkthrough, required environment variables, and access checks.

## Credential precedence

For each provider attribute, an explicit value takes precedence over its mapped environment variable. `client_id`, `client_secret`, and `tsg_id` resolve from `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` and are passed to every service client.

Service-specific endpoint variables are supported. Separate SDK service credential prefixes are not exposed by this provider; use explicit credentials on aliased providers for separate identities. See [Configuration](../getting-started/configuration.md) and [Environment variables](../reference/environment-variables.md).

## Local environment files

From an example directory such as `examples/model-security/`, the repository helper loads `.env` before invoking Terraform:

```bash
../../scripts/terraform-env.sh plan
```

For configurations outside this repository, export variables through your normal environment-loading or secret-store workflow. Terraform itself does not load `.env` files.
