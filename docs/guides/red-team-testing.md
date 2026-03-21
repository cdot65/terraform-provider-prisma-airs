# Red Team Testing

This guide covers managing red team targets and custom prompt sets with the Prisma AIRS Terraform provider.

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

## Step 2: Create Custom Prompt Sets

Create custom prompt sets for targeted testing:

```hcl
resource "prisma-airs_red_team_custom_prompt_set" "injection_tests" {
  name        = "custom-injection-tests"
  description = "Custom prompt injection test cases for our domain"
}
```
