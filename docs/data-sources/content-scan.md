# prisma-airs_content_scan

Performs a synchronous content scan using the AI Runtime Security Scan API.

## Example Usage

```hcl
data "prisma-airs_content_scan" "check_prompt" {
  profile_name = "my-security-profile"

  prompt   = "What is the capital of France?"
  response = "The capital of France is Paris."
}

output "verdict" {
  value = data.prisma-airs_content_scan.check_prompt.category
}

output "action" {
  value = data.prisma-airs_content_scan.check_prompt.action
}
```

## Argument Reference

- `profile_name` - (Optional) AI security profile name. Falls back to provider `profile_name`.
- `profile_id` - (Optional) AI security profile ID. One of `profile_name` or `profile_id` is required.
- `prompt` - (Optional) Prompt content to scan.
- `response` - (Optional) Response content to scan.
- `context` - (Optional) Context content to scan.
- `code_prompt` - (Optional) Code prompt content to scan.
- `code_response` - (Optional) Code response content to scan.

At least one content field must be provided.

## Attribute Reference

- `scan_id` - The scan ID.
- `report_id` - The report ID.
- `category` - Detection category (`benign`, `malicious`, `unknown`).
- `action` - Recommended action (`allow`, `block`, `alert`).
- `prompt_detected` - Detection results for prompt content.
- `response_detected` - Detection results for response content.
