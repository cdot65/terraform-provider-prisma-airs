# prisma-airs_red_team_reports

Reads scan report data from Prisma AIRS Red Team API.

## Example Usage

```hcl
data "prisma-airs_red_team_reports" "latest" {
  job_id      = prisma-airs_red_team_scan.static_test.id
  report_type = "static"
}

output "risk_score" {
  value = data.prisma-airs_red_team_reports.latest.risk_score
}
```

## Argument Reference

- `job_id` - (Required) Job ID of the scan.
- `report_type` - (Required) Type of report. Valid values: `static`, `dynamic`.

## Attribute Reference

- `risk_score` - Overall risk score.
- `risk_rating` - Risk rating (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`).
- `categories` - Category-level results.
- `remediation` - Remediation recommendations.
