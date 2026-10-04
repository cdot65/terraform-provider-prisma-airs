# AI Gateway workspace

`prisma-airs_gateway_workspace` coordinates workspace creation and archival with an explicit IAM scope. Available in provider v0.10.0 using published Go SDK v0.8.1.

## Example

```hcl
# Workspace: Own a dedicated scope and pass the workspace UUID to child resources.
variable "workspace_metadata" {
  description = "Tenant-approved metadata including every required workspace property."
  type        = map(string)
}

resource "prisma-airs_gateway_workspace" "applications" {
  name             = "Application gateway"
  scope_name       = "tf_gateway_apps_prod"
  scope_management = "managed"

  defaults = {
    metadata = var.workspace_metadata
  }
  usage_limits = [
    {
      type         = "tokens"
      credit_limit = 100000
    }
  ]
  rate_limits = [
    {
      type  = "requests"
      unit  = "rpm"
      value = 60
    }
  ]
}

resource "prisma-airs_gateway_config" "applications" {
  name         = "Application routing"
  workspace_id = prisma-airs_gateway_workspace.applications.id
  config = {
    provider = "openai"
    retry = {
      attempts = 1
    }
  }
}
```

The shared OAuth account needs Gateway admin and IAM permissions. Scope binding associates the workspace slug with a scope; it does not assign a role or grant access to the account. Existing access-policy grants remain an explicit prerequisite.

## Scope ownership

| Mode | Creation | Destruction |
| --- | --- | --- |
| `managed` | Create a new dedicated scope, create workspace, bind its returned slug | Archive workspace, confirm inactivity, delete the owned scope, confirm absence |
| `external` | Read the specified existing scope and create the workspace | Archive the workspace; make no IAM writes |

External owners maintain their bindings and role grants. A name collision never grants Terraform ownership. A managed scope must remain dedicated to this workspace; unrelated resource bindings stop binding updates and destruction. Read/replace binding operations have no atomic concurrency contract: keep the managed scope under exclusive ownership.

`scope_name` and `scope_management` changes require replacement. The detail API’s icon decoration is normalized out of `name`. `name` is a display label: edits preserve the workspace UUID, slug and scope name. Keep Terraform's default destroy-before-create ordering when reusing a scope name. `create_before_destroy` cannot create a second scope with the same name.

Workspace DELETE archives; historical rows remain. Refresh retains archived identities and plans replacement so the old owned scope is cleaned before a new workspace is created. Child resources reference `id`, establishing destruction order. If an owned scope disappears, the workflow plans replacement while retaining the service workspace status. If an external scope disappears, refresh preserves the workspace UUID and reports a warning; ordinary plans fail until its owner restores the scope. An explicit destroy still archives the workspace without IAM writes.

## Settings ownership

Configured `defaults`, `usage_limits`, and `rate_limits` are complete objects/collections. Removing a previously managed setting clears it; never configured settings remain unmanaged. Native fields remain visible in plans, without JSON-string inputs.

- Defaults support `config_id` and native object `metadata`. Omitted fields inside a configured defaults object clear the old default selection/metadata. Embedded credentials are rejected; use references. The API does not echo `allow_config_override`, so this field is excluded.
- The service permits one workspace usage policy. Credit changes preserve its UUID; changing measurement type or removing optional policy settings can clear/reconfigure its association. This may generate a new backing-policy UUID. Hidden backing-record deletion is not asserted.
- Rate-limit updates replace the whole collection; an empty list clears it. Empty usage lists use the API's explicit null clearing operation.
- Removing a configured icon clears it. Blank descriptions do not clear through this API; removing `description` stops managing it. Nonempty descriptions update normally.

`defaults.config_id` must reference a config that already exists. A config in this same workspace depends on workspace creation, so referencing that child during workspace creation produces a Terraform dependency cycle. Create the workspace/config first, then set its default in a subsequent apply.

## Import and recovery

```bash
terraform import prisma-airs_gateway_workspace.applications '<workspace-uuid>/<scope-name>'
```

Ordinary import defaults to `scope_management = "external"`. Import reads metadata and leaves optional settings unmanaged. Configure the settings deliberately to take ownership. A UUID-only import works only when the detail response supplies `scope_name`.

Explicit dedicated-scope adoption requires an existing scope bound only to this workspace slug:

```bash
terraform import prisma-airs_gateway_workspace.applications 'managed/<workspace-uuid>/<scope-name>'
```

Provisioning saves identities and `provisioning_stage` before follow-up writes. An uncertain scope POST is reconciled using its saved random `scope_ownership_token`; unrelated name collisions are never adopted. An uncertain workspace POST requires a complete active/archived inventory and a unique scope relationship. Ambiguous POSTs are not blindly replayed.

After a failure, inspect the remote objects and state. If Terraform marked the resource tainted and you intend to resume the retained workflow, run `terraform untaint <address>`, then review and apply the recovery plan. Failed binding resumes without creating another workspace. Failed cleanup retains the archived workspace and scope identity for retry. Inconclusive discovery requires the manual recovery below; a partial inventory cannot prove absence. A process crash before Terraform saves state can still require manual recovery; no external transaction journal is implied.

## Complete schema

See the [exact workspace schema](../reference/generated/prisma-airs_gateway_workspace.md) and [workspace workflow example](../examples/gateway-workspaces.md).

## Manual recovery when creation is uncertain

Stop concurrent applies. Keep a secure backup before changing state; state can contain sensitive values:

```bash
umask 077
terraform state pull > recovery.tfstate
terraform state show prisma-airs_gateway_workspace.applications
airs cli aigateway scopes get '<scope-name>'
airs cli aigateway workspaces get '<recovered-workspace-uuid>' --plane admin
```

Use the selected CLI tenant matching the provider. Recover the exact UUID from the create response, audit records or owner inspection. Verify its slug and tenant, the configured scope name, and that the scope description matches `Terraform workspace ownership <scope_ownership_token>` from state. Inspect **all** scope bindings. An incomplete list, a matching display name, or a missing response cannot establish absence or ownership.

For a confirmed workspace and its owned, dedicated scope, complete the missing binding outside Terraform, then replace the uncertain checkpoint with an explicit managed import:

```bash
airs cli aigateway scopes bind '<scope-name>' --workspace '<verified-workspace-slug>'
airs cli aigateway scopes get '<scope-name>'
terraform state rm prisma-airs_gateway_workspace.applications
terraform import prisma-airs_gateway_workspace.applications 'managed/<verified-workspace-uuid>/<scope-name>'
terraform plan
```

Proceed only if the verified scope contains exactly this workspace binding and no unrelated resources. These commands bind the scope; they do not assign access policies. Keep `scope_management = "managed"` in HCL. Review the plan before applying: imported optional settings are unmanaged until configured. State removal only forgets the checkpoint; it does not delete remote objects. If import fails, keep the backup and recover that same UUID before any new create/apply.

Alternatively, set `scope_management = "external"` in HCL, omit the binding command, remove the checkpoint and import `<verified-workspace-uuid>/<scope-name>`. Terraform then archives only the recovered workspace on destroy; the external owner handles bindings and any retained scope cleanup after confirming workspace inactivity. Do not claim managed adoption of an unbound scope.

For `scope_create_uncertain` with a **different or missing ownership token**, treat the scope as foreign. Do not bind, adopt or delete it. Back up state, remove the uncertain checkpoint, choose a fresh scope name and review a new plan. If a workspace was also created, recover/import that exact workspace first. If the token matches but no workspace UUID can be established, stop and resolve the create outcome through audit/owner inspection; neither deleting the scope nor retrying POST is safe based on an incomplete inventory.
