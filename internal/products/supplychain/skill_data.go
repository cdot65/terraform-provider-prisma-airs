package supplychain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	ag "github.com/cdot65/prisma-airs-go/aisec/agentguard"
	s "github.com/cdot65/prisma-airs-go/aisec/agentguard/schema"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type skillDataSource struct {
	kind    string
	clients *Clients
}

var _ datasource.DataSourceWithConfigure = &skillDataSource{}

func skillDataSources() []product.DataSource {
	entries := []product.DataSource{{New: NewModelSecurityRulesDataSource, Guide: "data-sources/model-security-rules"}}
	for _, kind := range []string{"instance", "rules", "rule_instances", "overrides", "scan", "scans", "vulnerabilities", "attack_chains", "statistics"} {
		entries = append(entries, product.DataSource{New: func() datasource.DataSource { return &skillDataSource{kind: kind} }, Guide: "data-sources/skill-scanning-" + strings.ReplaceAll(kind, "_", "-")})
	}
	return entries
}
func (d *skillDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supply_chain_skill_scanning_" + d.kind
}
func (d *skillDataSource) attributes() map[string]schema.Attribute {
	a := map[string]schema.Attribute{
		"id":     schema.StringAttribute{Computed: true, Description: "Discovery identity."},
		"result": schema.DynamicAttribute{Computed: true, Sensitive: d.kind == "instance" || d.kind == "scan" || d.kind == "scans" || d.kind == "vulnerabilities" || d.kind == "attack_chains" || d.kind == "overrides", Description: "Native Terraform object containing the SDK response, including nulls and nested lists. Access fields directly (for example result.rules); no JSON decoding is needed. Scans, overrides and findings return one API page. Rules/effective rules use offset traversal until empty (at most 100 pages of 100); concurrent service edits can affect this non-snapshot inventory."},
	}
	optional := func(description string, v ...validator.String) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Description: description, Validators: v}
	}
	required := func(description string, v ...validator.String) schema.StringAttribute {
		return schema.StringAttribute{Required: true, Description: description, Validators: append([]validator.String{stringvalidator.LengthAtLeast(1)}, v...)}
	}
	switch d.kind {
	case "instance":
		a["tenant_id"] = required("Tenant instance ID.")
	case "scan":
		a["scan_uuid"] = optional("Scan UUID. Supply either scan_uuid or fingerprint.", append(uuidValidators(), stringvalidator.ExactlyOneOf(path.MatchRoot("fingerprint")))...)
		a["fingerprint"] = optional("Lowercase SHA-256 fingerprint for scan lookup.", stringvalidator.RegexMatches(fingerprintRegex, "Must be 64 lowercase hexadecimal characters."))
	case "scans", "overrides", "vulnerabilities", "attack_chains":
		upper := int64(100)
		a["limit"] = schema.Int64Attribute{Optional: true, Description: "Page size from 1 to 100; omission uses the server default.", Validators: []validator.Int64{int64validator.Between(1, upper)}}
		a["skip"] = schema.Int64Attribute{Optional: true, Description: "Nonnegative page offset.", Validators: []validator.Int64{int64validator.AtLeast(0)}}
	}
	switch d.kind {
	case "scans":
		a["sort_order"] = optional("Creation-time sort order.", stringvalidator.OneOf("asc", "desc"))
		a["search_query"] = optional("Search by scan name or UUID.", stringvalidator.LengthAtMost(256))
		a["start_time"] = optional("RFC3339 start time.", stringvalidator.RegexMatches(timeRegex, "Use an RFC3339 timestamp."))
		a["end_time"] = optional("RFC3339 end time.", stringvalidator.RegexMatches(timeRegex, "Use an RFC3339 timestamp."))
		a["fingerprint"] = optional("Exact fingerprint filter.", stringvalidator.RegexMatches(fingerprintRegex, "Must be 64 lowercase hexadecimal characters."))
		a["statuses"] = schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Repeated status filters.", Validators: []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf("UPLOADING", "CLASSIFYING", "ANALYZING", "COMPLETED", "FAILED", "ERROR"))}}
		a["artifact_types"] = schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Repeated artifact type filters.", Validators: []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf("AGENT", "SKILL"))}}
	case "overrides":
		for _, k := range []string{"skill_name", "fingerprint", "trusted_by"} {
			a[k] = optional("Filter by "+strings.ReplaceAll(k, "_", " ")+".", stringvalidator.LengthAtMost(256))
			if k == "fingerprint" {
				a[k] = optional("Exact lowercase SHA-256 fingerprint.", stringvalidator.RegexMatches(fingerprintRegex, "Must be 64 lowercase hexadecimal characters."))
			}
		}
		a["q"] = optional("Broad name or trusting-identity search; at least three characters.", stringvalidator.LengthAtLeast(3))
	case "vulnerabilities", "attack_chains":
		a["scan_uuid"] = required("Scan UUID.", uuidValidators()...)
		if d.kind == "vulnerabilities" {
			a["type"] = optional("Vulnerability type filter.", stringvalidator.OneOf("PROMPT_INJECTION", "CODE_EXECUTION", "NETWORK_EXPOSURE", "DATA_EXFILTRATION", "PRIVILEGE_ESCALATION", "AUTHENTICATION_BYPASS", "XSS", "SSRF", "SQL_INJECTION", "PATH_TRAVERSAL", "INFORMATION_DISCLOSURE", "DENIAL_OF_SERVICE", "SUPPLY_CHAIN", "SECRET_EXPOSURE", "MEMORY_POISONING", "SILENT_OPERATION", "BEHAVIOR_MANIPULATION", "UNSAFE_AUTONOMY", "API_PROXY_ATTACK"))
			a["in_chain"] = schema.BoolAttribute{Optional: true, Description: "Filter findings inside/outside attack chains. Explicit false is preserved."}
		} else {
			a["chain_uuid"] = optional("Read one attack-chain detail instead of a list page. Cannot be combined with pagination.", append(uuidValidators(), stringvalidator.ConflictsWith(path.MatchRoot("limit"), path.MatchRoot("skip")))...)
		}
	case "statistics":
		a["time_period"] = optional("Statistics window; omission uses the server's 30_DAYS default.", stringvalidator.OneOf("1_HOUR", "3_HOURS", "24_HOURS", "7_DAYS", "30_DAYS", "ALL_TIME"))
	}
	return a
}
func (d *skillDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Reads Skill Scanning " + strings.ReplaceAll(d.kind, "_", " ") + " without starting a scan or changing tenant settings.", Attributes: d.attributes()}
}
func (d *skillDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, diags := getSkillClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.clients = c
}
func (d *skillDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := config.Attributes()
	result, err := d.fetch(ctx, v)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Skill Scanning "+d.kind, skillError(err))
		return
	}
	native, err := skillNative(ctx, result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map Skill Scanning response", skillError(err))
		return
	}
	v["id"] = types.StringValue(d.clients.TenantID + "/" + d.kind)
	v["result"] = types.DynamicValue(native)
	t := map[string]attr.Type{}
	for k, a := range d.attributes() {
		t[k] = a.GetType()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, types.ObjectValueMust(t, v))...)
}
func skillListOpts(v map[string]attr.Value) ag.ListOpts {
	o := ag.ListOpts{}
	if n, ok := v["limit"].(types.Int64); ok {
		o.Limit = int(n.ValueInt64())
	}
	if n, ok := v["skip"].(types.Int64); ok {
		o.Skip = int(n.ValueInt64())
	}
	return o
}
func (d *skillDataSource) fetch(ctx context.Context, v map[string]attr.Value) (any, error) {
	c := d.clients.Skills
	o := skillListOpts(v)
	switch d.kind {
	case "instance":
		instance, e := c.Instances.Get(ctx, skillString(v, "tenant_id"))
		if e != nil {
			return nil, e
		}
		if instance.TSGID != d.clients.TenantID || instance.TenantID != skillString(v, "tenant_id") {
			return nil, fmt.Errorf("instance response does not match the configured tenant")
		}
		return instance, nil
	case "rules":
		rows, e := allRules(ctx, c)
		return map[string]any{"rules": rows}, e
	case "rule_instances":
		rows, e := allRuleInstances(ctx, c)
		return map[string]any{"rule_instances": rows}, e
	case "overrides":
		return c.SkillOverrides.List(ctx, ag.SkillOverrideListOpts{ListOpts: o, SkillName: skillString(v, "skill_name"), Fingerprint: skillString(v, "fingerprint"), TrustedBy: skillString(v, "trusted_by"), Q: skillString(v, "q")})
	case "scan":
		if id := skillString(v, "scan_uuid"); id != "" {
			return c.Scans.Get(ctx, id)
		}
		return c.Scans.Lookup(ctx, skillString(v, "fingerprint"))
	case "scans":
		filter := ag.ScanFilter{SortOrder: s.SortDirection(skillString(v, "sort_order")), SearchQuery: skillString(v, "search_query"), StartTime: skillString(v, "start_time"), EndTime: skillString(v, "end_time"), Fingerprint: skillString(v, "fingerprint")}
		if err := validateTimeRange(filter.StartTime, filter.EndTime); err != nil {
			return nil, err
		}
		statuses, artifacts := []string{}, []string{}
		if ds := v["statuses"].(types.List).ElementsAs(ctx, &statuses, false); ds.HasError() {
			return nil, fmt.Errorf("invalid statuses")
		}
		if ds := v["artifact_types"].(types.List).ElementsAs(ctx, &artifacts, false); ds.HasError() {
			return nil, fmt.Errorf("invalid artifact_types")
		}
		for _, value := range statuses {
			filter.Statuses = append(filter.Statuses, s.AgentGuardScanStatus(value))
		}
		for _, value := range artifacts {
			filter.ArtifactTypes = append(filter.ArtifactTypes, s.ArtifactType(value))
		}
		return c.Scans.List(ctx, ag.ScanListOpts{ListOpts: o, ScanFilter: filter})
	case "vulnerabilities":
		opts := ag.VulnerabilityListOpts{ListOpts: o, Type: s.VulnerabilityType(skillString(v, "type"))}
		if value := v["in_chain"].(types.Bool); !value.IsNull() && !value.IsUnknown() {
			b := value.ValueBool()
			opts.InChain = &b
		}
		return c.Scans.ListVulnerabilities(ctx, skillString(v, "scan_uuid"), opts)
	case "attack_chains":
		if chain := skillString(v, "chain_uuid"); chain != "" {
			return c.Scans.GetAttackChain(ctx, skillString(v, "scan_uuid"), chain)
		}
		return c.Scans.ListAttackChains(ctx, skillString(v, "scan_uuid"), o)
	case "statistics":
		period := s.TimePeriod(skillString(v, "time_period"))
		scans, e := c.Statistics.Scans(ctx, period)
		if e != nil {
			return nil, e
		}
		rules, e := c.Statistics.Rules(ctx, period)
		return map[string]any{"scans": scans, "rules": rules}, e
	}
	return nil, fmt.Errorf("unsupported Skill Scanning data source")
}

