# Quick start

For a complete installation-to-cleanup walkthrough, use [Getting started](index.md). The examples here require the updated provider installed from [source](installation.md#use-the-updated-provider-from-source).

## Security profile

```hcl
resource "prisma-airs_runtime_security_profile" "example" {
  profile_name = "my-ai-security-profile"

  ai_security_profile {
    model_type = "default"

    model_protection {
      name   = "prompt-injection"
      action = "block"
    }
  }
}
```

Policy changes create a new AIRS UUID and revision while Terraform keeps the resource address. Destroy deletes every revision under the managed name.

## Custom REST target

```hcl
resource "prisma-airs_red_team_target" "app" {
  name = "my-ai-application"

  rest {
    api_endpoint    = "https://app.example.com/chat"
    request_headers = { "Content-Type" = "application/json" }
    request_body    = { prompt = "{INPUT}", temperature = 0, stream = false }
    response_body   = { output = "{RESPONSE}" }
    response_key    = "output"
  }
}
```

Target creation registers configuration without executing inference or connection validation. [Target authentication](../resources/red-team-target.md#authentication-and-state) shows credential blocks and replacement behavior.

## Review the result

```bash
terraform validate
terraform plan -out=changes.tfplan
terraform apply changes.tfplan
terraform plan -detailed-exitcode
```

Inspect the plan before applying. Protect saved plans alongside state because they can contain secret values. Choose a complete configuration from the [example catalog](../examples/index.md).
