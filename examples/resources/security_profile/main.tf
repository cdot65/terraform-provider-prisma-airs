resource "prisma-airs_security_profile" "example" {
  profile_name = "example-profile"

  policy = jsonencode({
    injection = {
      action = "block"
    }
    toxic_content = {
      action = "alert"
    }
  })
}
