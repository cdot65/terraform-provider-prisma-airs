# Skill Scanning Terraform scope

Source: published prisma-airs-go v0.7.0, consuming owner-supplied AgentGuard public preview 08212026 contracts (API version0.1.0). SDK package/wire names remain upstream identifiers; the Terraform product vocabulary is Skill Scanning within AI Supply Chain Security.

| Terraform ownership | SDK operations |
| --- | --- |
| Tenant instance resource/data source | Instances.Create/Get/Update/Delete |
| Individual policy rule resource | RuleInstances.List/Update and Rules.List |
| Catalog/effective policy discovery | Rules.List and RuleInstances.List |
| Trusted skill override resource/data source | SkillOverrides.List/Create/Delete |
| Existing scan detail, fingerprint lookup, and inventory | Scans.Get/Lookup/List |
| Findings and attack-chain discovery/detail | Scans.ListVulnerabilities/ListAttackChains/GetAttackChain |
| Scan/rule statistics | Statistics.Scans/Rules |

The authoritative management preview schema describes rule_configurations as a map of **rule_uuid** to desired state, with UUID property names. The generated SDK SkillSecurityRuleInstancesUpdateRequest and its HTTP contract test use that mapping. These are catalog rule UUIDs, distinct from rule-instance UUIDs. SDK handoff prose initially called them rule-instance UUIDs; the provider follows the pinned schema, and live no-op policy updates accepted the catalog UUID. That no-op is non-discriminating evidence; the authoritative preview schema establishes the mapping, and mock changed-state tests verify provider behavior.

Scans.UploadURL, Scans.UploadComplete, signed storage transfer, polling, and Scans.ExportCSV are operational workflows, not Terraform-owned lifecycle objects. Scans and catalog rules expose no delete operation. Policy destroy restores the adopted effective state; overrides have no update operation, so edits require replacement.

Instance resource exposes typed registration identity, support-account metadata, IAM control and authorization code. The required sensitive native HCL `registration_details` object owns the complete provisioning metadata, entitlements, deployment profiles and SDK extensions; updates resend it because PUT can deactivate removed profiles. Dedicated identity/authentication attributes cannot be overridden inside that object. GET cannot recover the full original registration payload, so import requires the complete desired configuration before an update. JSON strings are rejected; the instance data source returns complete SDK response details. Shared credentials remain authoritative across both SDK clients.
