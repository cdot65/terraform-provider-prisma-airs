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
	AgentApp         types.Bool   `tfsdk:"agent_app"`
	AiAgentFramework types.String `tfsdk:"ai_agent_framework"`
	AiSecProfileName types.String `tfsdk:"ai_sec_profile_name"`
}

func (r *customerAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer_app"
}

func (r *customerAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an existing customer application in Prisma AIRS. Customer apps are created externally (via the AIRS console or when apps register); use `terraform import` to bring them under management.",
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
				Computed:    true,
				Description: "Tenant service group ID.",
			},
			"model_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Model name associated with the app.",
			},
			"cloud_provider": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cloud provider for the app.",
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
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
			"agent_app": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is an agent application.",
			},
			"ai_agent_framework": schema.StringAttribute{
				Computed:    true,
				Description: "AI agent framework.",
			},
			"ai_sec_profile_name": schema.StringAttribute{
				Computed:    true,
				Description: "Associated AI security profile name.",
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

func (r *customerAppResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Create not supported",
		"Customer apps cannot be created via the API. Use `terraform import prisma-airs_customer_app.<name> <app_name>` to manage an existing app.",
	)
}

func (r *customerAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomerAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CustomerApps.Get(ctx, state.AppName.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read customer app", err.Error())
		return
	}

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
	state.AgentApp = types.BoolValue(app.AgentApp)
	state.AiAgentFramework = types.StringValue(app.AiAgentFramework)
	state.AiSecProfileName = types.StringValue(app.AiSecProfileName)
}
