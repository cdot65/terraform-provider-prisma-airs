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
	ID               types.String `tfsdk:"id"`
	CustomerAppID    types.String `tfsdk:"customer_app_id"`
	AppName          types.String `tfsdk:"app_name"`
	TsgID            types.String `tfsdk:"tsg_id"`
	ModelName        types.String `tfsdk:"model_name"`
	CloudProvider    types.String `tfsdk:"cloud_provider"`
	Environment      types.String `tfsdk:"environment"`
	Status           types.String `tfsdk:"status"`
	CreatedBy        types.String `tfsdk:"created_by"`
	UpdatedBy        types.String `tfsdk:"updated_by"`
	AiAgentFramework types.String `tfsdk:"ai_agent_framework"`
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
				Description: "Terraform resource ID (same as customer_app_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_app_id": schema.StringAttribute{
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
			"tsg_id": schema.StringAttribute{
				Optional:    true,
				Description: "Tenant service group ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_name": schema.StringAttribute{
				Optional:    true,
				Description: "Model name associated with the app.",
			},
			"cloud_provider": schema.StringAttribute{
				Optional:    true,
				Description: "Cloud provider for the app.",
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Description: "Deployment environment.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "App status.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity of the user who created the app.",
			},
			"updated_by": schema.StringAttribute{
				Optional:    true,
				Description: "Identity of the user updating the app.",
			},
			"ai_agent_framework": schema.StringAttribute{
				Computed:    true,
				Description: "AI agent framework.",
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
		AppName:       plan.AppName.ValueString(),
		TsgID:         plan.TsgID.ValueString(),
		CloudProvider: plan.CloudProvider.ValueString(),
		Environment:   plan.Environment.ValueString(),
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

	// SDK Get uses app_name as the lookup parameter.
	app, err := r.client.CustomerApps.Get(ctx, state.AppName.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read customer app", err.Error())
		return
	}

	// Preserve write-only values from state.
	updatedByVal := state.UpdatedBy
	mapAppToState(app, &state)
	state.UpdatedBy = updatedByVal
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
		AppName:       plan.AppName.ValueString(),
		ModelName:     plan.ModelName.ValueString(),
		CloudProvider: plan.CloudProvider.ValueString(),
		Environment:   plan.Environment.ValueString(),
	}

	// SDK Update uses customer_app_id as the identifier.
	app, err := r.client.CustomerApps.Update(ctx, state.CustomerAppID.ValueString(), updateReq)
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

	updatedBy := state.UpdatedBy.ValueString()
	if updatedBy == "" {
		updatedBy = "terraform"
	}

	// SDK Delete uses app_name and updated_by.
	_, err := r.client.CustomerApps.Delete(ctx, state.AppName.ValueString(), updatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "failed to parse response JSON") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete customer app", err.Error())
		return
	}
}

func (r *customerAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by app_name since SDK Get uses app_name.
	appName := req.ID

	app, err := r.client.CustomerApps.Get(ctx, appName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import customer app", err.Error())
		return
	}

	var state CustomerAppResourceModel
	mapAppToState(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapAppToState(app *management.CustomerApp, state *CustomerAppResourceModel) {
	state.ID = types.StringValue(app.CustomerAppID)
	state.CustomerAppID = types.StringValue(app.CustomerAppID)
	state.AppName = types.StringValue(app.AppName)
	state.TsgID = types.StringValue(app.TsgID)
	state.ModelName = types.StringValue(app.ModelName)
	state.CloudProvider = types.StringValue(app.CloudProvider)
	state.Environment = types.StringValue(app.Environment)
	state.Status = types.StringValue(app.Status)
	state.CreatedBy = types.StringValue(app.CreatedBy)
	state.AiAgentFramework = types.StringValue(app.AiAgentFramework)
}
