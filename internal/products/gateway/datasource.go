package gateway

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gatewayDataSource struct {
	definition definition
	client     *client
}

var _ datasource.DataSourceWithConfigure = &gatewayDataSource{}
var summaryKeys = []string{"id", "name", "slug", "status", "workspace_id", "organisation_id", "integration_id", "mcp_integration_id", "ai_provider_id", "user_id", "type", "created_at", "last_updated_at", "version_id"}

func summaryTypes() map[string]attr.Type {
	ts := map[string]attr.Type{}
	for _, k := range summaryKeys {
		ts[k] = types.StringType
	}
	ts["is_default"] = types.BoolType
	ts["enabled"] = types.BoolType
	return ts
}
func plural(name string) string {
	switch name {
	case "usage_limit":
		return "usage_limits"
	case "rate_limit":
		return "rate_limits"
	}
	return name + "s"
}
func (d *gatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_" + plural(d.definition.name)
}
func (d *gatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	items := map[string]schema.Attribute{}
	for _, k := range summaryKeys {
		items[k] = schema.StringAttribute{Computed: true, Description: "Remote " + k + " when available; otherwise null."}
	}
	for _, k := range []string{"is_default", "enabled"} {
		items[k] = schema.BoolAttribute{Computed: true, Description: "Remote " + k + " when available; otherwise null."}
	}
	attrs := map[string]schema.Attribute{"items": schema.ListNestedAttribute{Computed: true, Description: "Safe resource metadata for the returned page. Credentials and config documents are never included.", NestedObject: schema.NestedAttributeObject{Attributes: items}}, "total_count": schema.Int64Attribute{Computed: true, Description: "Server-reported total where available; otherwise the number of returned items."}}
	if d.definition.workspace {
		attrs["workspace_id"] = schema.StringAttribute{Required: true, Description: "Existing Gateway workspace UUID.", Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
	}
	if d.definition.pagination {
		attrs["page_size"] = schema.Int64Attribute{Optional: true, Description: "Results per page (default 100).", Validators: []validator.Int64{int64validator.Between(1, 100)}}
		attrs["current_page"] = schema.Int64Attribute{Optional: true, Description: "One-based page index (default 1).", Validators: []validator.Int64{int64validator.AtLeast(1)}}
	}
	resp.Schema = schema.Schema{Description: "Lists Gateway " + plural(d.definition.name) + " metadata. This is a read-only page, not an exhaustive inventory; archived records may be returned. No secrets are stored.", Attributes: attrs}
}
func (d *gatewayDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ds := product.Client[*client](req.ProviderData, "gateway", "AI Gateway")
	resp.Diagnostics.Append(ds...)
	d.client = c
}
func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var model types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	values := model.Attributes()
	w := ""
	if x, ok := values["workspace_id"].(types.String); ok {
		w = x.ValueString()
	}
	size, page := int64(100), int64(1)
	if x, ok := values["page_size"].(types.Int64); ok && !x.IsNull() && !x.IsUnknown() {
		size = x.ValueInt64()
	}
	if x, ok := values["current_page"].(types.Int64); ok && !x.IsNull() && !x.IsUnknown() {
		page = x.ValueInt64()
	}
	items, total, e := d.definition.list(ctx, d.client, w, size, page)
	if e != nil {
		resp.Diagnostics.AddError("Failed to list Gateway "+plural(d.definition.name), gatewayError(e))
		return
	}
	native := make([]attr.Value, len(items))
	ts := summaryTypes()
	for i, item := range items {
		attrs := map[string]attr.Value{}
		for _, k := range summaryKeys {
			attrs[k] = types.StringNull()
			if x, ok := item[k].(string); ok {
				attrs[k] = types.StringValue(x)
			}
		}
		for _, k := range []string{"is_default", "enabled"} {
			attrs[k] = types.BoolNull()
			switch x := item[k].(type) {
			case bool:
				attrs[k] = types.BoolValue(x)
			case json.Number:
				if x == "0" || x == "1" {
					attrs[k] = types.BoolValue(x == "1")
				}
			}
		}
		native[i] = types.ObjectValueMust(ts, attrs)
	}
	list, ds := types.ListValue(types.ObjectType{AttrTypes: ts}, native)
	resp.Diagnostics.Append(ds...)
	values["items"] = list
	values["total_count"] = types.Int64Value(total)
	model = types.ObjectValueMust(model.AttributeTypes(ctx), values)
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
