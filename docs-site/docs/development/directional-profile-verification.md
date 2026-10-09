# Directional profile live verification — 2026-10-09

The provider's Go SDK v0.9.0 integration was exercised against the live Runtime management API on 2026-10-09 UTC. The run used the complete [directional security profile example](../examples/directional-security-profile.md), a local provider binary built with Go 1.27.1, Terraform 1.16.4, and the published Go module without a local replacement. The six revisions below are actual service receipts from the disposable profile `tf-live-directional-5fbaff621178`.

## Reproduce the run

Load `PANW_MGMT_CLIENT_ID`, `PANW_MGMT_CLIENT_SECRET`, and `PANW_MGMT_TSG_ID` into the environment using your existing credential manager. Optional endpoint overrides use `PANW_MGMT_ENDPOINT` and `PANW_MGMT_TOKEN_ENDPOINT`.

```bash
make build
python3 -u e2e/directional-live.py
```

The runner copies `examples/directional-security-profile/main.tf` into a private temporary directory and uses a Terraform development override for the built provider. It chooses a fresh profile name and borrows an existing DLP profile by name/version without changing that DLP profile. It refuses an existing profile name before acquiring cleanup ownership. Run it in a tenant where creating disposable Runtime profiles is appropriate.

Every API receipt and full Terraform output is retained outside the repository with owner-only permissions. Stdout reports verification results without credentials, tenant identifiers, DLP reference names, or audit identities. A receipt SHA-256 manifest accompanies the private logs. Cleanup runs on failure as well as success and inventories every revision under the owned name.

## Captured Terraform output

These lines are copied from the successful live run. The excerpt omits development-override warnings and the expanded policy listing; it retains the actual Terraform results.

```text
UTC: 2026-10-09T21:25:28.779808+00:00
Owned profile: tf-live-directional-5fbaff621178
Borrowed DLP reference: read-only=true
validate: exit=0
create: exit=0
Plan: 1 to add, 0 to change, 0 to destroy.
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
refresh-stable: exit=0
No changes. Your infrastructure matches the configuration.
import: exit=0
Import successful!
import-stable: exit=0
No changes. Your infrastructure matches the configuration.
response-update: exit=0
Plan: 0 to add, 1 to change, 0 to destroy.
Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
Response-only update: prompt/tool-call/tool-response preserved=true
response-stable: exit=0
No changes. Your infrastructure matches the configuration.
shared-update: exit=0
Plan: 0 to add, 1 to change, 0 to destroy.
Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
Shared latency update: all four directions preserved=true
detector-remove: exit=0
Plan: 0 to add, 1 to change, 0 to destroy.
Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
removal-stable: exit=0
No changes. Your infrastructure matches the configuration.
Prompt detector removal: persisted=true
drift-plan: exit=2
Plan: 0 to add, 1 to change, 0 to destroy.
Remote response severity drift: detected=true
drift-repair: exit=0
Plan: 0 to add, 1 to change, 0 to destroy.
Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
final-stable: exit=0
No changes. Your infrastructure matches the configuration.
destroy: exit=0
Plan: 0 to add, 0 to change, 1 to destroy.
Destroy complete! Resources: 1 destroyed.
Cleanup: remaining_revisions=0 observed_revision_ids=6
```

## Service revisions and preservation

| Revision | Operation | Actual returned profile UUID |
| --- | --- | --- |
| 1 | Create all four directions | `b7f687fe-4d8a-4991-89d7-6cc6a7e8e7dc` |
| 2 | Change response toxicity high-confidence severity from medium to high | `8d151804-3825-44a1-8e5b-ff3bf28e7b22` |
| 3 | Change shared maximum inline latency from 5 to 6 | `3b80fbb7-72ce-4a09-8c70-1c11742917c3` |
| 4 | Remove prompt injection from the prompt direction | `f9e544ab-54e6-46a7-9060-21ffaca2b323` |
| 5 | Change response severity remotely from high to low | `20a97262-baf4-462f-8fa6-15b15efd4f6a` |
| 6 | Apply Terraform to restore response severity to high | `e58b3fc5-a535-4f4e-afa2-3c844f142ba1` |

The response-only update's real plan contained:

```text
~ severity_by_confidence {
    ~ high     = "medium" -> "high"
      # (1 unchanged attribute hidden)
  }
```

The runner compared the entire prompt, tool-call, and tool-response JSON objects before and after revision 2. They were equal. It then compared all four direction objects before and after the shared latency edit; those were also equal. Each command launches a separate Terraform process, exercising persisted provider private state.

The removed detector was absent from revision 4's API policy and did not reappear after refresh. The remote mutation used a direct HTTP PUT scoped to this disposable profile; Terraform's subsequent plan returned exit code 2 and showed `high = "low" -> "high"`. Applying that plan restored the policy, and the following plan returned exit code 0.

Independent inspection of the captured creation GET confirmed both shared boolean values were explicit `false`, prompt/response/tool-call inline masking was `true`, and tool-response inline masking was omitted. GET omitted top-level `dlp_tenant_id`. No inference requests were issued; this verifies management-plane Terraform lifecycle behavior rather than detector efficacy.

## Legacy profile live checks

The existing legacy HCL tests were run against the same live tenant with SDK v0.9.0:

```bash
TF_ACC=1 go test -race ./internal/provider \
  -run '^TestAccSecurityProfileResource_(revisionsAndRename|richPolicy)$' \
  -v -count=1 -timeout 20m
```

The successful run captured:

```text
--- PASS: TestAccSecurityProfileResource_revisionsAndRename (49.09s)
--- PASS: TestAccSecurityProfileResource_richPolicy (33.17s)
PASS
```

These checks cover changing revision UUIDs, import by name, remote action drift, following the highest remaining revision after remote deletion, rename preserving the old name's history, and refusing to take ownership of an existing name. The rich-policy test verifies explicit inline masking `false`, an action-only update, stable subsequent configuration, and switching between two existing read-only DLP references without carrying unrelated metadata across identities. Cleanup removes every revision of each uniquely named fixture and verifies absence.

The first legacy runs exposed two outdated test assertions: literal JSON comparison treated omitted unknown optional/computed fields differently from prior null values, and import verification required write-only DLP tenant metadata to appear in GET. The corrected tests still compare every populated policy leaf and ignore only `dlp_tenant_id` during import verification. Both final live tests passed; the earlier failed runs are not counted as successful verification.

## Offline and documentation checks

`make check`, `make build`, and `make generate` passed. `make docs-check` validated 98 HCL snippets, including eight complete roots, checked the generated schemas and Registry pages, built Docusaurus, and passed 11 browser tests plus nine visual parity comparisons. The directional example also passed live `terraform validate` and its post-apply plans returned exit code 0.

## Additional evidence

The original sanitized POST fixture and mock lifecycle tests remain useful for null lists, empty lists/objects, unknown direction keys, future extension fields, failed reads, and old-state compatibility. Those cases are fixture/mocked evidence and do not become live service claims because the directional lifecycle passed.

The historical [SDK upgrade verification](sdk-upgrade-verification.md) records earlier provider versions separately.
