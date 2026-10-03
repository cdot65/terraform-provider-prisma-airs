# Model Security Workflow

This guide covers managing model security groups and reviewing security rules with the Prisma AIRS provider.

## Step 1: Create a Security Group

```hcl
# Model group: Organize models; onboarding and scans are separate operations.
resource "prisma-airs_supply_chain_security_group" "ml_models" {
  name        = "production-models"
  description = "Security group for production ML models"
  source_type = "HUGGING_FACE"
}
```

## Step 2: Review Security Rules

```hcl
# Discovery: Read existing metadata without changing remote configuration.
data "prisma-airs_supply_chain_security_rules" "all" {}

# Outputs: Expose results for the next configuration or application step.
output "available_rules" {
  value = data.prisma-airs_supply_chain_security_rules.all.rules[*].name
}
```
