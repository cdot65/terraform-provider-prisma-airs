package runtime

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.ResourceWithValidateConfig = &securityProfileResource{}

func profileOptionalString(description string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Computed: true, Description: description,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
}
func profileSeverityString(description, identity string) schema.StringAttribute {
	a := profileOptionalString(description)
	a.PlanModifiers = []planmodifier.String{sameNamedStringState(identity)}
	return a
}
func confidenceSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{Description: "Optional toxicity severity overrides by confidence; values are open-ended.", Attributes: map[string]schema.Attribute{
		"high": profileOptionalString("Severity at high confidence."), "moderate": profileOptionalString("Severity at moderate confidence."),
	}}
}

// Extend the legacy schema in place, then reuse its protection families for all directions.
// No defaults are introduced in directional protection settings.
func extendProfileSchema(s *schema.Schema) {
	s.Attributes["dlp_tenant_id"] = schema.StringAttribute{Computed: true, Description: "Optional DLP tenant metadata observed in write responses; retrieval may omit it."}
	asp := s.Blocks["ai_security_profile"].(schema.ListNestedBlock)
	asp.NestedObject.Attributes["content_type_mode"] = profileOptionalString("Content layout mode, for example per_content_type. Values are open-ended.")
	asp.NestedObject.Attributes["enable_full_conversation_inspection"] = schema.BoolAttribute{Optional: true, Computed: true, Description: "Inspect the full conversation. Omitted values remain omitted; false is explicit."}
	blocks := asp.NestedObject.Blocks
	dp := blocks["data_protection"].(schema.SingleNestedBlock)
	ds := dp.Blocks["database_security"].(schema.ListNestedBlock)
	ds.NestedObject.Attributes["severity"] = profileSeverityString("Database detector severity.", "name")
	dp.Blocks["database_security"] = ds
	dp.Blocks["source_code_detection"] = schema.SingleNestedBlock{Description: "Source-code detection settings.", Attributes: map[string]schema.Attribute{
		"action": profileOptionalString("Action on source-code detection."), "severity": profileOptionalString("Source-code detector severity."),
	}}
	blocks["data_protection"] = dp
	app := blocks["app_protection"].(schema.SingleNestedBlock)
	app.Attributes["url_detected_severity"] = profileOptionalString("URL detector severity.")
	malicious := app.Blocks["malicious_code_protection"].(schema.SingleNestedBlock)
	malicious.Attributes["severity"] = profileSeverityString("Malicious-code detector severity.", "name")
	app.Blocks["malicious_code_protection"] = malicious
	blocks["app_protection"] = app
	mp := blocks["model_protection"].(schema.ListNestedBlock)
	mp.NestedObject.Attributes["severity"] = profileSeverityString("Model detector severity.", "name")
	mp.NestedObject.Blocks["severity_by_confidence"] = confidenceSchema()
	tc := mp.NestedObject.Blocks["toxic_category"].(schema.ListNestedBlock)
	tc.NestedObject.Blocks = map[string]schema.Block{"severity_by_confidence": confidenceSchema()}
	mp.NestedObject.Blocks["toxic_category"] = tc
	tl := mp.NestedObject.Blocks["topic_list"].(schema.ListNestedBlock)
	topic := tl.NestedObject.Blocks["topic"].(schema.ListNestedBlock)
	topic.NestedObject.Attributes["severity"] = profileSeverityString("Topic-reference severity.", "topic_name")
	tl.NestedObject.Blocks["topic"] = topic
	mp.NestedObject.Blocks["topic_list"] = tl
	// Detector identifiers and actions may evolve, including compound action strings.
	for _, name := range []string{"name", "action"} {
		a := mp.NestedObject.Attributes[name].(schema.StringAttribute)
		a.Validators = nil
		if name == "name" {
			a.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
		}
		mp.NestedObject.Attributes[name] = a
	}
	blocks["model_protection"] = mp
	agent := blocks["agent_protection"].(schema.ListNestedBlock)
	agent.NestedObject.Attributes["severity"] = profileSeverityString("Agent detector severity.", "name")
	for _, name := range []string{"name", "action"} {
		a := agent.NestedObject.Attributes[name].(schema.StringAttribute)
		a.Validators = []validator.String{stringvalidator.LengthAtLeast(1)}
		agent.NestedObject.Attributes[name] = a
	}
	blocks["agent_protection"] = agent
	directions := map[string]schema.Block{}
	for _, name := range []string{"prompt", "response", "tool_call", "tool_response"} {
		directions[name] = schema.SingleNestedBlock{Description: "Independent protections for " + name + " content.", Blocks: map[string]schema.Block{
			"data_protection": dp, "app_protection": app, "model_protection": mp, "agent_protection": agent,
		}}
	}
	blocks["content_type_configurations"] = schema.SingleNestedBlock{Description: "Per-direction protections. Shared latency, storage and conversation settings remain on ai_security_profile.", Blocks: directions}
	s.Blocks["ai_security_profile"] = asp
}

func (r *securityProfileResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config SecurityProfileResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if req.Config.Raw.IsFullyKnown() {
		validateProfileActions(&config, &resp.Diagnostics)
	}
	for _, p := range config.AiSecurityProfiles {
		if p.ContentTypeConfigurations != nil && (p.DataProtection != nil || p.AppProtection != nil || len(p.ModelProtection) > 0 || len(p.AgentProtection) > 0) {
			resp.Diagnostics.AddError("Conflicting protection layouts", "Configure legacy protection blocks or content_type_configurations within each ai_security_profile, while keeping latency, storage and conversation settings shared. Imported mixed policies can be read; resolve their protection ownership before mutation.")
		}
	}
}
