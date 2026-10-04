# Adoption fidelity and adapter verification

This candidate extends published v0.11.0 using the existing Go SDK v0.8.1. No SDK or tenant-access changes are required for the provider fixes demonstrated here. It remains unreleased pending independent Claude Code review scoring at least 9.1/10; the actual CLI currently returns HTTP429 for a monthly spend limit. No score is claimed.

## Provider behavior

- Red Team adapter ownership includes decoded scripts, complete variable key/type inventory, explicit draft writes, opt-in broker execution, import, and deletion. Discovery reads use bounded pagination and do not execute scripts.
- Adapter target requests use CUSTOM_TARGET_ADAPTER, an adapter UUID, Network Broker connectivity, null connection parameters, and a **null response mode**. A real REST-mode attempt returned HTTP400; a one-variable null-mode correction succeeded and the probe was deleted.
- Target import recovers usable observable request/response payloads, request headers, and OAuth injection templates. Configured payload types and credentials are preserved on ordinary refresh. Missing secrets allow read-only no-op configuration; full writes reject incomplete or masked inputs. Explicitly redacted payload/header paths are retained in `unavailable_fields` so metadata edits cannot silently drop credentials.
- Existing empty streaming stop strings remain adoptable; writes require nonempty stop settings. Empty Runtime toxic-content actions are accepted only with explicit category settings, without assigning allow/block semantics.
- Newly imported Gateway JSON nulls use native HCL null types. Repeated refresh preserves their shape; real null-to-value drift remains visible. Older typed-null state can be backed up and reimported without API writes.

## Actual validation

Focused tests first reproduced lost target payloads, the Gateway imported-null type mismatch, and repeated-refresh null wrapping. The fixed tests cover recovery, missing-input write refusal, masked-secret rejection, observable drift, and redacted-header metadata-update safety. A real Terraform mock lifecycle covers adapter create/import/update/discovery/no-op/destroy and secret retention. Required `make check` passes formatting, vet, CI-version lint, and race tests across the provider.

Vulture disposable lifecycle validation created a draft adapter, updated a plain variable while retaining a redacted SECRET, opted into bounded text-only broker execution, and registered an adapter-backed target. The script asserted the original stored secret without printing it. Adapter activation succeeded despite the channel reporting an outdated client; this proves this controlled text path, not every broker capability. Reimport matched the default transient validation prompt, then an ordinary live plan exited 0. Terraform destroyed the target before the adapter, and separate SDK reads confirmed HTTP404 for both. The adapter list returned to its three original records. A second disposable API probe was deleted as well.

Private state copies exercise existing Vulture adoption without changing established state or tenant configuration. The primary 339-resource copy plans cleanly. Both existing routing configurations were reimported and their typed-null HCL workarounds removed; an ordinary plan exited 0. Recovery preview exposed real request headers on two targets that the old provider omitted; matching values stay in private inputs. After matching all observable recovery settings, twelve recovery objects were transferred between private state copies and three adapters plus three adapter-backed targets were imported. The combined preview contains 357 managed objects and an ordinary live plan exits 0. Original primary/recovery state hashes still match their pre-preview copies; established state and desired edits remain preserved until the reviewed release is installed.

A disposable Runtime profile also preserved the exact observed empty toxic-content/category representation through create, a storage-masking update that created a new revision, clean ordinary plans, and destruction of its owned history. The public adapter example itself ran unchanged from a private copy: draft apply, activation/target registration, clean plan, and destroy all passed. Its README links recorded sanitized output and source hashes.

## Remaining API/access diagnoses

Fresh raw GETs independently bypass the Go SDK and confirm these six blockers with the same tenant credentials:

| Cases | Observable result | Next action |
| --- | --- | --- |
| Four Gateway workspace IAM scopes | Scope detail HTTP404; absent from 11 tenant-visible scopes | Authorized tenant administrator should compare workspace scope references, current IAM scope names, visibility, and bindings. Decide whether a reference is stale or visibility is restricted before approving a repair. |
| One listed Gateway workspace provider | Active row in the workspace list; UUID and slug detail HTTP404 in the same workspace | Supply sanitized list/detail evidence and privately retained identifiers/request timestamps to the service owner. Determine stale listing, lookup inconsistency, or access filtering before modifying or recreating it. |
| Skill Scanning instance | Instance detail HTTP403; seven policy rule instances remain readable | Verify instance-management entitlement and role grants separately from rule-read permissions. Instance existence is not established by a forbidden GET; do not infer onboarding is absent or attempt registration PUT. |

These are service/access observations, not proven SDK bugs. Existing scope, membership, credentials, instance registration, and broker installation remain unchanged. No external support ticket or message was sent.
