package runtime

import (
	"context"

	"github.com/cdot65/prisma-airs-provider/internal/tfutil"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                   = &securityProfileResource{}
	_ resource.ResourceWithImportState    = &securityProfileResource{}
	_ resource.ResourceWithValidateConfig = &securityProfileResource{}
)

func NewSecurityProfileResource() resource.Resource {
	return &securityProfileResource{}
}

type securityProfileResource struct {
	client *airsruntime.Client
}

// ── Model types ──────────────────────────────────────────────────────

type SecurityProfileResourceModel struct {
	DLPTenantID        types.String             `tfsdk:"dlp_tenant_id"`
	ID                 types.String             `tfsdk:"id"`
	ProfileID          types.String             `tfsdk:"profile_id"`
	Revision           types.Int64              `tfsdk:"revision"`
	ProfileName        types.String             `tfsdk:"profile_name"`
	Active             types.Bool               `tfsdk:"active"`
	CreatedAt          types.String             `tfsdk:"created_at"`
	UpdatedAt          types.String             `tfsdk:"updated_at"`
	AiSecurityProfiles []AiSecurityProfileModel `tfsdk:"ai_security_profile"`
	DlpDataProfiles    []DlpDataProfileModel    `tfsdk:"dlp_data_profile"`
}

type AiSecurityProfileModel struct {
	ContentTypeMode                  types.String                    `tfsdk:"content_type_mode"`
	ContentTypeConfigurations        *ContentTypeConfigurationsModel `tfsdk:"content_type_configurations"`
	EnableFullConversationInspection types.Bool                      `tfsdk:"enable_full_conversation_inspection"`
	ModelType                        types.String                    `tfsdk:"model_type"`
	ContentType                      types.String                    `tfsdk:"content_type"`
	MaskDataInStorage                types.Bool                      `tfsdk:"mask_data_in_storage"`
	Latency                          *LatencyModel                   `tfsdk:"latency"`
	DataProtection                   *DataProtectionModel            `tfsdk:"data_protection"`
	AppProtection                    *AppProtectionModel             `tfsdk:"app_protection"`
	ModelProtection                  []ModelProtectionModel          `tfsdk:"model_protection"`
	AgentProtection                  []AgentProtectionModel          `tfsdk:"agent_protection"`
}

type LatencyModel struct {
	InlineTimeoutAction types.String `tfsdk:"inline_timeout_action"`
	MaxInlineLatency    types.Int64  `tfsdk:"max_inline_latency"`
}

type DataProtectionModel struct {
	SourceCodeDetection *SourceCodeDetectionModel `tfsdk:"source_code_detection"`
	DataLeakDetection   *DataLeakDetectionModel   `tfsdk:"data_leak_detection"`
	DatabaseSecurity    []DatabaseSecurityModel   `tfsdk:"database_security"`
}

type DatabaseSecurityModel struct {
	Severity types.String `tfsdk:"severity"`
	Name     types.String `tfsdk:"name"`
	Action   types.String `tfsdk:"action"`
}

type DataLeakDetectionModel struct {
	Action         types.String          `tfsdk:"action"`
	MaskDataInline types.Bool            `tfsdk:"mask_data_inline"`
	Members        []DataLeakMemberModel `tfsdk:"member"`
}

