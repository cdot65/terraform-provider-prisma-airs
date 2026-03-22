package provider

import (
	"context"

	"github.com/cdot65/prisma-airs-go/aisec/management"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &customTopicResource{}
	_ resource.ResourceWithImportState = &customTopicResource{}
)

func NewCustomTopicResource() resource.Resource {
	return &customTopicResource{}
}

type customTopicResource struct {
	client *management.Client
}

type CustomTopicResourceModel struct {
	ID          types.String `tfsdk:"id"`
	TopicID     types.String `tfsdk:"topic_id"`
	TopicName   types.String `tfsdk:"topic_name"`
	Description types.String `tfsdk:"description"`
	Examples    types.List   `tfsdk:"examples"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *customTopicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_topic"
}

func (r *customTopicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom detection topic.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as topic_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"topic_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the topic.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"topic_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the custom topic.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the custom topic.",
			},
			"examples": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Example strings for topic detection.",
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

func (r *customTopicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := management.CreateTopicRequest{
		TopicName:   plan.TopicName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Examples.IsNull() && !plan.Examples.IsUnknown() {
		var examples []string
		resp.Diagnostics.Append(plan.Examples.ElementsAs(ctx, &examples, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Examples = examples
	}

	topic, err := r.client.Topics.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create custom topic", err.Error())
		return
	}

	mapTopicToState(ctx, topic, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK Topics.List has an offset bug (offset=0 is omitted but required by API).
	// Work around by paginating with offset=1 and searching all pages.
	found := findTopicByID(ctx, r.client, state.TopicID.ValueString())
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapTopicToState(ctx, found, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CustomTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := management.UpdateTopicRequest{
		TopicName:   plan.TopicName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Examples.IsNull() && !plan.Examples.IsUnknown() {
		var examples []string
		resp.Diagnostics.Append(plan.Examples.ElementsAs(ctx, &examples, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Examples = examples
	}

	topic, err := r.client.Topics.Update(ctx, state.TopicID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update custom topic", err.Error())
		return
	}

	mapTopicToState(ctx, topic, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Topics.ForceDelete(ctx, state.TopicID.ValueString(), "terraform")
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete custom topic", err.Error())
		return
	}
}

func (r *customTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	topicID := req.ID

	found := findTopicByID(ctx, r.client, topicID)
	if found == nil {
		resp.Diagnostics.AddError("Topic not found", "No topic with ID: "+topicID)
		return
	}

	var state CustomTopicResourceModel
	mapTopicToState(ctx, found, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findTopicByID searches for a topic by ID using paginated list calls.
func findTopicByID(ctx context.Context, client *management.Client, topicID string) *management.CustomTopic {
	offset := 0
	limit := 100
	for {
		listResp, err := client.Topics.List(ctx, management.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil
		}
		for i := range listResp.Items {
			if listResp.Items[i].TopicID == topicID {
				return &listResp.Items[i]
			}
		}
		if len(listResp.Items) < limit {
			return nil
		}
		offset += limit
	}
}

func mapTopicToState(ctx context.Context, topic *management.CustomTopic, state *CustomTopicResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(topic.TopicID)
	state.TopicID = types.StringValue(topic.TopicID)
	state.TopicName = types.StringValue(topic.TopicName)
	state.Description = types.StringValue(topic.Description)
	state.CreatedAt = types.StringValue(topic.CreatedTs)
	state.UpdatedAt = types.StringValue(topic.LastModifiedTs)

	if len(topic.Examples) > 0 {
		examplesList, d := types.ListValueFrom(ctx, types.StringType, topic.Examples)
		diags.Append(d...)
		state.Examples = examplesList
	} else {
		state.Examples = types.ListNull(types.StringType)
	}
}
