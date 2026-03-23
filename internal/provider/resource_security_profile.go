package provider

import (
	"context"
	"strings"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &securityProfileResource{}
	_ resource.ResourceWithImportState = &securityProfileResource{}
)

func NewSecurityProfileResource() resource.Resource {
	return &securityProfileResource{}
}

type securityProfileResource struct {
	client *airsruntime.Client
}

// ── Model types ──────────────────────────────────────────────────────

type SecurityProfileResourceModel struct {
	ID                 types.String             `tfsdk:"id"`
	ProfileID          types.String             `tfsdk:"profile_id"`
	ProfileName        types.String             `tfsdk:"profile_name"`
	Active             types.Bool               `tfsdk:"active"`
	CreatedAt          types.String             `tfsdk:"created_at"`
	UpdatedAt          types.String             `tfsdk:"updated_at"`
	AiSecurityProfiles []AiSecurityProfileModel `tfsdk:"ai_security_profile"`
	DlpDataProfiles    []DlpDataProfileModel    `tfsdk:"dlp_data_profile"`
}

type AiSecurityProfileModel struct {
	ModelType         types.String           `tfsdk:"model_type"`
	ContentType       types.String           `tfsdk:"content_type"`
	MaskDataInStorage types.Bool             `tfsdk:"mask_data_in_storage"`
	Latency           *LatencyModel          `tfsdk:"latency"`
	DataProtection    *DataProtectionModel   `tfsdk:"data_protection"`
	AppProtection     *AppProtectionModel    `tfsdk:"app_protection"`
	ModelProtection   []ModelProtectionModel `tfsdk:"model_protection"`
	AgentProtection   []AgentProtectionModel `tfsdk:"agent_protection"`
}

type LatencyModel struct {
	InlineTimeoutAction types.String `tfsdk:"inline_timeout_action"`
	MaxInlineLatency    types.Int64  `tfsdk:"max_inline_latency"`
}

type DataProtectionModel struct {
	DataLeakDetection *DataLeakDetectionModel `tfsdk:"data_leak_detection"`
	DatabaseSecurity  []DatabaseSecurityModel `tfsdk:"database_security"`
}

type DatabaseSecurityModel struct {
	Name   types.String `tfsdk:"name"`
	Action types.String `tfsdk:"action"`
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
	AlertURLCategory        types.List                    `tfsdk:"alert_url_category"`
	BlockURLCategory        types.List                    `tfsdk:"block_url_category"`
	AllowURLCategory        types.List                    `tfsdk:"allow_url_category"`
	DefaultURLCategory      types.List                    `tfsdk:"default_url_category"`
	UrlDetectedAction       types.String                  `tfsdk:"url_detected_action"`
	MaliciousCodeProtection *MaliciousCodeProtectionModel `tfsdk:"malicious_code_protection"`
}

type MaliciousCodeProtectionModel struct {
	Name   types.String `tfsdk:"name"`
	Action types.String `tfsdk:"action"`
}

type ModelProtectionModel struct {
	Name            types.String         `tfsdk:"name"`
	Action          types.String         `tfsdk:"action"`
	ToxicCategories []ToxicCategoryModel `tfsdk:"toxic_category"`
	TopicLists      []TopicListModel     `tfsdk:"topic_list"`
}

type ToxicCategoryModel struct {
	Category types.String `tfsdk:"category"`
	Action   types.String `tfsdk:"action"`
}

type TopicListModel struct {
	Action types.String    `tfsdk:"action"`
	Topics []TopicRefModel `tfsdk:"topic"`
}

type TopicRefModel struct {
	TopicName types.String `tfsdk:"topic_name"`
	TopicID   types.String `tfsdk:"topic_id"`
	Revision  types.Int64  `tfsdk:"revision"`
}

