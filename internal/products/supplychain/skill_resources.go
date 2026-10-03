package supplychain

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec"
	ag "github.com/cdot65/prisma-airs-go/aisec/agentguard"
	s "github.com/cdot65/prisma-airs-go/aisec/agentguard/schema"
	"github.com/cdot65/prisma-airs-provider/internal/tfutil"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API owns catalog rules and one tenant policy. Terraform owns individual
// rule settings, recording a restoration state rather than inventing DELETE.
type skillResource struct {
	kind    string
	clients *Clients
}

var _ resource.ResourceWithConfigure = &skillResource{}
var _ resource.ResourceWithImportState = &skillResource{}

func NewSkillScanningInstanceResource() resource.Resource { return &skillResource{kind: "instance"} }
func NewSkillScanningRuleResource() resource.Resource     { return &skillResource{kind: "rule"} }
func NewSkillScanningOverrideResource() resource.Resource { return &skillResource{kind: "override"} }
func (r *skillResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supply_chain_skill_scanning_" + r.kind
}
func requiredSkillString(description string, immutable bool, validators ...validator.String) schema.StringAttribute {
	a := schema.StringAttribute{Description: description, Required: true, Validators: append([]validator.String{stringvalidator.LengthAtLeast(1)}, validators...)}
	if immutable {
		a.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return a
}
func optionalSkillString(description string, immutable, sensitive bool, validators ...validator.String) schema.StringAttribute {
	a := schema.StringAttribute{Description: description, Optional: true, Sensitive: sensitive, Validators: validators}
	if immutable {
		a.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
	}
	return a
}
func computedSkillString(description string, stable bool) schema.StringAttribute {
	a := schema.StringAttribute{Description: description, Computed: true}
	if stable {
		a.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	}
	return a
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func uuidValidators() []validator.String {
	return []validator.String{stringvalidator.RegexMatches(uuidPattern, "Must be a lowercase UUID.")}
}
func (r *skillResource) attributes() map[string]schema.Attribute {
	a := map[string]schema.Attribute{"id": computedSkillString("Terraform identity.", true), "tsg_id": computedSkillString("Tenant Service Group ID.", true)}
	switch r.kind {
	case "instance":
		a["tenant_id"] = requiredSkillString("Skill Scanning tenant ID. Changing it replaces the instance.", true)
		a["support_account_id"] = requiredSkillString("Support account ID.", false)
		a["created_by"] = requiredSkillString("Creator identity required by the service.", false)
		a["support_account_name"] = optionalSkillString("Support account name; removing a configured value sends null on update.", false, false)
		a["auth_code"] = schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true, Description: "Deployment authorization code, never stored in plan or state. Requires Terraform 1.11+. Increment auth_code_version to change or clear it.", Validators: []validator.String{stringvalidator.AlsoRequires(path.MatchRoot("auth_code_version"))}}
		a["auth_code_version"] = schema.Int64Attribute{Optional: true, Description: "Nonsecret trigger for authorization-code changes. Increment to send the configured code, or explicit null when clearing it.", Validators: []validator.Int64{int64validator.AtLeast(1)}}
		a["registration_details"] = schema.DynamicAttribute{Required: true, Sensitive: true, Description: "Complete desired native registration object for SDK provisioning metadata (region, license_name, entitlements, tsg_instances, extra, and extensions). PUT owns this inventory and can deactivate removed profiles. Identity/authentication fields belong to their dedicated attributes. Never supply a JSON string. This desired object is retained because GET cannot recover the complete registration request."}
		a["iam_controlled"] = schema.BoolAttribute{Description: "Request-only IAM setting, persisted in state. Explicit false is sent; removal sends null on update. GET cannot detect drift.", Optional: true}
	case "rule":
		a["rule_uuid"] = requiredSkillString("Catalog rule UUID, not rule-instance UUID. Manage each tenant/rule pair once.", true, uuidValidators()...)
		a["state"] = requiredSkillString("Desired rule state.", false, stringvalidator.OneOf("DISABLED", "ALLOWING", "BLOCKING"))
		a["original_state"] = computedSkillString("Effective state captured at adoption. Destroy restores this state. Import captures the current state.", true)
		a["rule_instance_uuid"] = computedSkillString("Server policy rule-instance UUID; may change when the service recreates it.", false)
		a["name"] = computedSkillString("Catalog rule name.", true)
	case "override":
		a["skill_name"] = requiredSkillString("Trusted skill name. Override changes require replacement; the API has no update operation.", true, stringvalidator.LengthAtMost(256))
		a["fingerprint"] = requiredSkillString("Exact skill fingerprint.", true, stringvalidator.RegexMatches(fingerprintRegex, "Must be 64 lowercase hexadecimal characters."))
		a["trusted_by"] = requiredSkillString("Identity trusting the skill.", true, stringvalidator.LengthAtMost(256))
		a["reason"] = optionalSkillString("Reason for trusting the skill.", true, false, stringvalidator.LengthAtMost(1024))
		a["original_scan_uuid"] = optionalSkillString("Original scan UUID.", true, false, uuidValidators()...)
		a["decision"] = computedSkillString("Trusted-skill decision (ALLOW).", true)
		a["created_at"] = computedSkillString("Creation timestamp.", true)
		a["updated_at"] = computedSkillString("Last modification timestamp.", false)
	}
	return a
}
func (r *skillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	descriptions := map[string]string{"instance": "Manages a Skill Scanning tenant instance. Existing instances require import; destroy deletes the tenant instance, not individual scans.", "rule": "Manages one Skill Scanning policy rule. Create/update sets its state; destroy restores its captured original state because the API has no rule delete.", "override": "Manages a trusted Skill Scanning fingerprint override. Edits replace the override; destroy removes trust."}
	resp.Schema = schema.Schema{Description: descriptions[r.kind], Attributes: r.attributes()}
}
func (r *skillResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, d := getSkillClient(req.ProviderData)
	resp.Diagnostics.Append(d...)
	r.clients = c
}
func skillString(v map[string]attr.Value, k string) string {
	if x, ok := v[k].(types.String); ok {
		return x.ValueString()
	}
	return ""
}
func skillOptional(v map[string]attr.Value, k string, clearing bool) aisec.Optional[string] {
	x, ok := v[k].(types.String)
	if !ok || x.IsUnknown() {
		return aisec.Optional[string]{}
	}
	if x.IsNull() {
		if clearing {
			return aisec.Null[string]()
		}
		return aisec.Optional[string]{}
	}
	return aisec.Value(x.ValueString())
}
func (r *skillResource) object(ctx context.Context, v map[string]attr.Value) types.Object {
	t := map[string]attr.Type{}
	for k, a := range r.attributes() {
		t[k] = a.GetType()
	}
	return types.ObjectValueMust(t, v)
}
func (r *skillResource) checkpoint(ctx context.Context, v map[string]attr.Value) types.Object {
	if r.kind == "instance" {
		v["auth_code"] = types.StringNull()
	}
	for k, a := range r.attributes() {
		if a.IsComputed() && (v[k] == nil || v[k].IsUnknown()) {
			v[k] = types.StringNull()
		}
	}
	return r.object(ctx, v)
}
func (r *skillResource) instanceRequest(v, prior map[string]attr.Value) (s.InstanceCreateModel, error) {
	q, e := skillRegistration(v["registration_details"])
	if e != nil {
		return q, e
	}

	clear := func(k string) bool { return prior != nil && prior[k] != nil && !prior[k].IsNull() }
	q.TenantID = skillString(v, "tenant_id")
	q.TSGID = r.clients.TenantID
	q.SupportAccountID = skillString(v, "support_account_id")
	q.CreatedBy = skillString(v, "created_by")
	q.SupportAccountName = skillOptional(v, "support_account_name", clear("support_account_name"))
	q.AuthCode = aisec.Optional[string]{}
	x := v["iam_controlled"].(types.Bool)
	if !x.IsNull() && !x.IsUnknown() {
		q.IamControlled = aisec.Value(x.ValueBool())
	} else if clear("iam_controlled") {
		q.IamControlled = aisec.Null[bool]()
	}
	return q, nil
}
func checkedReceipt(receipt *s.InstanceResponseModel, tenant, tsg string) error {
	if receipt == nil || !receipt.IsSuccess || receipt.TenantID != tenant {
		return fmt.Errorf("skill scanning instance mutation did not return a successful receipt for the requested tenant")
	}
	if id, ok := receipt.TSGID.Get(); ok && id != tsg {
		return fmt.Errorf("skill scanning receipt belongs to another TSG")
	}
	return nil
}
func (r *skillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan types.Object
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := plan.Attributes()
	if err := r.validPlan(v); err != nil {
		resp.Diagnostics.AddError("Invalid Skill Scanning plan", skillError(err))
		return
	}
	v["tsg_id"] = types.StringValue(r.clients.TenantID)
	switch r.kind {
	case "instance":
		tenant := skillString(v, "tenant_id")
		_, err := r.clients.Skills.Instances.Get(ctx, tenant)
		if err == nil {
			resp.Diagnostics.AddError("Existing Skill Scanning instance", "Import the existing tenant instance before managing it.")
			return
		}
		if !tfutil.IsNotFound(err) {
			resp.Diagnostics.AddError("Failed to check Skill Scanning instance", skillError(err))
			return
		}
		body, e := r.instanceRequest(v, nil)
		if e != nil {
			resp.Diagnostics.AddError("Invalid native registration", skillError(e))
			return
		}
		body.AuthCode, err = instanceAuthCode(ctx, req.Config, false)
		if err != nil {
			resp.Diagnostics.AddError("Invalid instance authorization code", skillError(err))
			return
		}
		receipt, err := r.clients.Skills.Instances.Create(ctx, body)
		knownSuccess := err == nil || mutationMayHaveApplied(err)
		if err == nil {
			err = checkedReceipt(receipt, tenant, r.clients.TenantID)
		}
		if err != nil {
			if knownSuccess {
				if receipt != nil {
					if tsg, ok := receipt.TSGID.Get(); ok && tsg != r.clients.TenantID {
						v["tsg_id"] = types.StringValue(tsg)
					}
				}
				v["id"] = types.StringValue(tenant)
				resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
			}
			resp.Diagnostics.AddError("Failed to create Skill Scanning instance", skillError(err)+" If the service applied the request, inspect and import the requested tenant ID before retrying.")
			return
		}
		v["id"] = types.StringValue(tenant)
	case "rule":
		rule, err := r.findRule(ctx, skillString(v, "rule_uuid"))
		if err != nil {
			resp.Diagnostics.AddError("Failed to find Skill Scanning rule", skillError(err))
			return
		}
		v["original_state"] = types.StringValue(string(rule.State))
		v["id"] = v["rule_uuid"]
		if err = r.writeRule(ctx, v); err != nil {
			if mutationMayHaveApplied(err) {
				resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
			}
			resp.Diagnostics.AddError("Failed to set Skill Scanning rule", skillError(err))
			return
		}
	case "override":
		before, err := allOverrides(ctx, r.clients.Skills)
		if err != nil {
			resp.Diagnostics.AddError("Failed to check existing trusted skills", skillError(err))
			return
		}
		for _, existing := range before {
			if existing.Fingerprint == skillString(v, "fingerprint") {
				resp.Diagnostics.AddError("Existing trusted skill override", "Import the existing override UUID before managing this fingerprint.")
				return
			}
		}
		q := s.SkillOverrideCreateRequest{SkillName: skillString(v, "skill_name"), Fingerprint: skillString(v, "fingerprint"), TrustedBy: skillString(v, "trusted_by"), Decision: s.OverrideDecisionAllow, Reason: skillOptional(v, "reason", false), OriginalScanUUID: skillOptional(v, "original_scan_uuid", false)}
		override, err := r.clients.Skills.SkillOverrides.Create(ctx, q)
		if err != nil || override == nil || !aisec.IsValidUUID(override.UUID) || override.TSGID != r.clients.TenantID {
			if override != nil && aisec.IsValidUUID(override.UUID) {
				v["id"] = types.StringValue(override.UUID)
				if override.TSGID != "" {
					v["tsg_id"] = types.StringValue(override.TSGID)
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
			} else if err == nil || mutationMayHaveApplied(err) {
				recovered, recoveryErr := r.recoverOverride(ctx, v)
				if recoveryErr == nil {
					r.mapOverride(v, recovered)
					resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
				}
			}
			message := "The service did not return an override UUID in the configured TSG."
			if err != nil {
				message = skillError(err)
			}
			resp.Diagnostics.AddError("Trusted skill creation could not be confirmed", message+" Inspect the fingerprint through the overrides data source and import its UUID before retrying. Any recovered identity is retained in state.")
			return
		}
		r.mapOverride(v, override)
	}
	// A successful mutation must retain a usable identity if the follow-up GET
	// fails; otherwise Terraform cannot recover or destroy the remote object.
	resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
	if err := r.refresh(ctx, v); err != nil {
		resp.Diagnostics.AddError("Skill Scanning change saved; refresh failed", "Terraform retained the identity for recovery. "+skillError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.object(ctx, v))...)
}
func (r *skillResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := state.Attributes()
	if err := r.refresh(ctx, v); err != nil {
		if tfutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Skill Scanning "+r.kind, skillError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.object(ctx, v))...)
}
func (r *skillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state types.Object
	resp.State = req.State
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := plan.Attributes()
	prior := state.Attributes()
	if err := r.validPlan(v); err != nil {
		resp.Diagnostics.AddError("Invalid Skill Scanning plan", skillError(err))
		return
	}
	if err := r.checkTenant(prior); err != nil {
		resp.Diagnostics.AddError("Cannot update Skill Scanning resource", skillError(err))
		return
	}
	var err error
	switch r.kind {
	case "instance":
		body, e := r.instanceRequest(v, prior)
		if e != nil {
			resp.Diagnostics.AddError("Invalid native registration", skillError(e))
			return
		}
		if !v["auth_code_version"].Equal(prior["auth_code_version"]) {
			var e error
			body.AuthCode, e = instanceAuthCode(ctx, req.Config, true)
			if e != nil {
				resp.Diagnostics.AddError("Invalid instance authorization code", skillError(e))
				return
			}
		}
		receipt, e := r.clients.Skills.Instances.Update(ctx, skillString(prior, "id"), body)
		err = e
		if err == nil {
			err = checkedReceipt(receipt, skillString(prior, "id"), r.clients.TenantID)
		}
	case "rule":
		err = r.writeRule(ctx, v)
	default:
		resp.Diagnostics.AddError("Override update is unsupported", "Trusted skill edits require replacement.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Skill Scanning "+r.kind, skillError(err))
		return
	}
	v["id"] = prior["id"]
	v["tsg_id"] = prior["tsg_id"]
	if r.kind == "rule" {
		v["original_state"] = prior["original_state"]
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.checkpoint(ctx, v))...)
	if err = r.refresh(ctx, v); err != nil {
		resp.Diagnostics.AddError("Skill Scanning change saved; refresh failed", skillError(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.object(ctx, v))...)
}
func (r *skillResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state types.Object
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := state.Attributes()
	if err := r.checkTenant(v); err != nil {
		resp.Diagnostics.AddError("Cannot destroy Skill Scanning resource", skillError(err))
		return
	}
	switch r.kind {
	case "rule":
		original := skillString(v, "original_state")
		if original != "DISABLED" && original != "ALLOWING" && original != "BLOCKING" {
			resp.Diagnostics.AddError("Missing rule restoration state", "The original policy state is required for safe destroy.")
			return
		}
		v["state"] = types.StringValue(original)
		if err := r.writeRule(ctx, v); err != nil {
			resp.Diagnostics.AddError("Failed to restore Skill Scanning rule", skillError(err))
			return
		}
		rule, err := r.findRule(ctx, skillString(v, "rule_uuid"))
		if err != nil || string(rule.State) != original {
			resp.Diagnostics.AddError("Rule restoration not confirmed", "The effective policy did not confirm the captured original state.")
			return
		}
	case "instance":
		receipt, err := r.clients.Skills.Instances.Delete(ctx, skillString(v, "id"))
		if err == nil {
			err = checkedReceipt(receipt, skillString(v, "id"), r.clients.TenantID)
		}
		tfutil.FinishDelete(ctx, skillSafeError(err), "Skill Scanning instance", func(ctx context.Context) (bool, error) {
			_, e := r.clients.Skills.Instances.Get(ctx, skillString(v, "id"))
			if tfutil.IsNotFound(e) {
				return true, nil
			}
			return false, skillSafeError(e)
		}, &resp.Diagnostics)
	case "override":
		err := r.clients.Skills.SkillOverrides.Delete(ctx, skillString(v, "id"))
		tfutil.FinishDelete(ctx, skillSafeError(err), "trusted skill override", func(ctx context.Context) (bool, error) {
			_, e := findOverride(ctx, r.clients.Skills, skillString(v, "id"), skillString(v, "fingerprint"))
			if tfutil.IsNotFound(e) {
				return true, nil
			}
			return false, skillSafeError(e)
		}, &resp.Diagnostics)
	}
}
func (r *skillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.kind != "instance" && !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Use a lowercase catalog rule UUID or override UUID.")
		return
	}
	v := map[string]attr.Value{}
	for k, a := range r.attributes() {
		v[k] = types.StringNull()
		if a.GetType().Equal(types.BoolType) {
			v[k] = types.BoolNull()
		} else if a.GetType().Equal(types.Int64Type) {
			v[k] = types.Int64Null()
		} else if a.GetType().Equal(types.DynamicType) {
			v[k] = types.DynamicNull()
		}
	}
	v["id"] = types.StringValue(req.ID)
	v["tsg_id"] = types.StringValue(r.clients.TenantID)
	if r.kind == "instance" {
		v["tenant_id"] = types.StringValue(req.ID)
	}
	if r.kind == "rule" {
		v["rule_uuid"] = types.StringValue(req.ID)
	}
	if err := r.refresh(ctx, v); err != nil {
		resp.Diagnostics.AddError("Failed to import Skill Scanning "+r.kind, skillError(err))
		return
	}
	if r.kind == "rule" {
		v["original_state"] = v["state"]
	}
	if r.kind == "instance" {
		resp.Diagnostics.AddWarning("Write-only instance settings are unavailable", "The complete registration_details request is not returned by GET; the first apply after import always issues a full PUT. Verify and supply the full desired provisioning payload before that first apply. auth_code is never stored in state and cannot be recovered on import. Configure it with auth_code_version when needed. iam_controlled is a retained request-only setting and cannot be recovered from GET.")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, r.object(ctx, v))...)
}
func (r *skillResource) checkTenant(v map[string]attr.Value) error {
	if t := skillString(v, "tsg_id"); t == "" || t != r.clients.TenantID {
		return fmt.Errorf("resource belongs to another TSG; use its original provider configuration")
	}
	return nil
}
func (r *skillResource) refresh(ctx context.Context, v map[string]attr.Value) error {
	if err := r.checkTenant(v); err != nil {
		return err
	}
	switch r.kind {
	case "instance":
		instance, err := r.clients.Skills.Instances.Get(ctx, skillString(v, "id"))
		if err != nil {
			return err
		}
		if instance.TenantID != skillString(v, "id") || instance.TSGID != r.clients.TenantID {
			return fmt.Errorf("instance identity does not match the configured tenant")
		}
		v["tenant_id"] = types.StringValue(instance.TenantID)
		// GET includes audit/registration metadata, which may be omitted or
		// normalized. Retain desired registration inputs; import reads observed values.
		if v["created_by"].IsNull() {
			v["support_account_id"] = types.StringValue(instance.SupportAccountID)
			v["created_by"] = stringPointer(instance.CreatedBy)
			v["support_account_name"] = stringPointer(instance.SupportAccountName)
		}
		v["auth_code"] = types.StringNull()
	case "rule":
		rule, err := r.findRule(ctx, skillString(v, "rule_uuid"))
		if err != nil {
			return err
		}
		if rule.TSGID != r.clients.TenantID {
			return fmt.Errorf("rule instance belongs to another TSG")
		}
		v["state"] = types.StringValue(string(rule.State))
		v["rule_instance_uuid"] = types.StringNull()
		if rule.UUID != "" {
			v["rule_instance_uuid"] = types.StringValue(rule.UUID)
		}
		v["name"] = types.StringValue(rule.Rule.Name)
	case "override":
		override, err := findOverride(ctx, r.clients.Skills, skillString(v, "id"), skillString(v, "fingerprint"))
		if err != nil {
			return err
		}
		if override.TSGID != r.clients.TenantID {
			return fmt.Errorf("override belongs to another TSG")
		}
		r.mapOverride(v, override)
	}
	v["tsg_id"] = types.StringValue(r.clients.TenantID)
	return nil
}
func (r *skillResource) mapOverride(v map[string]attr.Value, o *s.SkillOverrideResponse) {
	v["id"] = types.StringValue(o.UUID)
	v["tsg_id"] = types.StringValue(o.TSGID)
	if strings.TrimSpace(skillString(v, "skill_name")) != o.SkillName || v["skill_name"] == nil || v["skill_name"].IsNull() {
		v["skill_name"] = types.StringValue(o.SkillName)
	}
	v["fingerprint"] = types.StringValue(o.Fingerprint)
	if o.TrustedBy != nil || v["trusted_by"] == nil || v["trusted_by"].IsNull() {
		v["trusted_by"] = stringPointer(o.TrustedBy)
	}
	// The API canonicalizes an empty optional reason to null. Preserve the
	// configured empty spelling so a successful create remains plan-consistent.
	if o.Reason != nil || skillString(v, "reason") != "" || v["reason"] == nil || v["reason"].IsNull() {
		v["reason"] = stringPointer(o.Reason)
	}
	if o.OriginalScanUUID == nil || !strings.EqualFold(skillString(v, "original_scan_uuid"), *o.OriginalScanUUID) {
		v["original_scan_uuid"] = stringPointer(o.OriginalScanUUID)
	}
	v["decision"] = types.StringValue(string(o.Decision))
	v["created_at"] = types.StringValue(o.CreatedAt)
	v["updated_at"] = types.StringValue(o.UpdatedAt)
}
func stringPointer(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}
func (r *skillResource) findRule(ctx context.Context, id string) (*s.SkillSecurityRuleInstanceResponse, error) {
	rules, err := allRuleInstances(ctx, r.clients.Skills)
	if err != nil {
		return nil, err
	}
	for _, rule := range rules {
		if rule.RuleUUID == id {
			return &rule, nil
		}
	}
	// A rule without an instance uses its catalog default. A missing catalog entry
	// cannot be silently recreated or interpreted as a disabled policy.
	catalog, err := allRules(ctx, r.clients.Skills)
	if err != nil {
		return nil, err
	}
	for _, rule := range catalog {
		if rule.UUID == id {
			return &s.SkillSecurityRuleInstanceResponse{RuleUUID: id, State: rule.DefaultState, Rule: rule, TSGID: r.clients.TenantID}, nil
		}
	}
	return nil, aisec.NewHTTPError("Skill Scanning catalog rule not found", aisec.ClientSideError, 404)
}
func (r *skillResource) writeRule(ctx context.Context, v map[string]attr.Value) error {
	if err := r.checkTenant(v); err != nil {
		return err
	}
	id, state := skillString(v, "rule_uuid"), skillString(v, "state")
	_, err := r.clients.Skills.RuleInstances.Update(ctx, s.SkillSecurityRuleInstancesUpdateRequest{RuleConfigurations: map[string]s.SkillSecurityRuleConfiguration{id: {State: s.RuleState(state)}}})
	return err
}
func allRules(ctx context.Context, c *ag.Client) ([]s.SkillSecurityRuleResponse, error) {
	return skillPages(ctx, func(skip int) ([]s.SkillSecurityRuleResponse, error) {
		p, e := c.Rules.List(ctx, ag.ListOpts{Limit: 100, Skip: skip})
		if e != nil {
			return nil, e
		}
		return p.Rules, nil
	}, func(v s.SkillSecurityRuleResponse) string { return v.UUID })
}
func allRuleInstances(ctx context.Context, c *ag.Client) ([]s.SkillSecurityRuleInstanceResponse, error) {
	return skillPages(ctx, func(skip int) ([]s.SkillSecurityRuleInstanceResponse, error) {
		p, e := c.RuleInstances.List(ctx, ag.ListOpts{Limit: 100, Skip: skip})
		if e != nil {
			return nil, e
		}
		return p.RuleInstances, nil
	}, func(v s.SkillSecurityRuleInstanceResponse) string { return v.UUID })
}
func allOverrides(ctx context.Context, c *ag.Client) ([]s.SkillOverrideResponse, error) {
	return skillPages(ctx, func(skip int) ([]s.SkillOverrideResponse, error) {
		p, e := c.SkillOverrides.List(ctx, ag.SkillOverrideListOpts{ListOpts: ag.ListOpts{Limit: 100, Skip: skip}})
		if e != nil {
			return nil, e
		}
		return p.SkillOverrides, nil
	}, func(v s.SkillOverrideResponse) string { return v.UUID })
}
func skillPages[T any](ctx context.Context, fetch func(int) ([]T, error), id func(T) string) ([]T, error) {
	result := []T{}
	seen := map[string]bool{}
	// A final empty page confirms completion even if the server clamps page size.
	// Reported totals are not used: the preview rule catalog reports page length.
	for page := 0; page < 100; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		items, err := fetch(len(result))
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return result, nil
		}
		for _, item := range items {
			key := id(item)
			if key == "" || seen[key] {
				return nil, fmt.Errorf("skill scanning pagination returned a missing or repeated UUID; inventory is incomplete")
			}
			seen[key] = true
		}
		result = append(result, items...)
	}
	return nil, fmt.Errorf("skill scanning pagination exceeded 100 pages; inventory is incomplete")
}
func findOverride(ctx context.Context, c *ag.Client, id string, fingerprints ...string) (*s.SkillOverrideResponse, error) {
	fetch := func() ([]s.SkillOverrideResponse, error) {
		if len(fingerprints) > 0 && fingerprints[0] != "" {
			return skillPages(ctx, func(skip int) ([]s.SkillOverrideResponse, error) {
				p, e := c.SkillOverrides.List(ctx, ag.SkillOverrideListOpts{ListOpts: ag.ListOpts{Limit: 100, Skip: skip}, Fingerprint: fingerprints[0]})
				if e != nil {
					return nil, e
				}
				return p.SkillOverrides, nil
			}, func(v s.SkillOverrideResponse) string { return v.UUID })
		}
		return allOverrides(ctx, c)
	}
	var identities map[string]bool
	for pass := 0; pass < 2; pass++ {
		items, e := fetch()
		if e != nil {
			return nil, e
		}
		current := map[string]bool{}
		for _, item := range items {
			if item.UUID == id {
				return &item, nil
			}
			current[item.UUID] = true
		}
		if pass == 1 {
			if len(current) != len(identities) {
				return nil, fmt.Errorf("override inventory changed while confirming absence; retry refresh")
			}
			for key := range current {
				if !identities[key] {
					return nil, fmt.Errorf("override inventory changed while confirming absence; retry refresh")
				}
			}
		}
		identities = current
	}
	return nil, aisec.NewHTTPError("trusted skill override not found", aisec.ClientSideError, 404)
}

