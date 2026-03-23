# Model Security

Manages model security groups for monitoring AI models from various sources. Reads existing model security rules for integration with other tooling.

## Resources

| Resource | Source | Description |
|----------|--------|-------------|
| `hugging_face` | Hugging Face Hub | Monitor for supply chain attacks, malicious weights |
| `custom_models` | Custom/Internal | Monitor internally-trained or fine-tuned models |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `prisma-airs_model_security_rules.all` | All model security rules |

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars if needed

# Set provider credentials
export PANW_MGMT_CLIENT_ID="..."
export PANW_MGMT_CLIENT_SECRET="..."
export PANW_MGMT_TSG_ID="..."

terraform init
terraform plan
terraform apply
```

## Files

| File | Purpose |
|------|---------|
| `main.tf` | Provider configuration |
| `variables.tf` | Variable declarations |
| `terraform.tfvars.example` | Example variable values (copy to `terraform.tfvars`) |
| `groups.tf` | Model security group resources |
| `data.tf` | Model security rules data source |
| `outputs.tf` | Group IDs, names, state, and rule count |
