---
title: Skill Scanning Instance
---

# Skill Scanning Instance

Manages a tenant's Skill Scanning instance through create, read, update, and delete. Existing instances must be imported; creation checks for an existing instance before writing. This is tenant onboarding, not an individual skill scan.

```hcl
# Registration: Supply the entire onboarding payload; import existing tenant instances first.
resource "prisma-airs_supply_chain_skill_scanning_instance" "tenant" {
  tenant_id            = var.skill_tenant_id
  support_account_id   = var.support_account_id
  created_by           = "terraform@example.com"
  support_account_name = "Example"
  registration_details = {
    region        = "us"
    license_name  = var.skill_license_name
    entitlements  = var.skill_entitlements
    tsg_instances = var.skill_deployment_profiles
  }
  iam_controlled    = false
  auth_code         = var.skill_auth_code
  auth_code_version = 1
}
```

Changing `tenant_id` requires replacement. `auth_code` is a true Terraform write-only input (Terraform 1.11+): it is read from configuration and never stored in plan or state. Set `auth_code_version` when supplying a code; increment that nonsecret version to apply a new code or, with the code omitted, clear it with explicit null. Other instance updates do not resend the code. Import cannot recover it.

`registration_details` is the complete desired provisioning object in native HCL. Supply the entitled license, region, entitlements and deployment profiles from your onboarding payload; SDK extensions are supported. PUT resends this object and can deactivate profiles omitted from it. It is sensitive but stored in Terraform state, so secure the state backend. Identity and authorization fields use their dedicated attributes and cannot be overridden inside it.

Registration metadata (`created_by`, `support_account_id`, and `support_account_name`) retains the configured values because GET can return audit values or normalized/omitted fields. Import initializes them from the observed response; use the instance data source for current registration details. `iam_controlled` is a retained request-only setting: explicit `false` is sent, removal sends null, and GET cannot detect its drift. No JSON request document is accepted.

Destroy deletes the tenant instance and requires a successful receipt plus confirmed absence. Review this lifecycle before applying destroy to an existing tenant. The API does not expose individual scan deletion.

```bash
terraform import prisma-airs_supply_chain_skill_scanning_instance.tenant '<tenant-id>'
```

GET cannot recover the complete original provisioning payload. The first apply after import always issues a full PUT because the required `registration_details` is absent from imported state. Verify and supply the entire onboarding payload before that first apply. Import by itself only reads. Instance CRUD has mock Terraform lifecycle coverage; live instance mutations remain unverified because the available live instance read returned HTTP 403.

Skill Scanning belongs to **AI Supply Chain Security**. Configure both base URLs in `supply_chain`, using shared OAuth credentials. The Go SDK package name remains `aisec/agentguard`; Terraform names and guides use Skill Scanning.

```hcl
provider "prisma-airs" {
  supply_chain {
    skill_scanning_data_endpoint = "https://api.apps.paloaltonetworks.com/aiag/data"
    skill_scanning_mgmt_endpoint = "https://api.apps.paloaltonetworks.com/aiag/mgmt"
  }
}
```

Alternatively set `PANW_SKILL_SCANNING_DATA_ENDPOINT` and `PANW_SKILL_SCANNING_MGMT_ENDPOINT`. The established SDK variables `PANW_AGENT_GUARD_*_ENDPOINT` are fallback aliases. No endpoint is inferred from credentials.

See [Skill Scanning workflow](../guides/skill-scanning-workflow.md) and the exact schema in [Supply Chain Security](../reference/index.md).
