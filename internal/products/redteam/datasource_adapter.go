package redteam

import (
	"context"
	"time"

	"github.com/cdot65/prisma-airs-go/aisec"
	"github.com/cdot65/prisma-airs-go/aisec/redteam"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type adapterDataSource struct {
	client *redteam.Client
	list   bool
}

var _ datasource.DataSourceWithConfigure = &adapterDataSource{}

func NewAdapterDataSource() datasource.DataSource  { return &adapterDataSource{} }
func NewAdaptersDataSource() datasource.DataSource { return &adapterDataSource{list: true} }
func (d *adapterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_red_team_adapter"
	if d.list {
		resp.TypeName += "s"
	}
}
func (d *adapterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	if d.list {
		resp.Schema = schema.Schema{Description: "Discovers all Red Team adapters with bounded pagination. Reading never executes scripts or takes ownership.", Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{Computed: true, Description: "Adapter identity and lifecycle metadata.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true, Description: "Adapter UUID."}, "name": schema.StringAttribute{Computed: true, Description: "Adapter name."}, "status": schema.StringAttribute{Computed: true, Description: "DRAFT or ACTIVE."},
			}}},
			"ids_by_name": schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Adapter UUIDs keyed by exact name. Duplicate names fail rather than silently selecting an adapter."},
		}}
		return
	}
	resp.Schema = schema.Schema{Description: "Reads an existing Red Team adapter without managing or executing it. Secret values remain null.", Attributes: map[string]schema.Attribute{
		"id":                          schema.StringAttribute{Required: true, Description: "Adapter UUID, typically selected from red_team_adapters.ids_by_name."},
		"name":                        schema.StringAttribute{Computed: true, Description: "Adapter name."},
		"description":                 schema.StringAttribute{Computed: true, Description: "Adapter description."},
		"script":                      schema.StringAttribute{Computed: true, Sensitive: true, Description: "Decoded Python script; never executed by this data source."},
		"network_broker_channel_uuid": schema.StringAttribute{Computed: true, Description: "Existing channel reference, if present."},
		"status":                      schema.StringAttribute{Computed: true, Description: "DRAFT or ACTIVE."},
		"variables": schema.MapNestedAttribute{Computed: true, Sensitive: true, Description: "Variable key/type inventory; redacted secrets have null values.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{Computed: true, Description: "VAR or SECRET."}, "value": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Visible value or null if redacted."},
		}}},
	}}
}
func (d *adapterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		c, ds := getRedTeamClient(req.ProviderData)
		resp.Diagnostics.Append(ds...)
		d.client = c
	}
}
func (d *adapterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if d.list {
		rows, err := d.client.Adapters.ListAll(ctx, redteam.AdapterListOpts{}, aisec.CollectOptions{})
		if err != nil {
			resp.Diagnostics.AddError("Failed to discover adapters", safeTargetError(err))
			return
		}
		kinds := map[string]attr.Type{"id": types.StringType, "name": types.StringType, "status": types.StringType}
		items := []attr.Value{}
		ids := map[string]string{}
		seen := map[string]bool{}
		for _, row := range rows {
			if row.UUID == "" || row.Name == "" || seen[row.UUID] {
				resp.Diagnostics.AddError("Invalid adapter identity", "The service returned empty or duplicate adapter identities.")
				return
			}
			if _, duplicate := ids[row.Name]; duplicate {
				resp.Diagnostics.AddError("Ambiguous adapter name", "Multiple adapters have the same name. Select a UUID from the API inventory instead.")
				return
			}
			seen[row.UUID] = true
			ids[row.Name] = row.UUID
			items = append(items, types.ObjectValueMust(kinds, map[string]attr.Value{"id": types.StringValue(row.UUID), "name": types.StringValue(row.Name), "status": types.StringValue(string(row.Status))}))
		}
		list, ds := types.ListValue(types.ObjectType{AttrTypes: kinds}, items)
		resp.Diagnostics.Append(ds...)
		lookup, ds := types.MapValueFrom(ctx, types.StringType, ids)
		resp.Diagnostics.Append(ds...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
			Items types.List `tfsdk:"items"`
			IDs   types.Map  `tfsdk:"ids_by_name"`
		}{list, lookup})...)
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Adapters.Get(ctx, id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read adapter", safeTargetError(err))
		return
	}
	var state adapterModel
	mapAdapter(ctx, remote, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		ID                       types.String `tfsdk:"id"`
		Name                     types.String `tfsdk:"name"`
		Description              types.String `tfsdk:"description"`
		Script                   types.String `tfsdk:"script"`
		NetworkBrokerChannelUUID types.String `tfsdk:"network_broker_channel_uuid"`
		Status                   types.String `tfsdk:"status"`
		Variables                types.Map    `tfsdk:"variables"`
	}{state.ID, state.Name, state.Description, state.Script, state.NetworkBrokerChannelUUID, state.Status, state.Variables})...)
}
