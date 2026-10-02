# Repository configurations

The checkout includes configuration roots in `examples/security-profiles/`, `examples/model-security/`, and `examples/ai-red-teaming/`. Read the README and variable definitions in the chosen directory before applying.

```bash
cd examples/ai-red-teaming
cp terraform.tfvars.example terraform.tfvars
../../scripts/terraform-env.sh validate
../../scripts/terraform-env.sh plan
```

Use [source installation](../getting-started/installation.md#use-the-updated-provider-from-source) for the updated schema. Set tenant OAuth values through the environment and target credentials through sensitive variables or your secret manager. Replace repository-specific endpoint examples with endpoints you control.

The Red Team configuration includes custom REST targets and Bedrock. Its retained historical resource addresses do not imply multi-turn support. Network Broker examples require an existing channel and leave its lifecycle outside Terraform.

The Go acceptance tests cover create, read, update, import, drift, archive/tombstone, and cleanup behavior. The [verification report](../development/sdk-upgrade-verification.md) records what was exercised live.
