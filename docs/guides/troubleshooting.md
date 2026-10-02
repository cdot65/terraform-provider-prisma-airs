---
page_title: "Troubleshooting"
---

# Troubleshooting

Start with the error from the actual operation and confirm which provider binary, tenant, and service it used.

| Symptom | Check |
| --- | --- |
| Native target block is unsupported | Select `~> 0.8.0` and run `terraform init -upgrade`; remove any development override and review migration |
| Management client is unavailable | Set all three `PANW_MGMT_*` credentials or explicit provider attributes |
| No active Model Security license | Use a tenant with the required entitlement |
| Profile name already exists | Import by name; creation does not adopt existing history |
| Customer app creation or rename rejected | Establish the app externally and import its name |
| Customer-app deployment code is ambiguous | Resolve the app's distinct deployment associations in AIRS |
| Imported API key has a null secret | The create-time key cannot be recovered or regenerated through import |
| Deployment-profile output rejected | Declare the output `sensitive = true` |
| Target plans replacement | Inspect native connection edits, removed auth/fields, category changes, or cleared channels |
| Prompt set or group disappears on refresh | Archived sets and tombstoned groups are treated as absent |

## Isolate a configuration problem

```bash
terraform providers
terraform providers schema -json
terraform validate
terraform plan
```

Validation proves schema compatibility; planning can perform API reads and establish live authorization. After apply, use `terraform plan -detailed-exitcode` to check for a stable state.

## Report a reproducible failure

Include the provider version or local commit, Terraform version, resource type, a minimal redacted configuration, and the operation that failed. Preserve meaningful error status and pagination behavior. Keep credentials, state, saved plans, auth codes, and raw API-key receipts out of issue reports.
