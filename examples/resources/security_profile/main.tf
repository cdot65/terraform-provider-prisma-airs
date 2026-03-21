resource "prisma-airs_security_profile" "example" {
  profile_name = "example-profile"

  policy = jsonencode({
    "ai-security-profiles" = [
      {
        "model-type" = "default"
        "model-configuration" = {
          "latency" = {
            "inline-timeout-action" = "allow"
            "max-inline-latency"    = 30
          }
          "model-protection" = [
            {
              name   = "prompt-injection"
              action = "block"
            },
            {
              name   = "toxic-content"
              action = "alert"
              "toxic-category-list" = [
                { category = "harassment", action = "alert" },
                { category = "violence", action = "block" }
              ]
            }
          ]
        }
      }
    ]
  })
}
