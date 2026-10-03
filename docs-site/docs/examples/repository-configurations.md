# Repository configurations

## Start with a complete product project

The [Prisma AIRS Terraform examples repository](https://github.com/cdot65/prisma-airs-terraform-examples) provides one independent root per product. Each README covers prerequisites, environment credentials, configuration, apply, the next application step, updates, and cleanup.

For example, after loading management credentials:

```bash
git clone https://github.com/cdot65/prisma-airs-terraform-examples.git
cd prisma-airs-terraform-examples/examples/ai-runtime-security
cp terraform.tfvars.example terraform.tfvars
# Edit the nonsecret inputs and choose an unused prefix.
terraform init
terraform plan -out=create.tfplan
terraform apply create.tfplan
```

Choose another product from the [example catalog](index.md) and follow its prerequisites before planning. Gateway's project includes explicit inference and MCP helpers; the other projects distinguish policy registration from application scans or assessments.

## Explore provider-checkout scenarios

The provider checkout also retains these configurations:

| Directory | Purpose |
| --- | --- |
| `examples/security-profiles/` | Multiple profiles with different policy controls |
| `examples/model-security/` | Two named Hugging Face groups and rule discovery |
| `examples/ai-red-teaming/` | Custom REST endpoints, Bedrock, and prompt-set containers |
| `examples/gateway/` | A focused integration/binding/provider/routing chain |

Read the chosen README, provider version requirement, and variable definitions before applying. Some retained scenarios use project-specific endpoints or existing topic/DLP names; replace them with resources you control. Resource addresses are retained for existing users, so inspect any update or replacement against your own state.

Supply `PANW_MGMT_*` credentials and sensitive `TF_VAR_*` inputs through the environment or your secret manager. `terraform.tfvars.example` is a template for nonsecret inputs. The optional `scripts/terraform-env.sh` helper supports local development; [source installation](../getting-started/installation.md#use-the-updated-provider-from-source) is needed only when testing an unreleased provider build.

Network Broker targets require an existing channel and leave its lifecycle outside Terraform. The retained `litellm_multiturn` address does not imply multi-turn support.

The provider's Go acceptance tests cover lifecycle behavior. [Recorded verification](../development/sdk-upgrade-verification.md) distinguishes those tests from the application's own connectivity and assessment steps.