// Read-only responses are native objects/tuples. JSON is used only internally
// to respect SDK wire tags and omitted/null intent, never as Terraform input.
func skillNative(ctx context.Context, value any) (attr.Value, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var wire any
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err = decoder.Decode(&wire); err != nil {
		return nil, err
	}
	return skillValue(ctx, wire)
}
func skillValue(ctx context.Context, v any) (attr.Value, error) {
	switch x := v.(type) {
	case nil:
		return types.StringNull(), nil
	case string:
		return types.StringValue(x), nil
	case bool:
		return types.BoolValue(x), nil
	case json.Number:
		n, _, e := big.ParseFloat(string(x), 10, 512, big.ToNearestEven)
		if e != nil {
			return nil, e
		}
		return types.NumberValue(n), nil
	case map[string]any:
		ts, vs := map[string]attr.Type{}, map[string]attr.Value{}
		for k, item := range x {
			a, e := skillValue(ctx, item)
			if e != nil {
				return nil, e
			}
			ts[k] = a.Type(ctx)
			vs[k] = a
		}
		o, diags := types.ObjectValue(ts, vs)
		if diags.HasError() {
			return nil, fmt.Errorf("invalid Skill Scanning object")
		}
		return o, nil
	case []any:
		ts, vs := make([]attr.Type, len(x)), make([]attr.Value, len(x))
		for i, item := range x {
			a, e := skillValue(ctx, item)
			if e != nil {
				return nil, e
			}
			ts[i] = a.Type(ctx)
			vs[i] = a
		}
		o, diags := types.TupleValue(ts, vs)
		if diags.HasError() {
			return nil, fmt.Errorf("invalid Skill Scanning collection")
		}
		return o, nil
	}
	return nil, fmt.Errorf("unsupported Skill Scanning value")
}
