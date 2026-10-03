# Model Security

Creates two independently named Hugging Face model-security groups and reads the existing rule catalog. This configures collections; it does not scan a model.

## Resources

| Resource | Source | Description |
|----------|--------|-------------|
| `hugging_face` | Hugging Face Hub | Monitor for supply chain attacks, malicious weights |
| `custom_models` | Hugging Face Hub | Second collection; the legacy Terraform address is retained |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `prisma-airs_supply_chain_security_rules.all` | All model security rules |

## Before you start

Use a tenant with Model Security enabled. Both groups declare `source_type = "HUGGING_FACE"`; choose distinct names for the collections you intend to manage.

Load `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` from your secret store into the environment. See the [authentication guide](https://cdot65.github.io/terraform-provider-prisma-airs/getting-started/authentication/).

Copy `terraform.tfvars.example` to `terraform.tfvars` and set `group_prefix` to distinguish your example resources.

## Apply and inspect

```bash
terraform init
terraform validate
terraform plan -out=create.tfplan
terraform apply create.tfplan
terraform output
```

After editing the configuration, review another saved plan before applying it. An unchanged `terraform plan -detailed-exitcode` should exit 0.

Outputs expose the group IDs and catalog rule count. Change names or descriptions to practice updates. Changing `source_type` replaces a group.

## Clean up

Use the same inputs, credentials, and state to remove the owned resources:

```bash
terraform plan -destroy -out=cleanup.tfplan
terraform apply cleanup.tfplan
```

Deleted groups remain visible as tombstones in the service. For the broader workflow, use the [public Supply Chain Security project](https://github.com/cdot65/prisma-airs-terraform-examples/tree/main/examples/ai-supply-chain-security).

## Files

| File | Purpose |
|------|---------|
| `main.tf` | Provider configuration |
| `variables.tf` | Variable declarations |
| `terraform.tfvars.example` | Example variable values (copy to `terraform.tfvars`) |
| `groups.tf` | Model security group resources |
| `data.tf` | Model security rules data source |
| `outputs.tf` | Group IDs, names, state, and rule count |
