# prisma-airs_red_team_categories

Reads available attack categories from Prisma AIRS Red Team API.

## Example Usage

```hcl
data "prisma-airs_red_team_categories" "all" {}

output "categories" {
  value = data.prisma-airs_red_team_categories.all.items
}
```

## Argument Reference

No arguments required.

## Attribute Reference

- `items` - List of attack categories. Each item contains:
    - `name` - Category name.
    - `description` - Category description.
    - `subcategories` - List of subcategories.
