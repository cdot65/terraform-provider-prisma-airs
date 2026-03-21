data "prisma-airs_content_scan" "example" {
  profile_name = "my-security-profile"

  prompt   = "What is the capital of France?"
  response = "The capital of France is Paris."
}

output "category" {
  value = data.prisma-airs_content_scan.example.category
}

output "action" {
  value = data.prisma-airs_content_scan.example.action
}
