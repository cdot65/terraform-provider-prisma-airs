package gateway

import (
	"context"
	"encoding/json"
	"time"

	gw "github.com/cdot65/prisma-airs-go/aisec/gateway"
	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceDataSource struct {
	client *client
	list   bool
}

var _ datasource.DataSourceWithConfigure = &workspaceDataSource{}

func workspaceSummaryTypes() map[string]attr.Type {
	ts := map[string]attr.Type{}
	for _, k := range []string{"id", "name", "description", "slug", "scope_name", "status", "created_at", "last_updated_at"} {
		ts[k] = types.StringType
	}
	ts["is_default"] = types.BoolType
	return ts
}
func workspaceSummaryAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{}
	for k, t := range workspaceSummaryTypes() {
		if t.Equal(types.BoolType) {
			attrs[k] = schema.BoolAttribute{Computed: true, Description: "Whether this is the tenant default workspace."}
		} else {
			attrs[k] = schema.StringAttribute{Computed: true, Description: "Remote workspace " + k + " when available."}
		}
	}
	return attrs
}
func (d *workspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	name := "workspace"
	if d.list {
		name = "workspaces"
	}
	resp.TypeName = req.ProviderTypeName + "_gateway_" + name
}
func (d *workspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := workspaceSummaryAttributes()
	description := "Reads safe workspace metadata from the tenant admin plane. Does not expose defaults, credentials, users, or security settings."
	if d.list {
		attrs = map[string]schema.Attribute{"status": schema.StringAttribute{Optional: true, Description: "Lifecycle filter; defaults to active.", Validators: []validator.String{stringvalidator.OneOf("active", "archived")}}, "items": schema.ListNestedAttribute{Computed: true, Description: "Safe metadata for the returned workspace page.", NestedObject: schema.NestedAttributeObject{Attributes: workspaceSummaryAttributes()}}, "total_count": schema.Int64Attribute{Computed: true, Description: "Reported inventory total."}, "has_more": schema.BoolAttribute{Computed: true, Description: "Whether more records are reported; null if the response omits this flag."}, "complete": schema.BoolAttribute{Computed: true, Description: "True only when the response explicitly establishes a complete inventory."}}
		description = "Lists safe admin-plane workspace metadata. The captured route has no paging parameters; incomplete results cannot prove absence."
	} else {
		attrs["workspace_id"] = schema.StringAttribute{Required: true, Description: "Workspace UUID to read.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
	}
	resp.Schema = schema.Schema{Description: description, Attributes: attrs}
}
func (d *workspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, x := product.Client[*client](req.ProviderData, "gateway", "AI Gateway")
	resp.Diagnostics.Append(x...)
	d.client = c
}
func workspaceSummary(doc document) types.Object {
	ts := workspaceSummaryTypes()
	v := map[string]attr.Value{}
	for k, t := range ts {
		if t.Equal(types.BoolType) {
			v[k] = types.BoolNull()
			switch x := doc[k].(type) {
			case bool:
				v[k] = types.BoolValue(x)
			case json.Number:
				if x == "0" || x == "1" {
					v[k] = types.BoolValue(x == "1")
				}
			}
		} else {
			v[k] = types.StringNull()
			if x, ok := doc[k].(string); ok {
				v[k] = types.StringValue(x)
			}
		}
	}
	return types.ObjectValueMust(ts, v)
}
func (d *workspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var m types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	values := m.Attributes()
	if !d.list {
		id := values["workspace_id"].(types.String).ValueString()
		remote, e := readSDK(d.client.sdk.Workspaces.Get(ctx, id, gw.WorkspaceGetOptions{Plane: gw.WorkspaceAdmin}))
		if e != nil {
			resp.Diagnostics.AddError("Workspace lookup failed", gatewayError(e))
			return
		}
		// Only detail responses decorate names; list rows already contain the literal label.
		remote["name"] = workspaceLabel(remote)
		for k, v := range workspaceSummary(remote).Attributes() {
			values[k] = v
		}
	} else {
		status := values["status"].(types.String).ValueString()
		if status == "" {
			status = "active"
		}
		r := &workspaceResource{client: d.client}
		rows, total, complete, hasMore, e := r.inventory(ctx, status)
		if e != nil {
			resp.Diagnostics.AddError("Workspace listing failed", gatewayError(e))
			return
		}
		items := make([]attr.Value, len(rows))
		for i, row := range rows {
			items[i] = workspaceSummary(row)
		}
		list, x := types.ListValue(types.ObjectType{AttrTypes: workspaceSummaryTypes()}, items)
		resp.Diagnostics.Append(x...)
		values["items"] = list
		values["total_count"] = types.Int64Value(total)
		values["complete"] = types.BoolValue(complete)
		values["has_more"] = types.BoolNull()
		if hasMore != nil {
			values["has_more"] = types.BoolValue(*hasMore)
		}
		if !complete {
			resp.Diagnostics.AddWarning("Workspace inventory is incomplete", "The service did not establish a complete inventory and this route has no captured pagination parameters. These records cannot prove absence or unique ownership.")
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, types.ObjectValueMust(m.AttributeTypes(ctx), values))...)
}
