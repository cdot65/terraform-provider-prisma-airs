package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cdot65/prisma-airs-provider/internal/product"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type catalogDataSource struct{ client *client }

var _ datasource.DataSourceWithConfigure = &catalogDataSource{}

func (d *catalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_ai_providers"
}

func (d *catalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the upstream AI provider catalog without managing integrations. The catalog contains provider-family UUIDs for integration creation, not existing integration or workspace-provider UUIDs. No credentials are returned.",
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{Computed: true, Description: "Provider definitions returned by the catalog, including inactive definitions.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id":     schema.StringAttribute{Computed: true, Description: "Provider-family UUID accepted as integration ai_provider_id."},
				"slug":   schema.StringAttribute{Computed: true, Description: "Exact catalog slug; OpenAI uses open-ai and Anthropic uses anthropic."},
				"name":   schema.StringAttribute{Computed: true, Description: "Catalog display name."},
				"status": schema.StringAttribute{Computed: true, Description: "Catalog lifecycle status."},
			}}},
			"ids_by_slug": schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Active provider-family UUIDs keyed by exact catalog slug. Duplicate or empty identities cause an error rather than an ambiguous lookup."},
		},
	}
}

func (d *catalogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ds := product.Client[*client](req.ProviderData, "gateway", "AI Gateway")
	resp.Diagnostics.Append(ds...)
	d.client = c
}

func (d *catalogDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	reply, err := d.client.sdk.Integrations.Catalog(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Gateway AI provider catalog", gatewayError(err))
		return
	}
	if reply == nil || (reply.Success != nil && !*reply.Success) {
		resp.Diagnostics.AddError("Invalid Gateway AI provider catalog", "The service did not return a successful catalog response.")
		return
	}
	itemTypes := map[string]attr.Type{"id": types.StringType, "slug": types.StringType, "name": types.StringType, "status": types.StringType}
	items := make([]attr.Value, 0, len(reply.Data))
	ids, seenIDs, seenSlugs := map[string]string{}, map[string]bool{}, map[string]bool{}
	for _, row := range reply.Data {
		if strings.TrimSpace(row.ID) == "" || strings.TrimSpace(row.Slug) == "" || seenIDs[row.ID] || seenSlugs[row.Slug] {
			resp.Diagnostics.AddError("Invalid Gateway AI provider catalog", fmt.Sprintf("Catalog entry %q has an empty or duplicate identity. No lookup map was published.", row.Slug))
			return
		}
		seenIDs[row.ID], seenSlugs[row.Slug] = true, true
		items = append(items, types.ObjectValueMust(itemTypes, map[string]attr.Value{
			"id": types.StringValue(row.ID), "slug": types.StringValue(row.Slug),
			"name": types.StringValue(row.Name), "status": types.StringValue(row.Status),
		}))
		if row.Status == "active" {
			ids[row.Slug] = row.ID
		}
	}
	list, ds := types.ListValue(types.ObjectType{AttrTypes: itemTypes}, items)
	resp.Diagnostics.Append(ds...)
	lookup, ds := types.MapValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
		Items     types.List `tfsdk:"items"`
		IDsBySlug types.Map  `tfsdk:"ids_by_slug"`
	}{list, lookup})...)
}
