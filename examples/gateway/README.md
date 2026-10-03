# AI Gateway

Use an existing Gateway workspace. Supply `workspace_id`, `ai_provider_id`, and sensitive `upstream_api_key` through your secure variable source. Configure shared `PANW_MGMT_*` OAuth variables, then run `terraform init`, `terraform plan`, `terraform apply`, and a second plan. Destroy removes only owned objects and disables their workspace binding. See the documentation Gateway workflow for secret and archive behavior.