type AgentProtectionModel struct {
	Name   types.String `tfsdk:"name"`
	Action types.String `tfsdk:"action"`
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
	resp.TypeName = req.ProviderTypeName + "_security_profile"
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
			"profile_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the security profile.",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the profile is active.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
							Optional:    true,
							Description: "Model type (e.g., 'default').",
						},
						"content_type": schema.StringAttribute{
							Optional:    true,
							Description: "Content type.",
						},
						"mask_data_in_storage": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to mask data in storage.",
						},
					},
					Blocks: map[string]schema.Block{
						"latency": schema.SingleNestedBlock{
							Description: "Latency configuration for inline scanning.",
							Attributes: map[string]schema.Attribute{
								"inline_timeout_action": schema.StringAttribute{
									Optional:    true,
									Description: "Action on inline timeout ('allow' or 'block').",
									Validators: []validator.String{
										stringvalidator.OneOf("allow", "block"),
									},
								},
								"max_inline_latency": schema.Int64Attribute{
									Optional:    true,
									Description: "Maximum inline latency in seconds.",
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
											Optional:    true,
											Description: "Action on detection: 'block' or 'allow'.",
										},
										"mask_data_inline": schema.BoolAttribute{
											Optional:    true,
											Description: "Whether to mask detected data inline.",
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
														Optional:    true,
														Computed:    true,
														Description: "Member ID.",
													},
													"version": schema.StringAttribute{
														Optional:    true,
														Computed:    true,
														Description: "Member version.",
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
									Optional:    true,
									Computed:    true,
									ElementType: types.StringType,
									Description: "Default URL categories (e.g., 'malicious').",
								},
								"url_detected_action": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Action when a URL matches configured categories: 'block' or empty to disable.",
								},
							},
							Blocks: map[string]schema.Block{
								"malicious_code_protection": schema.SingleNestedBlock{
									Description: "Malicious code protection configuration.",
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Optional:    true,
											Computed:    true,
											Description: "Protection name (e.g., 'malicious-code').",
										},
										"action": schema.StringAttribute{
											Optional:    true,
											Computed:    true,
											Description: "Action to take: 'block' or 'allow'.",
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
										Description: "Action to take: 'block', 'allow', or compound toxic-content values like 'high:block, moderate:allow'.",
										Validators: []validator.String{
											stringvalidator.OneOf(
												"block", "allow",
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
																Optional:    true,
																Computed:    true,
																Description: "Topic ID.",
															},
															"revision": schema.Int64Attribute{
																Optional:    true,
																Computed:    true,
																Description: "Topic revision.",
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
							Optional:    true,
							Description: "Profile name.",
						},
						"uuid": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Profile UUID.",
						},
						"profile_id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Profile ID.",
						},
						"version": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Profile version.",
						},
						"log_severity": schema.StringAttribute{
							Optional:    true,
							Description: "Log severity level.",
						},
						"non_file_based": schema.StringAttribute{
							Optional:    true,
							Description: "Non-file-based detection action.",
						},
						"file_based": schema.StringAttribute{
							Optional:    true,
							Description: "File-based detection action.",
						},
					},
				},
			},
		},
	}
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

	profile, err := r.client.Profiles.Create(ctx, createReq)
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			found, lookupErr := r.client.Profiles.GetByName(ctx, plan.ProfileName.ValueString())
			if lookupErr == nil && found != nil {
				tflog.Warn(ctx, "profile create returned 409 but profile exists; treating as success")
				mapProfileToState(ctx, found, &plan, &resp.Diagnostics)
				resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
				return
			}
		}
		resp.Diagnostics.AddError("Failed to create security profile", err.Error())
		return
	}

	mapProfileToState(ctx, profile, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.client.Profiles.GetByID(ctx, state.ProfileID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read security profile", err.Error())
		return
	}

	mapProfileToState(ctx, found, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := airsruntime.UpdateProfileRequest{
		ProfileName: plan.ProfileName.ValueString(),
		Policy:      planToSDKPolicy(ctx, &plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.Profiles.Update(ctx, state.ProfileID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update security profile", err.Error())
		return
	}

	mapProfileToState(ctx, profile, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Profiles.ForceDelete(ctx, state.ProfileID.ValueString(), "terraform")
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete security profile", err.Error())
		return
	}
}

func (r *securityProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	found, err := r.client.Profiles.GetByName(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Profile not found", "No profile with name: "+req.ID+": "+err.Error())
		return
	}

	var state SecurityProfileResourceModel
	mapProfileToState(ctx, found, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		}

		if asp.DataProtection != nil {
			dpConfig := &airsruntime.DataProtectionConfig{}

			if asp.DataProtection.DataLeakDetection != nil {
				dld := asp.DataProtection.DataLeakDetection
				sdkDLD := airsruntime.DataLeakDetectionConfig{
					Action:         airsruntime.ProfileAction(dld.Action.ValueString()),
					MaskDataInline: dld.MaskDataInline.ValueBool(),
				}
				for _, m := range dld.Members {
					sdkDLD.Member = append(sdkDLD.Member, airsruntime.DataLeakMember{
						Text:    m.Text.ValueString(),
						ID:      m.ID.ValueString(),
						Version: m.Version.ValueString(),
					})
				}
				dpConfig.DataLeakDetection = &sdkDLD
			}

			for _, ds := range asp.DataProtection.DatabaseSecurity {
				dpConfig.DatabaseSecurity = append(dpConfig.DatabaseSecurity, airsruntime.DatabaseSecurityConfig{
					Name:   ds.Name.ValueString(),
					Action: ds.Action.ValueString(),
				})
			}

			mc.DataProtection = dpConfig
		}

		if asp.AppProtection != nil {
			mc.AppProtection = &airsruntime.AppProtectionConfig{
				AlertURLCategory:   listToURLCategory(ctx, asp.AppProtection.AlertURLCategory, diags),
				BlockURLCategory:   listToURLCategory(ctx, asp.AppProtection.BlockURLCategory, diags),
				AllowURLCategory:   listToURLCategory(ctx, asp.AppProtection.AllowURLCategory, diags),
				DefaultURLCategory: listToURLCategory(ctx, asp.AppProtection.DefaultURLCategory, diags),
				UrlDetectedAction:  asp.AppProtection.UrlDetectedAction.ValueString(),
			}
			if asp.AppProtection.MaliciousCodeProtection != nil {
				mc.AppProtection.MaliciousCodeProtection = &airsruntime.MaliciousCodeProtectionConfig{
					Name:   asp.AppProtection.MaliciousCodeProtection.Name.ValueString(),
					Action: asp.AppProtection.MaliciousCodeProtection.Action.ValueString(),
				}
			}
		}

		for _, mp := range asp.ModelProtection {
			sdkMP := airsruntime.ModelProtectionConfig{
				Name:   mp.Name.ValueString(),
				Action: airsruntime.ProfileAction(mp.Action.ValueString()),
			}
			for _, tc := range mp.ToxicCategories {
				sdkMP.ToxicCategoryList = append(sdkMP.ToxicCategoryList, airsruntime.ToxicCategoryConfig{
					Category: tc.Category.ValueString(),
					Action:   tc.Action.ValueString(),
				})
			}
			for _, tl := range mp.TopicLists {
				sdkTL := airsruntime.TopicArrayConfig{
					Action: airsruntime.ProfileAction(tl.Action.ValueString()),
				}
				for _, t := range tl.Topics {
					sdkTL.Topic = append(sdkTL.Topic, airsruntime.TopicRef{
						TopicName: t.TopicName.ValueString(),
						TopicID:   t.TopicID.ValueString(),
						Revision:  t.Revision.ValueInt64(),
					})
				}
				sdkMP.TopicList = append(sdkMP.TopicList, sdkTL)
			}
			mc.ModelProtection = append(mc.ModelProtection, sdkMP)
		}

		for _, ap := range asp.AgentProtection {
			mc.AgentProtection = append(mc.AgentProtection, airsruntime.AgentProtectionConfig{
				Name:   ap.Name.ValueString(),
				Action: airsruntime.ProfileAction(ap.Action.ValueString()),
			})
		}

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
	var members []string
	diags.Append(list.ElementsAs(ctx, &members, false)...)
	if len(members) == 0 {
		return nil
	}
	return &airsruntime.URLCategoryMember{Member: members}
}

// ── Conversion: SDK → Terraform state ────────────────────────────────

func mapProfileToState(ctx context.Context, profile *airsruntime.SecurityProfile, state *SecurityProfileResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(profile.ProfileID)
	state.ProfileID = types.StringValue(profile.ProfileID)
	state.ProfileName = types.StringValue(profile.ProfileName)
	state.Active = types.BoolValue(profile.Active)
	if state.CreatedAt.IsNull() || state.CreatedAt.IsUnknown() || state.CreatedAt.ValueString() == "" {
		state.CreatedAt = types.StringValue(profile.LastModifiedTs)
	}
	state.UpdatedAt = types.StringValue(profile.LastModifiedTs)

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
		if asp.ContentType != "" {
			model.ContentType = types.StringValue(asp.ContentType)
		}

		if asp.ModelConfiguration != nil {
			mc := asp.ModelConfiguration
			model.MaskDataInStorage = types.BoolValue(mc.MaskDataInStorage)

			if mc.Latency != nil {
				model.Latency = &LatencyModel{
					InlineTimeoutAction: types.StringValue(string(mc.Latency.InlineTimeoutAction)),
					MaxInlineLatency:    types.Int64Value(int64(mc.Latency.MaxInlineLatency)),
				}
			}

			if mc.DataProtection != nil {
				dpModel := &DataProtectionModel{}

				if mc.DataProtection.DataLeakDetection != nil {
					dld := mc.DataProtection.DataLeakDetection
					dldModel := &DataLeakDetectionModel{
						Action: types.StringValue(string(dld.Action)),
					}
					if dld.MaskDataInline {
						dldModel.MaskDataInline = types.BoolValue(true)
					}
					for _, m := range dld.Member {
						member := DataLeakMemberModel{
							Text: types.StringValue(m.Text),
						}
						if m.ID != "" {
							member.ID = types.StringValue(m.ID)
						}
						if m.Version != "" {
							member.Version = types.StringValue(m.Version)
						}
						dldModel.Members = append(dldModel.Members, member)
					}
					dpModel.DataLeakDetection = dldModel
				}

				for _, ds := range mc.DataProtection.DatabaseSecurity {
					dpModel.DatabaseSecurity = append(dpModel.DatabaseSecurity, DatabaseSecurityModel{
						Name:   types.StringValue(ds.Name),
						Action: types.StringValue(ds.Action),
					})
				}

				model.DataProtection = dpModel
			}

			if mc.AppProtection != nil {
				model.AppProtection = &AppProtectionModel{
					AlertURLCategory:   urlCategoryToList(ctx, mc.AppProtection.AlertURLCategory, diags),
					BlockURLCategory:   urlCategoryToList(ctx, mc.AppProtection.BlockURLCategory, diags),
					AllowURLCategory:   urlCategoryToList(ctx, mc.AppProtection.AllowURLCategory, diags),
					DefaultURLCategory: urlCategoryToList(ctx, mc.AppProtection.DefaultURLCategory, diags),
					UrlDetectedAction:  types.StringValue(mc.AppProtection.UrlDetectedAction),
				}
				if mc.AppProtection.MaliciousCodeProtection != nil {
					model.AppProtection.MaliciousCodeProtection = &MaliciousCodeProtectionModel{
						Name:   types.StringValue(mc.AppProtection.MaliciousCodeProtection.Name),
						Action: types.StringValue(mc.AppProtection.MaliciousCodeProtection.Action),
					}
				}
			}

			for mpIdx, mp := range mc.ModelProtection {
				mpModel := ModelProtectionModel{
					Name:   types.StringValue(mp.Name),
					Action: types.StringValue(string(mp.Action)),
				}
				for _, tc := range mp.ToxicCategoryList {
					mpModel.ToxicCategories = append(mpModel.ToxicCategories, ToxicCategoryModel{
						Category: types.StringValue(tc.Category),
						Action:   types.StringValue(tc.Action),
					})
				}
				for tlIdx, tl := range mp.TopicList {
					tlModel := TopicListModel{
						Action: types.StringValue(string(tl.Action)),
					}
					if len(tl.Topic) > 0 {
						for _, t := range tl.Topic {
							tlModel.Topics = append(tlModel.Topics, TopicRefModel{
								TopicName: types.StringValue(t.TopicName),
								TopicID:   types.StringValue(t.TopicID),
								Revision:  types.Int64Value(t.Revision),
							})
						}
					} else {
						// API may not return topics; preserve from prior state
						tlModel.Topics = priorTopicsFromSlice(priorAiProfiles, mpIdx, tlIdx)
					}
					mpModel.TopicLists = append(mpModel.TopicLists, tlModel)
				}
				model.ModelProtection = append(model.ModelProtection, mpModel)
			}

			for _, ap := range mc.AgentProtection {
				model.AgentProtection = append(model.AgentProtection, AgentProtectionModel{
					Name:   types.StringValue(ap.Name),
					Action: types.StringValue(string(ap.Action)),
				})
			}
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
		if dlp.NonFileBased != "" {
			dlpModel.NonFileBased = types.StringValue(dlp.NonFileBased)
		}
		if dlp.FileBased != "" {
			dlpModel.FileBased = types.StringValue(dlp.FileBased)
		}
		state.DlpDataProfiles = append(state.DlpDataProfiles, dlpModel)
	}
}

func priorTopicsFromSlice(priorProfiles []AiSecurityProfileModel, mpIdx, tlIdx int) []TopicRefModel {
	if len(priorProfiles) == 0 {
		return nil
	}
	asp := priorProfiles[0]
	if mpIdx >= len(asp.ModelProtection) {
		return nil
	}
	mp := asp.ModelProtection[mpIdx]
	if tlIdx >= len(mp.TopicLists) {
		return nil
	}
	// Resolve any unknown values (from plan) to concrete defaults
	topics := make([]TopicRefModel, len(mp.TopicLists[tlIdx].Topics))
	for i, t := range mp.TopicLists[tlIdx].Topics {
		topics[i] = TopicRefModel{
			TopicName: t.TopicName,
		}
		if t.TopicID.IsUnknown() {
			topics[i].TopicID = types.StringValue("")
		} else {
			topics[i].TopicID = t.TopicID
		}
		if t.Revision.IsUnknown() {
			topics[i].Revision = types.Int64Value(0)
		} else {
			topics[i].Revision = t.Revision
		}
	}
	return topics
}

func urlCategoryToList(ctx context.Context, cat *airsruntime.URLCategoryMember, diags *diag.Diagnostics) types.List {
	if cat == nil || len(cat.Member) == 0 {
		return types.ListNull(types.StringType)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, cat.Member)
	diags.Append(d...)
	return list
}

func getMgmtClient(data any) (*airsruntime.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	pd, ok := data.(*ProviderData)
	if !ok || pd == nil {
		diags.AddError("Missing provider config", "Provider not configured")
		return nil, diags
	}
	if pd.MgmtClient == nil {
		diags.AddError("Management client not configured", "OAuth2 credentials required")
		return nil, diags
	}
	return pd.MgmtClient, diags
}
