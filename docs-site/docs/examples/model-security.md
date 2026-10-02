# Model Security

Use a tenant with active Model Security entitlement. This complete configuration creates a group and reads the rule catalog; it does not upload models or run scans.

```hcl
terraform {
  required_providers {
    prisma-airs = {
      source = "cdot65/prisma-airs"
    }
  }
}

provider "prisma-airs" {}

resource "prisma-airs_model_security_group" "models" {
  name        = "terraform-models"
  description = "Models managed through Terraform"
  source_type = "HUGGING_FACE"
}

data "prisma-airs_model_security_rules" "catalog" {}

output "rule_names" {
  value = data.prisma-airs_model_security_rules.catalog.rules[*].name
}
```

Changing `source_type` plans group replacement. Destroy leaves the old group tombstoned in the service, and Terraform treats it as absent. Verify the group and a no-change plan after apply.
