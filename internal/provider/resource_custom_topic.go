package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
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
	client *airsruntime.Client
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
				Computed:    true,
				Description: "Description of the custom topic. Omission adopts the server description; explicit empty strings are unsupported.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
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

	createReq := airsruntime.CreateTopicRequest{
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

	// There is no single-topic GET; page through the list and filter by ID.
	found, err := findTopicByID(ctx, r.client, state.TopicID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read custom topic", err.Error())
		return
	}
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

	name, description := plan.TopicName.ValueString(), plan.Description.ValueString()
	examples := []string{}
	updateReq := airsruntime.UpdateTopicFieldsRequest{TopicName: &name, Examples: &examples}
	if !plan.Description.IsUnknown() && !plan.Description.IsNull() {
		updateReq.Description = &description
	}

	if !plan.Examples.IsNull() && !plan.Examples.IsUnknown() {
		resp.Diagnostics.Append(plan.Examples.ElementsAs(ctx, &examples, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	topic, err := r.client.Topics.UpdateFields(ctx, state.TopicID.ValueString(), updateReq)
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
	finishDelete(ctx, err, "custom topic", func(ctx context.Context) (bool, error) {
		found, lookupErr := findTopicByID(ctx, r.client, state.TopicID.ValueString())
		return found == nil, lookupErr
	}, &resp.Diagnostics)
}

func (r *customTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	topicID := req.ID

	found, err := findTopicByID(ctx, r.client, topicID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import custom topic", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("Topic not found", "No topic with ID: "+topicID)
		return
	}

	var state CustomTopicResourceModel
	mapTopicToState(ctx, found, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// findTopicByID searches for a topic by ID using paginated list calls. A nil
// topic with a nil error means the topic does not exist; list failures are
// returned so a transient error is never mistaken for a deleted topic.
func findTopicByID(ctx context.Context, client *airsruntime.Client, topicID string) (*airsruntime.CustomTopic, error) {
	offset := 0
	limit := 100
	seen := map[string]bool{}
	for {
		listResp, err := client.Topics.List(ctx, airsruntime.ListOpts{Limit: limit, Offset: offset})
		if err != nil {
			return nil, err
		}
		for i := range listResp.Items {
			if listResp.Items[i].TopicID == topicID {
				return &listResp.Items[i], nil
			}
		}
		if len(listResp.Items) == 0 && listResp.NextOffset > offset {
			return nil, fmt.Errorf("topic list returned an empty page with an advancing cursor")
		}
		for _, item := range listResp.Items {
			if item.TopicID == "" || seen[item.TopicID] {
				return nil, fmt.Errorf("topic list repeated a page or omitted an identity")
			}
			seen[item.TopicID] = true
		}
		if listResp.NextOffset > offset {
			offset = listResp.NextOffset
		} else if len(listResp.Items) >= limit {
			offset += len(listResp.Items)
		} else {
			return nil, nil
		}
	}
}

func mapTopicToState(ctx context.Context, topic *airsruntime.CustomTopic, state *CustomTopicResourceModel, diags *diag.Diagnostics) {
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
	} else if state.Examples.IsNull() || state.Examples.IsUnknown() {
		state.Examples = types.ListNull(types.StringType)
	} else {
		state.Examples, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}
}