type DataLeakMemberModel struct {
	Text    types.String `tfsdk:"text"`
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

type AppProtectionModel struct {
	UrlDetectedSeverity     types.String                  `tfsdk:"url_detected_severity"`
	AlertURLCategory        types.List                    `tfsdk:"alert_url_category"`
	BlockURLCategory        types.List                    `tfsdk:"block_url_category"`
	AllowURLCategory        types.List                    `tfsdk:"allow_url_category"`
	DefaultURLCategory      types.List                    `tfsdk:"default_url_category"`
	UrlDetectedAction       types.String                  `tfsdk:"url_detected_action"`
	MaliciousCodeProtection *MaliciousCodeProtectionModel `tfsdk:"malicious_code_protection"`
}

type MaliciousCodeProtectionModel struct {
	Severity types.String `tfsdk:"severity"`
	Name     types.String `tfsdk:"name"`
	Action   types.String `tfsdk:"action"`
}

type ModelProtectionModel struct {
	SeverityByConfidence *SeverityByConfidenceModel `tfsdk:"severity_by_confidence"`
	Severity             types.String               `tfsdk:"severity"`
	Name                 types.String               `tfsdk:"name"`
	Action               types.String               `tfsdk:"action"`
	ToxicCategories      []ToxicCategoryModel       `tfsdk:"toxic_category"`
	TopicLists           []TopicListModel           `tfsdk:"topic_list"`
}

type ToxicCategoryModel struct {
	SeverityByConfidence *SeverityByConfidenceModel `tfsdk:"severity_by_confidence"`
	Category             types.String               `tfsdk:"category"`
	Action               types.String               `tfsdk:"action"`
}

type TopicListModel struct {
	Action types.String    `tfsdk:"action"`
	Topics []TopicRefModel `tfsdk:"topic"`
}

type TopicRefModel struct {
	Severity  types.String `tfsdk:"severity"`
	TopicName types.String `tfsdk:"topic_name"`
	TopicID   types.String `tfsdk:"topic_id"`
	Revision  types.Int64  `tfsdk:"revision"`
}

type AgentProtectionModel struct {
	Severity types.String `tfsdk:"severity"`
	Name     types.String `tfsdk:"name"`
	Action   types.String `tfsdk:"action"`
}

type DlpDataProfileModel struct {
	Name         types.String `tfsdk:"name"`
	UUID         types.String `tfsdk:"uuid"`
	ProfileID    types.String `tfsdk:"profile_id"`
	Version      types.String `tfsdk:"version"`
	LogSeverity  types.String `tfsdk:"log_severity"`
	NonFileBased types.String `tfsdk:"non_file_based"`
	FileBased    types.String `tfsdk:"file_based"`
}

// ── Schema ───────────────────────────────────────────────────────────

func (r *securityProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runtime_security_profile"
}

func (r *securityProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AI Runtime Security profile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform resource ID (same as profile_id).",
			},
			"profile_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the security profile.",
			},
			"revision": schema.Int64Attribute{
				Computed:    true,
				Description: "Highest numeric revision of the managed profile name; changes on policy updates.",
			},
			"profile_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the security profile.",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the profile is active.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of revision 1, when that revision remains available; null if original creation time is unavailable.",
				PlanModifiers: []planmodifier.String{
					sameNamedStringState("profile_name"),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
		Blocks: map[string]schema.Block{
			"ai_security_profile": schema.ListNestedBlock{
				Description: "AI security profile configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"model_type": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							Default:       stringdefault.StaticString("default"),
							Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
							Description:   "Model type (e.g., 'default').",
						},
						"content_type": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							Description:   "Content type.",
						},
						"mask_data_in_storage": schema.BoolAttribute{
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
							Description:   "Whether to mask data in storage.",
						},
					},
					Blocks: map[string]schema.Block{
						"latency": schema.SingleNestedBlock{
							Description: "Latency configuration for inline scanning.",
							Attributes: map[string]schema.Attribute{
								"inline_timeout_action": schema.StringAttribute{
									Optional: true, Computed: true,
									PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
									Description:   "Action on inline timeout ('allow' or 'block').",
									Validators: []validator.String{
										stringvalidator.OneOf("allow", "block"),
									},
								},
								"max_inline_latency": schema.Int64Attribute{
									Optional: true, Computed: true,
									PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
									Description:   "Maximum inline latency in seconds.",
								},
							},
						},
						"data_protection": schema.SingleNestedBlock{
							Description: "Data protection configuration.",
							Blocks: map[string]schema.Block{
								"data_leak_detection": schema.SingleNestedBlock{
									Description: "Data leak detection configuration.",
									Attributes: map[string]schema.Attribute{
										"action": schema.StringAttribute{
											Optional: true, Computed: true,
											PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
											Description:   "Action on detection: 'block' or 'allow'.",
										},
										"mask_data_inline": schema.BoolAttribute{
											Optional: true, Computed: true,
											PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
											Description:   "Whether to mask detected data inline.",
										},
									},
									Blocks: map[string]schema.Block{
										"member": schema.ListNestedBlock{
											Description: "Data leak detection members.",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"text": schema.StringAttribute{
														Required:    true,
														Description: "Member text identifier.",
													},
													"id": schema.StringAttribute{
														Optional:      true,
														Computed:      true,
														PlanModifiers: []planmodifier.String{sameNamedStringState("text")},
														Description:   "Member ID.",
													},
													"version": schema.StringAttribute{
														Optional:      true,
														Computed:      true,
														PlanModifiers: []planmodifier.String{sameNamedStringState("text")},
														Description:   "Member version.",
													},
												},
											},
										},
									},
								},
								"database_security": schema.ListNestedBlock{
									Description: "Database security CRUD action configuration.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Required:    true,
												Description: "Database operation name (e.g., 'database-security-create').",
											},
											"action": schema.StringAttribute{
												Required:    true,
												Description: "Action: 'block' or 'allow'.",
												Validators: []validator.String{
													stringvalidator.OneOf("block", "allow"),
												},
											},
										},
									},
								},
							},
						},
						"app_protection": schema.SingleNestedBlock{
							Description: "Application protection URL category configuration.",
							Attributes: map[string]schema.Attribute{
								"alert_url_category": schema.ListAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "URL categories to alert on.",
								},
								"block_url_category": schema.ListAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "URL categories to block.",
								},
								"allow_url_category": schema.ListAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "URL categories to allow.",
								},
								"default_url_category": schema.ListAttribute{
									Optional:      true,
									Computed:      true,
									PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
									ElementType:   types.StringType,
									Description:   "Default URL categories (e.g., 'malicious').",
								},
								"url_detected_action": schema.StringAttribute{
									Optional:      true,
									Computed:      true,
									PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
									Description:   "Action when a URL matches configured categories: 'block' or empty to disable.",
								},
							},
							Blocks: map[string]schema.Block{
								"malicious_code_protection": schema.SingleNestedBlock{
									Description: "Malicious code protection configuration.",
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
											Description:   "Protection name (e.g., 'malicious-code').",
										},
										"action": schema.StringAttribute{
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
											Description:   "Action to take: 'block' or 'allow'.",
											Validators: []validator.String{
												stringvalidator.OneOf("block", "allow"),
											},
										},
									},
								},
							},
						},
						"model_protection": schema.ListNestedBlock{
							Description: "Model protection rules.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: "Protection name: 'prompt-injection', 'toxic-content', 'contextual-grounding', or 'topic-guardrails'.",
										Validators: []validator.String{
											stringvalidator.OneOf("prompt-injection", "toxic-content", "contextual-grounding", "topic-guardrails"),
										},
									},
									"action": schema.StringAttribute{
										Required:    true,
										Description: "Action to take: 'block', 'allow', or compound toxic-content values like 'high:block, moderate:allow'. Empty is preserved only for imported toxic-content with explicit categories; its semantics are not inferred.",
										Validators: []validator.String{
											stringvalidator.OneOf(
												"block", "allow", "",
												string(airsruntime.ToxicContentHighBlockModerateAllow),
												string(airsruntime.ToxicContentHighBlockModerateBlock),
												string(airsruntime.ToxicContentHighAllowModerateAllow),
											),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"toxic_category": schema.ListNestedBlock{
										Description: "Per-category overrides for toxic content detection.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"category": schema.StringAttribute{
													Required:    true,
													Description: "Category name (e.g., 'harassment', 'violence', 'hate-speech', 'sexual-content').",
												},
												"action": schema.StringAttribute{
													Required:    true,
													Description: "Action for this category.",
												},
											},
										},
									},
									"topic_list": schema.ListNestedBlock{
										Description: "Topic-based detection configuration.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"action": schema.StringAttribute{
													Required:    true,
													Description: "Action for matched topics.",
												},
											},
											Blocks: map[string]schema.Block{
												"topic": schema.ListNestedBlock{
													Description: "Topic references.",
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"topic_name": schema.StringAttribute{
																Required:    true,
																Description: "Topic name.",
															},
															"topic_id": schema.StringAttribute{
																Optional:      true,
																Computed:      true,
																PlanModifiers: []planmodifier.String{sameNamedStringState("topic_name")},
																Description:   "Topic ID.",
															},
															"revision": schema.Int64Attribute{
																Optional:      true,
																Computed:      true,
																Description:   "Topic revision.",
																PlanModifiers: []planmodifier.Int64{sameNamedInt64State("topic_name")},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"agent_protection": schema.ListNestedBlock{
							Description: "Agent protection rules.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: "Protection name: 'agent-security'.",
										Validators: []validator.String{
											stringvalidator.OneOf("agent-security"),
										},
									},
									"action": schema.StringAttribute{
										Required:    true,
										Description: "Action to take: 'block'.",
										Validators: []validator.String{
											stringvalidator.OneOf("block"),
										},
									},
								},
							},
						},
					},
				},
			},
			"dlp_data_profile": schema.ListNestedBlock{
				Description: "DLP data profile configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "Profile name.",
						},
						"uuid": schema.StringAttribute{
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "Profile UUID.",
						},
						"profile_id": schema.StringAttribute{
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "Profile ID.",
						},
						"version": schema.StringAttribute{
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "Profile version.",
						},
						"log_severity": schema.StringAttribute{
							Required:    true,
							Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
							Description: "Log severity level.",
						},
						"non_file_based": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "Non-file-based detection action.",
						},
						"file_based": schema.StringAttribute{
							Optional: true, Computed: true,
							PlanModifiers: []planmodifier.String{sameDLPReferenceState()},
							Description:   "File-based detection action.",
						},
					},
				},
			},
		},
	}
	extendProfileSchema(&resp.Schema)
}