func (r *skillResource) validPlan(v map[string]attr.Value) error {
	for k, a := range r.attributes() {
		if !a.IsComputed() && (v[k] == nil || v[k].IsUnknown() || (a.IsRequired() && v[k].IsNull())) {
			return fmt.Errorf("%s must be known before applying", k)
		}
	}
	if r.kind == "instance" {
		_, e := skillRegistration(v["registration_details"])
		return e
	}
	return nil
}

func instanceAuthCode(ctx context.Context, config tfsdk.Config, clear bool) (aisec.Optional[string], error) {
	var code types.String
	if d := config.GetAttribute(ctx, path.Root("auth_code"), &code); d.HasError() {
		return aisec.Optional[string]{}, fmt.Errorf("cannot read authorization code from configuration")
	}
	if code.IsUnknown() {
		return aisec.Optional[string]{}, fmt.Errorf("authorization code must be known before applying")
	}
	if code.IsNull() {
		if clear {
			return aisec.Null[string](), nil
		}
		return aisec.Optional[string]{}, nil
	}
	return aisec.Value(code.ValueString()), nil
}
func (r *skillResource) recoverOverride(ctx context.Context, v map[string]attr.Value) (*s.SkillOverrideResponse, error) {
	rows, e := allOverrides(ctx, r.clients.Skills)
	if e != nil {
		return nil, e
	}
	var found *s.SkillOverrideResponse
	for _, row := range rows {
		if row.Fingerprint == skillString(v, "fingerprint") && row.SkillName == skillString(v, "skill_name") && row.TSGID == r.clients.TenantID && row.TrustedBy != nil && *row.TrustedBy == skillString(v, "trusted_by") && aisec.IsValidUUID(row.UUID) {
			if found != nil {
				return nil, fmt.Errorf("multiple matching overrides; import explicitly")
			}
			candidate := row
			found = &candidate
		}
	}
	if found == nil {
		return nil, fmt.Errorf("cannot recover override identity")
	}
	return found, nil
}

