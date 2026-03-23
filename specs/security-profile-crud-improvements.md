# Security Profile CRUD — Gap Analysis & Implementation Plan

## Current State

All four CRUD operations exist but two rely on workarounds for API limitations.

| Operation | Works? | Implementation | Root Cause of Limitation |
|-----------|--------|----------------|--------------------------|
| **Create** | Yes | SDK `Profiles.Create` (POST) | 409 conflict requires fallback lookup |
| **Read** | Partial | `findProfileByID()` — List + paginate + filter | **No GET-by-ID endpoint in the API** |
| **Update** | Yes | SDK `Profiles.Update` (PUT) | Direct, no issues |
| **Delete** | Partial | SDK `Profiles.Delete` (DELETE) | **API returns unparseable JSON response** |
| **Import** | Partial | `findProfileByName()` — List + paginate + filter | Same as Read — no GET endpoint |

---

## Issue 1: Read Has No Native GET Endpoint

### Problem

The AIRS Management API has **no** `GET /v1/mgmt/profile/{id}` endpoint. The only way to
retrieve a single profile is to list all profiles with pagination and filter client-side.

There IS a `GET /v1/mgmt/profile?profile_name=<name>` endpoint, but it **times out at 30s+**
(documented in SDK `API_ISSUES.md` #1), so the SDK doesn't use it.

### Current Workaround (Provider)

```go
// findProfileByID iterates all pages (limit=100) until match or exhaustion
func findProfileByID(ctx context.Context, client *management.Client, profileID string) *management.SecurityProfile {
    offset := 0
    limit := 100
    for {
        listResp, err := client.Profiles.List(ctx, management.ListOpts{Limit: limit, Offset: offset})
        // ... iterate listResp.Items, match by ProfileID
    }
}
```

### Risks

1. **O(n) API calls** — for n profiles, worst case is `ceil(n/100)` HTTP requests per
   `terraform plan/apply`. With 500 profiles and 10 managed resources, that's 50 API calls
   just for Read.
2. **No timeout protection** — if the API is slow, the pagination loop stalls indefinitely.
   No `context.WithTimeout` wrapper.
3. **Silent removal on API errors** — if `List` returns an error, `findProfileByID` returns
   nil, and the provider calls `resp.State.RemoveResource(ctx)`. This means a transient API
   error causes Terraform to think the resource was deleted, triggering a re-create on next apply.
4. **Race condition** — profile could be deleted between List pages; no consistency guarantee.

### Fix Plan

#### Layer 1: SDK (`prisma-airs-go`)

1. **Add `Profiles.Get(ctx, profileID)` method** that calls `GET /v1/mgmt/profile/uuid/{profileID}`.
   - The Update endpoint already uses `/v1/mgmt/profile/uuid/{id}` with PUT — test whether
     GET on the same path works. The OpenAPI spec may not document it, but many REST APIs
     support GET on the same resource path.
   - If the API does NOT support GET by UUID, open a feature request with Palo Alto Networks:
     "Add `GET /v1/mgmt/profile/uuid/{profileID}` endpoint for single-resource retrieval."

2. **If GET-by-UUID is not possible**, improve `GetByName` to handle pagination:
   ```go
   func (c *ProfilesClient) GetByName(ctx context.Context, name string) (*SecurityProfile, error) {
       offset := 0
       limit := 100
       for {
           resp, err := c.List(ctx, ListOpts{Limit: limit, Offset: offset})
           if err != nil {
               return nil, err
           }
           for _, p := range resp.Items {
               if p.ProfileName == name {
                   return &p, nil
               }
           }
           if len(resp.Items) < limit {
               return nil, aisec.NewAISecSDKError("profile not found: "+name, aisec.ClientSideError)
           }
           offset += limit
       }
   }
   ```
   Currently `GetByName` uses `Limit: 1000` without pagination — misses profiles beyond 1000.

#### Layer 2: Provider (`terraform-provider-prisma-airs`)

3. **Add timeout to Read** — wrap the lookup in a context with deadline:
   ```go
   readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
   defer cancel()
   found := findProfileByID(readCtx, r.client, state.ProfileID.ValueString())
   ```

4. **Distinguish "not found" from "API error"** — change `findProfileByID` signature to
   return `(*SecurityProfile, error)` instead of just `*SecurityProfile`. On error, emit a
   warning diagnostic instead of silently removing the resource:
   ```go
   found, err := findProfileByID(ctx, r.client, state.ProfileID.ValueString())
   if err != nil {
       resp.Diagnostics.AddWarning("Failed to read security profile", err.Error())
       return // preserve state, don't remove
   }
   if found == nil {
       resp.State.RemoveResource(ctx)
       return
   }
   ```

5. **Cache List results** — if managing multiple `prisma-airs_security_profile` resources,
   each one calls `findProfileByID` independently. Consider caching the List response within
   a single Terraform refresh cycle (provider-level cache with TTL).

---

## Issue 2: Delete Returns Unparseable JSON

### Problem

`DELETE /v1/mgmt/profile/{id}` returns a response body that fails JSON parsing in the SDK.
The delete likely succeeds server-side, but the SDK returns an error.

Documented in SDK `API_ISSUES.md` #3 (for topics, same pattern applies to profiles).

### Current Workaround (Provider)

```go
_, err := r.client.Profiles.Delete(ctx, state.ProfileID.ValueString())
if err != nil {
    if strings.Contains(err.Error(), "failed to parse response JSON") {
        return // silently ignore
    }
    resp.Diagnostics.AddError(...)
}
```

### Risks

1. **Cannot confirm deletion** — the provider assumes success on parse error, but the delete
   may have actually failed (e.g., 500 error with malformed body).
2. **State drift** — if the delete didn't actually succeed, Terraform state says "deleted" but
   the profile still exists. Next `terraform plan` would not show it (since it was removed from
   state). The orphaned profile stays in the API forever.

### Fix Plan

#### Layer 1: SDK (`prisma-airs-go`)

6. **Handle delete response as raw bytes** — instead of trying to unmarshal the response into
   `DeleteProfileResponse`, check the HTTP status code first:
   ```go
   func (c *ProfilesClient) Delete(ctx context.Context, profileID string) error {
       statusCode, _, err := internal.DoMgmtRequestRaw(ctx, c.svcCfg, internal.MgmtRequestOptions{
           Method: http.MethodDelete,
           Path:   aisec.MgmtProfilePath + "/" + profileID,
       })
       if err != nil {
           return err
       }
       if statusCode >= 200 && statusCode < 300 {
           return nil // success regardless of body
       }
       return fmt.Errorf("delete failed with status %d", statusCode)
   }
   ```
   This requires adding a `DoMgmtRequestRaw` helper that returns `(statusCode int, body []byte, err error)`.

7. **Add `ForceDelete` to provider** — the SDK already has `Profiles.ForceDelete()` but the
   provider doesn't use it. Add a `force_delete` boolean attribute (default false) to the
   resource schema. When true, use `ForceDelete` instead of `Delete` to handle 409 conflicts
   (profile in use by a deployment).

#### Layer 2: Provider (`terraform-provider-prisma-airs`)

8. **Verify deletion** — after delete (whether parse error or not), do a follow-up List check
   to confirm the profile is actually gone:
   ```go
   _, err := r.client.Profiles.Delete(ctx, state.ProfileID.ValueString())
   if err != nil && !strings.Contains(err.Error(), "failed to parse response JSON") {
       resp.Diagnostics.AddError(...)
       return
   }
   // Verify deletion
   stillExists := findProfileByID(ctx, r.client, state.ProfileID.ValueString())
   if stillExists != nil {
       resp.Diagnostics.AddError("Delete failed",
           "Profile still exists after delete. May be in use — try force_delete = true.")
       return
   }
   ```

---

## Issue 3: Create 409 Conflict Handling

### Problem

`POST /v1/mgmt/profile` returns 409 when a profile with the same name already exists.
The provider treats this as success by looking up the existing profile.

### Risks

1. **Adopts unmanaged state** — if someone manually created a profile with the same name,
   Terraform silently adopts it without the user knowing. The existing profile's policy may
   differ from what's in the Terraform config.
2. **No diff on adopt** — the adopted profile's policy is written to state as-is, potentially
   masking a config drift.

### Fix Plan

9. **Warn on 409 adopt** — change the 409 handling from `tflog.Warn` to
   `resp.Diagnostics.AddWarning` so the user sees it in `terraform apply` output:
   ```go
   resp.Diagnostics.AddWarning(
       "Profile already exists",
       fmt.Sprintf("Profile %q already exists (ID: %s). Importing into state. "+
           "Run 'terraform plan' to see if the existing config matches.", name, found.ProfileID),
   )
   ```

10. **Consider failing instead of adopting** — many providers return an error on 409 and
    require `terraform import` for existing resources. This is the safer default. Add a
    provider-level `adopt_existing_resources` bool config (default false) to control behavior.

---

## Issue 4: Import Uses Name-Only Lookup

### Problem

Import uses `findProfileByName` which paginated List + filter. This works but is fragile:
- Profile names are not guaranteed unique by the API (need to verify)
- If two profiles share a name, import picks the first one found

### Fix Plan

11. **Support import by ID** — check if `req.ID` looks like a UUID; if so, use
    `findProfileByID`. Otherwise fall back to `findProfileByName`:
    ```go
    func (r *securityProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
        var found *management.SecurityProfile
        if isUUID(req.ID) {
            found = findProfileByID(ctx, r.client, req.ID)
        } else {
            found = findProfileByName(ctx, r.client, req.ID)
        }
        // ...
    }
    ```

---

## Implementation Order

Priority based on impact and risk:

### Phase 1: Provider Hardening (no SDK changes needed)

| # | Task | Effort | Files |
|---|------|--------|-------|
| 3 | Add timeout to Read context | S | `resource_security_profile.go` |
| 4 | Return error from findProfileByID, don't silently remove | S | `resource_security_profile.go` |
| 8 | Verify deletion with follow-up List check | S | `resource_security_profile.go` |
| 9 | Upgrade 409 handling to user-visible warning | S | `resource_security_profile.go` |
| 11 | Support import by UUID or name | S | `resource_security_profile.go` |

### Phase 2: SDK Improvements

| # | Task | Effort | Files |
|---|------|--------|-------|
| 6 | Raw HTTP response handling for Delete | M | `aisec/internal/mgmthttpclient.go`, `aisec/management/client.go` |
| 2 | Fix GetByName pagination (currently capped at 1000) | S | `aisec/management/client.go` |
| 1 | Test GET `/v1/mgmt/profile/uuid/{id}` — add `Profiles.Get` if supported | M | `aisec/management/client.go` |

### Phase 3: Provider Enhancements (requires Phase 2 SDK)

| # | Task | Effort | Files |
|---|------|--------|-------|
| 5 | Cache List results across resources in a refresh cycle | M | `resource_security_profile.go`, `provider.go` |
| 7 | Add `force_delete` attribute using `ForceDelete` SDK method | S | `resource_security_profile.go` |
| 10 | Add `adopt_existing_resources` provider config | M | `provider.go`, `resource_security_profile.go` |

### Phase 4: Upstream API Requests (Palo Alto Networks)

These are API-side issues only Palo Alto can fix:

- [ ] `GET /v1/mgmt/profile/uuid/{id}` — add single-resource GET endpoint
- [ ] `GET /v1/mgmt/profile?profile_name=<name>` — fix 30s+ timeout
- [ ] `DELETE /v1/mgmt/profile/{id}` — return valid JSON response body
- [ ] `PUT /v1/mgmt/topic/{id}` — fix timeout on topic update

---

## Testing Strategy

Each phase should include:

1. **Unit tests** for all modified functions (`findProfileByID` error return, UUID detection,
   timeout behavior)
2. **Acceptance tests** with `TF_ACC=1` against live API:
   - Create → Read → Update → Read → Delete → verify gone
   - Import by name
   - Import by UUID
   - Delete of non-existent profile (should not error)
   - Create with existing name (409 path)
3. **E2E tests** in temp directory:
   - `terraform plan` / `terraform apply` / `terraform destroy` full lifecycle
   - `terraform import` both by name and UUID

---

## Unresolved Questions

- Does `GET /v1/mgmt/profile/uuid/{id}` actually work? needs live testing
- Are profile names unique? if not, import-by-name is ambiguous
- Should `force_delete` default to true or false?
- Does the API return different status codes for "profile not found" vs "profile in use" on delete?
- Is there an API changelog or roadmap indicating when the GET endpoint / delete response will be fixed?
