package provider

import (
	"context"
	"strings"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &customerAppResource{}
	_ resource.ResourceWithImportState = &customerAppResource{}
)

func NewCustomerAppResource() resource.Resource {
	return &customerAppResource{}
}

type customerAppResource struct {
	client *management.Client
}

type CustomerAppResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AppID       types.String `tfsdk:"app_id"`
	AppName     types.String `tfsdk:"app_name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *customerAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer_app"
}

func (r *customerAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a customer application registration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as app_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the customer app.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the customer application.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the customer application.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
	}
}

func (r *customerAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := getMgmtClient(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *customerAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomerAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := management.CreateAppRequest{
		AppName:     plan.AppName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	app, err := r.client.CustomerApps.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create customer app", err.Error())
		return
	}

	mapAppToState(app, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customerAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomerAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CustomerApps.Get(ctx, state.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read customer app", err.Error())
		return
	}

	mapAppToState(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customerAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomerAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CustomerAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := management.UpdateAppRequest{
		AppName:     plan.AppName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	app, err := r.client.CustomerApps.Update(ctx, state.AppID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update customer app", err.Error())
		return
	}

	mapAppToState(app, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customerAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomerAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CustomerApps.Delete(ctx, state.AppID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete customer app", err.Error())
		return
	}
}

func (r *customerAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	appID := req.ID

	app, err := r.client.CustomerApps.Get(ctx, appID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import customer app", err.Error())
		return
	}

	var state CustomerAppResourceModel
	mapAppToState(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapAppToState(app *management.CustomerApp, state *CustomerAppResourceModel) {
	state.ID = types.StringValue(app.AppID)
	state.AppID = types.StringValue(app.AppID)
	state.AppName = types.StringValue(app.AppName)
	state.Description = types.StringValue(app.Description)
	state.CreatedAt = types.StringValue(app.CreatedAt)
	state.UpdatedAt = types.StringValue(app.UpdatedAt)
}