// ── CRUD ─────────────────────────────────────────────────────────────

func (r *securityProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *securityProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() {
		validateProfileActions(&plan, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	r.resolveTopicRefs(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := airsruntime.CreateProfileRequest{
		ProfileName: plan.ProfileName.ValueString(),
		Policy:      planToSDKPolicy(ctx, &plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !profileNameAvailable(ctx, r.client, plan.ProfileName.ValueString(), &resp.Diagnostics) {
		return
	}
	profile, err := r.client.Profiles.Create(ctx, createReq)
	if err != nil {
		profileOperationError("create", plan.ProfileName.ValueString(), err, &resp.Diagnostics)
		return
	}
	profile = readCreatedProfile(ctx, r.client, profile, &plan, &resp.Diagnostics)
	if profile == nil {
		preservePartialProfileState(ctx, &plan, &resp.State, &resp.Diagnostics)
		return
	}

	mapProfileToState(ctx, profile, &plan, &resp.Diagnostics)
	saveProfileSnapshot(ctx, resp.Private, profile, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := latestNamedProfile(ctx, r.client, state.ProfileName.ValueString())
	if err != nil {

		resp.Diagnostics.AddError("Failed to read security profile", err.Error())
		return
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.CreatedAt = profileCreationTime(ctx, r.client, found.ProfileName, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	mapProfileToState(ctx, found, &state, &resp.Diagnostics)
	saveProfileSnapshot(ctx, resp.Private, found, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() {
		validateProfileActions(&plan, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.resolveTopicRefs(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := airsruntime.UpdateProfileRequest{
		ProfileJSON: airsruntime.ProfileJSON{Extensions: preservedProfileExtensions(ctx, req.Private, &resp.Diagnostics)},
		ProfileName: plan.ProfileName.ValueString(),
		Policy:      preservedProfilePolicy(ctx, req.Private, &state, &plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var profile *airsruntime.SecurityProfile
	var err error
	if !plan.ProfileName.Equal(state.ProfileName) {
		// Rename starts a new logical profile; the old name remains in AIRS.
		if !profileNameAvailable(ctx, r.client, plan.ProfileName.ValueString(), &resp.Diagnostics) {
			return
		}
		profile, err = r.client.Profiles.Create(ctx, airsruntime.CreateProfileRequest{
			ProfileName: updateReq.ProfileName, Policy: updateReq.Policy, ProfileJSON: updateReq.ProfileJSON,
		})
	} else {
		profile, err = r.client.Profiles.Update(ctx, state.ProfileID.ValueString(), updateReq)
	}
	if err != nil {
		profileOperationError("update", plan.ProfileName.ValueString(), err, &resp.Diagnostics)
		return
	}
	profile = readCreatedProfile(ctx, r.client, profile, &plan, &resp.Diagnostics)
	if profile == nil {
		preservePartialProfileState(ctx, &plan, &resp.State, &resp.Diagnostics)
		return
	}

	mapProfileToState(ctx, profile, &plan, &resp.Diagnostics)
	saveProfileSnapshot(ctx, resp.Private, profile, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profiles, err := namedProfileRevisions(ctx, r.client, state.ProfileName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list security profile history", err.Error())
		return
	}
	for _, profile := range profiles {
		_, err = r.client.Profiles.ForceDelete(ctx, profile.ProfileID, "terraform")
		id := profile.ProfileID
		tfutil.FinishDelete(ctx, err, "security profile revision "+id, func(ctx context.Context) (bool, error) {
			remaining, readErr := namedProfileRevisions(ctx, r.client, state.ProfileName.ValueString())
			for _, item := range remaining {
				if item.ProfileID == id {
					return false, readErr
				}
			}
			return true, readErr
		}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	remaining, err := namedProfileRevisions(ctx, r.client, state.ProfileName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify security profile destruction", err.Error())
	} else if len(remaining) != 0 {
		resp.Diagnostics.AddError("Security profile history remains", "Revisions still exist under the managed name; retry destroy to remove them.")
	}
}

func (r *securityProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	found, err := latestNamedProfile(ctx, r.client, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import security profile", err.Error())
		return
	}

	if found == nil {
		resp.Diagnostics.AddError("Profile not found", "No profile with name: "+req.ID)
		return
	}

	var state SecurityProfileResourceModel
	state.CreatedAt = profileCreationTime(ctx, r.client, found.ProfileName, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	mapProfileToState(ctx, found, &state, &resp.Diagnostics)
	saveProfileSnapshot(ctx, resp.Private, found, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Topic resolution ─────────────────────────────────────────────────

// resolveTopicRefs looks up all custom topics and fills in topic_id and
// revision for any topic ref that only has a topic_name.
func (r *securityProfileResource) resolveTopicRefs(ctx context.Context, plan *SecurityProfileResourceModel, diags *diag.Diagnostics) {
	// Collect topic names that need resolution.
	var needsResolve bool
	for _, protection := range profileProtectionModels(plan) {
		for _, mp := range protection.ModelProtection {
			for _, tl := range mp.TopicLists {
				for _, t := range tl.Topics {
					if !t.TopicName.IsNull() && !t.TopicName.IsUnknown() && t.TopicName.ValueString() != "" &&
						(t.TopicID.IsNull() || t.TopicID.IsUnknown() || t.TopicID.ValueString() == "") {
						needsResolve = true
					}
				}
			}
		}
	}
	if !needsResolve {
		return
	}

	// Fetch all topics (paginate).
	nameMap := make(map[string]airsruntime.CustomTopic)
	offset := 0
	for {
		result, err := r.client.Topics.List(ctx, airsruntime.ListOpts{Limit: 200, Offset: offset})
		if err != nil {
			diags.AddError("Failed to list topics for name resolution", err.Error())
			return
		}
		for _, t := range result.Items {
			nameMap[t.TopicName] = t
		}
		if result.NextOffset <= offset || len(result.Items) == 0 {
			break
		}
		offset = result.NextOffset
	}

	// Fill in topic_id and revision.
	for _, protection := range profileProtectionModels(plan) {
		for mpIdx := range protection.ModelProtection {
			for tlIdx := range protection.ModelProtection[mpIdx].TopicLists {
				for tIdx := range protection.ModelProtection[mpIdx].TopicLists[tlIdx].Topics {
					ref := &protection.ModelProtection[mpIdx].TopicLists[tlIdx].Topics[tIdx]
					name := ref.TopicName.ValueString()
					if name == "" {
						continue
					}
					if ref.TopicID.IsNull() || ref.TopicID.IsUnknown() || ref.TopicID.ValueString() == "" {
						if resolved, ok := nameMap[name]; ok {
							tflog.Debug(ctx, "resolved topic", map[string]any{
								"name": name, "id": resolved.TopicID, "revision": resolved.Revision,
							})
							ref.TopicID = types.StringValue(resolved.TopicID)
							ref.Revision = types.Int64Value(resolved.Revision)
						} else {
							diags.AddError("Topic not found",
								"Custom topic '"+name+"' does not exist. Create it first or specify topic_id directly.")
						}
					}
				}
			}
		}
	}
}

// ── Conversion: Terraform plan → SDK ─────────────────────────────────

func planToSDKPolicy(ctx context.Context, plan *SecurityProfileResourceModel, diags *diag.Diagnostics) *airsruntime.ProfilePolicy {
	if len(plan.AiSecurityProfiles) == 0 && len(plan.DlpDataProfiles) == 0 {
		return nil
	}

	policy := &airsruntime.ProfilePolicy{}

	for _, asp := range plan.AiSecurityProfiles {
		config := airsruntime.AiSecurityProfileConfig{
			ModelType:   asp.ModelType.ValueString(),
			ContentType: asp.ContentType.ValueString(),
		}

		mc := &airsruntime.ModelConfiguration{
			MaskDataInStorage: asp.MaskDataInStorage.ValueBool(),
		}

		if asp.Latency != nil {
			mc.Latency = &airsruntime.LatencyConfig{
				InlineTimeoutAction: airsruntime.ProfileAction(asp.Latency.InlineTimeoutAction.ValueString()),
				MaxInlineLatency:    int32(asp.Latency.MaxInlineLatency.ValueInt64()),
			}
			setStringPresence(&mc.Latency.ProfileJSON, "inline-timeout-action", asp.Latency.InlineTimeoutAction)
			if !asp.Latency.MaxInlineLatency.IsNull() && !asp.Latency.MaxInlineLatency.IsUnknown() {
				mc.Latency.SetFieldPresence("max-inline-latency", airsruntime.JSONPresent)
			}
		}

		protection := protectionToSDK(ctx, asp.protection(), diags)
		mc.DataProtection, mc.AppProtection = protection.DataProtection, protection.AppProtection
		mc.ModelProtection, mc.AgentProtection = protection.ModelProtection, protection.AgentProtection
		setBoolPresence(&mc.ProfileJSON, "mask-data-in-storage", asp.MaskDataInStorage)
		if !asp.EnableFullConversationInspection.IsNull() && !asp.EnableFullConversationInspection.IsUnknown() {
			value := asp.EnableFullConversationInspection.ValueBool()
			mc.EnableFullConversationInspection = &value
		}
		config.ContentTypeMode = asp.ContentTypeMode.ValueString()
		setStringPresence(&config.ProfileJSON, "content-type", asp.ContentType)
		setStringPresence(&config.ProfileJSON, "content-type-mode", asp.ContentTypeMode)
		config.ContentTypeConfigurations = directionsToSDK(ctx, asp.ContentTypeConfigurations, diags)

		config.ModelConfiguration = mc
		policy.AiSecurityProfiles = append(policy.AiSecurityProfiles, config)
	}

	for _, dlp := range plan.DlpDataProfiles {
		policy.DlpDataProfiles = append(policy.DlpDataProfiles, airsruntime.DLPDataProfileConfig{
			Name:         dlp.Name.ValueString(),
			UUID:         dlp.UUID.ValueString(),
			ID:           dlp.ProfileID.ValueString(),
			Version:      dlp.Version.ValueString(),
			LogSeverity:  dlp.LogSeverity.ValueString(),
			NonFileBased: dlp.NonFileBased.ValueString(),
			FileBased:    dlp.FileBased.ValueString(),
		})
	}

	return policy
}

func listToURLCategory(ctx context.Context, list types.List, diags *diag.Diagnostics) *airsruntime.URLCategoryMember {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	members := []string{}
	diags.Append(list.ElementsAs(ctx, &members, false)...)
	return &airsruntime.URLCategoryMember{Member: members}
}

// ── Conversion: SDK → Terraform state ────────────────────────────────

func mapProfileToState(ctx context.Context, profile *airsruntime.SecurityProfile, state *SecurityProfileResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(profile.ProfileID)
	state.ProfileID = types.StringValue(profile.ProfileID)
	state.Revision = types.Int64Value(int64(profile.Revision))
	state.ProfileName = types.StringValue(profile.ProfileName)
	state.Active = types.BoolValue(profile.Active)

	state.UpdatedAt = types.StringValue(profile.LastModifiedTs)
	// GET can omit response-only metadata. Keep a previously observed tenant.
	if profile.FieldPresence("dlp_tenant_id") != airsruntime.JSONOmitted {
		state.DLPTenantID = types.StringValue(profile.DLPTenantID)
	} else if state.DLPTenantID.IsUnknown() {
		state.DLPTenantID = types.StringNull()
	}

	if profile.Policy == nil {
		state.AiSecurityProfiles = nil
		state.DlpDataProfiles = nil
		return
	}

	priorAiProfiles := state.AiSecurityProfiles
	state.AiSecurityProfiles = nil
	for _, asp := range profile.Policy.AiSecurityProfiles {
		model := AiSecurityProfileModel{
			ModelType: types.StringValue(asp.ModelType),
		}
		model.ContentType = sdkString(asp.ContentType, asp.FieldPresence("content-type"))
		model.ContentTypeMode = sdkString(asp.ContentTypeMode, asp.FieldPresence("content-type-mode"))
		model.ContentTypeConfigurations = directionsFromSDK(ctx, asp.ContentTypeConfigurations, priorDirections(priorAiProfiles, asp), diags)

		if asp.ModelConfiguration != nil {
			mc := asp.ModelConfiguration
			model.MaskDataInStorage = types.BoolValue(mc.MaskDataInStorage)

			if mc.Latency != nil {
				model.Latency = &LatencyModel{
					InlineTimeoutAction: types.StringValue(string(mc.Latency.InlineTimeoutAction)),
					MaxInlineLatency:    types.Int64Value(int64(mc.Latency.MaxInlineLatency)),
				}
			}

			protection := protectionFromSDK(ctx, &airsruntime.ProtectionConfiguration{
				DataProtection: mc.DataProtection, AppProtection: mc.AppProtection,
				ModelProtection: mc.ModelProtection, AgentProtection: mc.AgentProtection,
			}, priorProfileProtection(priorAiProfiles, asp), diags)
			model.DataProtection, model.AppProtection = protection.DataProtection, protection.AppProtection
			model.ModelProtection, model.AgentProtection = protection.ModelProtection, protection.AgentProtection
			if mc.FieldPresence("mask-data-in-storage") == airsruntime.JSONOmitted {
				model.MaskDataInStorage = types.BoolNull()
			}
			model.EnableFullConversationInspection = types.BoolPointerValue(mc.EnableFullConversationInspection)

		}

		state.AiSecurityProfiles = append(state.AiSecurityProfiles, model)
	}

	state.DlpDataProfiles = nil
	for _, dlp := range profile.Policy.DlpDataProfiles {
		dlpModel := DlpDataProfileModel{
			Name:        types.StringValue(dlp.Name),
			UUID:        types.StringValue(dlp.UUID),
			ProfileID:   types.StringValue(dlp.ID),
			Version:     types.StringValue(dlp.Version),
			LogSeverity: types.StringValue(dlp.LogSeverity),
		}
		dlpModel.NonFileBased = types.StringValue(dlp.NonFileBased)
		dlpModel.FileBased = types.StringValue(dlp.FileBased)
		state.DlpDataProfiles = append(state.DlpDataProfiles, dlpModel)
	}
}

func urlCategoryToList(ctx context.Context, cat *airsruntime.URLCategoryMember, diags *diag.Diagnostics) types.List {
	if cat == nil || cat.FieldPresence("member") != airsruntime.JSONPresent {
		return types.ListNull(types.StringType)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, cat.Member)
	diags.Append(d...)
	return list
}
