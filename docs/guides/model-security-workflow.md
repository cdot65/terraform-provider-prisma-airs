# Model Security Workflow

This guide covers managing model security groups and reviewing security rules with the Prisma AIRS provider.

## Step 1: Create a Security Group

```hcl
resource "prisma-airs_model_security_group" "ml_models" {
  name        = "production-models"
  description = "Security group for production ML models"
  source_type = "HUGGING_FACE"
}
```

## Step 2: Review Security Rules

```hcl
data "prisma-airs_model_security_rules" "all" {}

output "available_rules" {
  value = data.prisma-airs_model_security_rules.all.rules[*].name
}
```
