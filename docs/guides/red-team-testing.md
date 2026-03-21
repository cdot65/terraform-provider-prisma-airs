# Red Team Testing

This guide covers automated red team testing with the Prisma AIRS Terraform provider.

## Step 1: Create a Target

```hcl
resource "prisma-airs_red_team_target" "chatbot" {
  name        = "customer-chatbot"
  target_type = "APPLICATION"
  description = "Customer-facing chatbot application"

  connection = jsonencode({
    type     = "REST"
    endpoint = "https://chatbot.example.com/api/chat"
    headers = {
      "Authorization" = "Bearer ${var.chatbot_api_key}"
      "Content-Type"  = "application/json"
    }
  })
}
```

## Step 2: Run a Static Scan

```hcl
resource "prisma-airs_red_team_scan" "security_audit" {
  target_uuid = prisma-airs_red_team_target.chatbot.id
  job_type    = "STATIC"
  categories  = ["SECURITY", "SAFETY"]
}
```

## Step 3: Review Results

```hcl
data "prisma-airs_red_team_reports" "audit_report" {
  job_id      = prisma-airs_red_team_scan.security_audit.id
  report_type = "static"
}

output "risk_rating" {
  value = data.prisma-airs_red_team_reports.audit_report.risk_rating
}
```

## Step 4: Custom Attack Testing

Create custom prompt sets for targeted testing:

```hcl
resource "prisma-airs_red_team_custom_prompt_set" "injection_tests" {
  name        = "custom-injection-tests"
  description = "Custom prompt injection test cases for our domain"
}

resource "prisma-airs_red_team_scan" "custom_test" {
  target_uuid     = prisma-airs_red_team_target.chatbot.id
  job_type        = "CUSTOM"
  prompt_set_uuid = prisma-airs_red_team_custom_prompt_set.injection_tests.id
}
```

## Check Quota

```hcl
data "prisma-airs_red_team_quota" "current" {}

output "scans_remaining" {
  value = data.prisma-airs_red_team_quota.current.remaining
}
```

## Check Available Categories

```hcl
data "prisma-airs_red_team_categories" "all" {}

output "attack_categories" {
  value = data.prisma-airs_red_team_categories.all.items[*].name
}
```
