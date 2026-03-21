# Model Security Workflow

This guide covers the end-to-end workflow for scanning ML models with the Prisma AIRS provider.

## Step 1: Create a Security Group

```hcl
resource "prisma-airs_model_security_group" "ml_models" {
  name        = "production-models"
  description = "Security group for production ML models"
  source_type = "HUGGING_FACE"
}
```

## Step 2: Create a Scan

```hcl
resource "prisma-airs_model_scan" "bert" {
  name                = "scan-bert-base"
  source_type         = "HUGGING_FACE"
  security_group_uuid = prisma-airs_model_security_group.ml_models.id

  source = jsonencode({
    model_name = "bert-base-uncased"
  })

  labels = {
    team        = "ml-platform"
    environment = "production"
  }
}
```

## Step 3: Check Results

```hcl
data "prisma-airs_model_scan_evaluations" "bert_results" {
  scan_uuid = prisma-airs_model_scan.bert.id
}

data "prisma-airs_model_scan_violations" "bert_violations" {
  scan_uuid = prisma-airs_model_scan.bert.id
}

output "scan_outcome" {
  value = prisma-airs_model_scan.bert.eval_outcome
}

output "violations_found" {
  value = length(data.prisma-airs_model_scan_violations.bert_violations.items)
}
```

## Step 4: Review Security Rules

```hcl
data "prisma-airs_model_security_rules" "all" {}

output "available_rules" {
  value = data.prisma-airs_model_security_rules.all.items[*].name
}
```