func skillRegistration(value attr.Value) (s.InstanceCreateModel, error) {
	details, e := skillInput(value)
	if e != nil {
		return s.InstanceCreateModel{}, e
	}
	object, ok := details.(map[string]any)
	if !ok {
		return s.InstanceCreateModel{}, fmt.Errorf("registration_details must be a native object")
	}
	// Required explicit ownership prevents an imported instance from being PUT
	// with a guessed/incomplete registration copied from its summary-only GET.
	for _, key := range []string{"tsg_id", "tenant_id", "created_by", "support_account_id", "support_account_name", "iam_controlled", "auth_code"} {
		if _, exists := object[key]; exists {
			return s.InstanceCreateModel{}, fmt.Errorf("registration_details cannot override %s", key)
		}
	}
	encoded, e := json.Marshal(object)
	if e != nil {
		return s.InstanceCreateModel{}, fmt.Errorf("registration_details must contain finite JSON-compatible native values")
	}
	var q s.InstanceCreateModel
	if e = json.Unmarshal(encoded, &q); e != nil {
		return q, fmt.Errorf("registration_details does not match the SDK registration shape")
	}

	// SDK decoding uses float64 for open containers. Keep the original native
	// values (json.Number) in those fields so registration PUTs are lossless.
	if items, ok := object["entitlements"].([]any); ok {
		q.Entitlements = aisec.Value(items)
	}
	if fields, ok := object["extra"].(map[string]any); ok {
		q.Extra = &fields
	}
	if items, ok := object["tsg_instances"].([]any); ok {
		rows := make([]map[string]any, len(items))
		for i, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				return q, fmt.Errorf("tsg_instances entries must be native objects")
			}
			rows[i] = row
		}
		q.TSGInstances = aisec.Value(rows)
	}
	return q, nil
}
