package runtime

import (
	"context"

	airsruntime "github.com/cdot65/prisma-airs-go/aisec/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProtectionModel is shared by the legacy layout and every direction.
type ProtectionModel struct {
	DataProtection  *DataProtectionModel   `tfsdk:"data_protection"`
	AppProtection   *AppProtectionModel    `tfsdk:"app_protection"`
	ModelProtection []ModelProtectionModel `tfsdk:"model_protection"`
	AgentProtection []AgentProtectionModel `tfsdk:"agent_protection"`
}
type ContentTypeConfigurationsModel struct {
	Prompt       *ProtectionModel `tfsdk:"prompt"`
	Response     *ProtectionModel `tfsdk:"response"`
	ToolCall     *ProtectionModel `tfsdk:"tool_call"`
	ToolResponse *ProtectionModel `tfsdk:"tool_response"`
}
type SeverityByConfidenceModel struct {
	High     types.String `tfsdk:"high"`
	Moderate types.String `tfsdk:"moderate"`
}
type SourceCodeDetectionModel struct {
	Action   types.String `tfsdk:"action"`
	Severity types.String `tfsdk:"severity"`
}

func (m AiSecurityProfileModel) protection() *ProtectionModel {
	return &ProtectionModel{m.DataProtection, m.AppProtection, m.ModelProtection, m.AgentProtection}
}
func protectionToSDK(ctx context.Context, model *ProtectionModel, diags *diag.Diagnostics) *airsruntime.ProtectionConfiguration {
	if model == nil {
		return nil
	}
	mc := &airsruntime.ProtectionConfiguration{}
	if model.DataProtection != nil {
		dpConfig := &airsruntime.DataProtectionConfig{}

		if model.DataProtection.DataLeakDetection != nil {
			dld := model.DataProtection.DataLeakDetection
			sdkDLD := airsruntime.DataLeakDetectionConfig{
				Action:         airsruntime.ProfileAction(dld.Action.ValueString()),
				MaskDataInline: dld.MaskDataInline.ValueBool(),
			}
			setBoolPresence(&sdkDLD.ProfileJSON, "mask-data-inline", dld.MaskDataInline)
			setStringPresence(&sdkDLD.ProfileJSON, "action", dld.Action)
			for _, m := range dld.Members {
				member := airsruntime.DataLeakMember{
					Text:    m.Text.ValueString(),
					ID:      m.ID.ValueString(),
					Version: m.Version.ValueString(),
				}
				setStringPresence(&member.ProfileJSON, "id", m.ID)
				setStringPresence(&member.ProfileJSON, "version", m.Version)
				sdkDLD.Member = append(sdkDLD.Member, member)
			}
			dpConfig.DataLeakDetection = &sdkDLD
		}

		for _, ds := range model.DataProtection.DatabaseSecurity {
			rule := airsruntime.DatabaseSecurityConfig{
				Severity: ds.Severity.ValueString(),
				Name:     ds.Name.ValueString(),
				Action:   ds.Action.ValueString(),
			}
			setStringPresence(&rule.ProfileJSON, "severity", ds.Severity)
			dpConfig.DatabaseSecurity = append(dpConfig.DatabaseSecurity, rule)
		}

		if model.DataProtection.SourceCodeDetection != nil {
			source := model.DataProtection.SourceCodeDetection
			dpConfig.SourceCodeDetection = &airsruntime.SourceCodeDetectionConfig{Action: airsruntime.ProfileAction(source.Action.ValueString()), Severity: source.Severity.ValueString()}
			setStringPresence(&dpConfig.SourceCodeDetection.ProfileJSON, "action", source.Action)
			setStringPresence(&dpConfig.SourceCodeDetection.ProfileJSON, "severity", source.Severity)
		}
		mc.DataProtection = dpConfig
	}

	if model.AppProtection != nil {
		mc.AppProtection = &airsruntime.AppProtectionConfig{
			AlertURLCategory:    listToURLCategory(ctx, model.AppProtection.AlertURLCategory, diags),
			BlockURLCategory:    listToURLCategory(ctx, model.AppProtection.BlockURLCategory, diags),
			AllowURLCategory:    listToURLCategory(ctx, model.AppProtection.AllowURLCategory, diags),
			DefaultURLCategory:  listToURLCategory(ctx, model.AppProtection.DefaultURLCategory, diags),
			UrlDetectedSeverity: model.AppProtection.UrlDetectedSeverity.ValueString(),
			UrlDetectedAction:   model.AppProtection.UrlDetectedAction.ValueString(),
		}
		setStringPresence(&mc.AppProtection.ProfileJSON, "url-detected-action", model.AppProtection.UrlDetectedAction)
		setStringPresence(&mc.AppProtection.ProfileJSON, "url-detected-severity", model.AppProtection.UrlDetectedSeverity)
		if model.AppProtection.MaliciousCodeProtection != nil {
			mc.AppProtection.MaliciousCodeProtection = &airsruntime.MaliciousCodeProtectionConfig{
				Severity: model.AppProtection.MaliciousCodeProtection.Severity.ValueString(),
				Name:     model.AppProtection.MaliciousCodeProtection.Name.ValueString(),
				Action:   model.AppProtection.MaliciousCodeProtection.Action.ValueString(),
			}
			setStringPresence(&mc.AppProtection.MaliciousCodeProtection.ProfileJSON, "name", model.AppProtection.MaliciousCodeProtection.Name)
			setStringPresence(&mc.AppProtection.MaliciousCodeProtection.ProfileJSON, "action", model.AppProtection.MaliciousCodeProtection.Action)
			setStringPresence(&mc.AppProtection.MaliciousCodeProtection.ProfileJSON, "severity", model.AppProtection.MaliciousCodeProtection.Severity)
		}
	}

	for _, mp := range model.ModelProtection {
		sdkMP := airsruntime.ModelProtectionConfig{
			Severity:             mp.Severity.ValueString(),
			SeverityByConfidence: confidenceToSDK(mp.SeverityByConfidence),
			Name:                 mp.Name.ValueString(),
			Action:               airsruntime.ProfileAction(mp.Action.ValueString()),
		}
		setStringPresence(&sdkMP.ProfileJSON, "severity", mp.Severity)
		for _, tc := range mp.ToxicCategories {
			sdkMP.ToxicCategoryList = append(sdkMP.ToxicCategoryList, airsruntime.ToxicCategoryConfig{
				SeverityByConfidence: confidenceToSDK(tc.SeverityByConfidence),
				Category:             tc.Category.ValueString(),
				Action:               tc.Action.ValueString(),
			})
		}
		for _, tl := range mp.TopicLists {
			sdkTL := airsruntime.TopicArrayConfig{
				Action: airsruntime.ProfileAction(tl.Action.ValueString()),
			}
			for _, t := range tl.Topics {
				topic := airsruntime.TopicRef{
					Severity:  t.Severity.ValueString(),
					TopicName: t.TopicName.ValueString(),
					TopicID:   t.TopicID.ValueString(),
					Revision:  t.Revision.ValueInt64(),
				}
				setStringPresence(&topic.ProfileJSON, "severity", t.Severity)
				sdkTL.Topic = append(sdkTL.Topic, topic)
			}
			sdkMP.TopicList = append(sdkMP.TopicList, sdkTL)
		}
		mc.ModelProtection = append(mc.ModelProtection, sdkMP)
	}

	for _, ap := range model.AgentProtection {
		agent := airsruntime.AgentProtectionConfig{
			Severity: ap.Severity.ValueString(),
			Name:     ap.Name.ValueString(),
			Action:   airsruntime.ProfileAction(ap.Action.ValueString()),
		}
		setStringPresence(&agent.ProfileJSON, "severity", ap.Severity)
		mc.AgentProtection = append(mc.AgentProtection, agent)
	}

	return mc
}
func protectionFromSDK(ctx context.Context, mc *airsruntime.ProtectionConfiguration, prior *ProtectionModel, diags *diag.Diagnostics) *ProtectionModel {
	if mc == nil {
		return nil
	}
	model := &ProtectionModel{}
	if mc.DataProtection != nil {
		dpModel := &DataProtectionModel{}

		if mc.DataProtection.DataLeakDetection != nil {
			dld := mc.DataProtection.DataLeakDetection
			dldModel := &DataLeakDetectionModel{
				Action: types.StringValue(string(dld.Action)),
			}
			dldModel.MaskDataInline = types.BoolNull()
			if dld.FieldPresence("mask-data-inline") == airsruntime.JSONPresent {
				dldModel.MaskDataInline = types.BoolValue(dld.MaskDataInline)
			}
			for _, m := range dld.Member {
				member := DataLeakMemberModel{
					Text: types.StringValue(m.Text),
				}
				if m.FieldPresence("id") == airsruntime.JSONPresent {
					member.ID = types.StringValue(m.ID)
				}
				if m.FieldPresence("version") == airsruntime.JSONPresent {
					member.Version = types.StringValue(m.Version)
				}
				dldModel.Members = append(dldModel.Members, member)
			}
			dpModel.DataLeakDetection = dldModel
		}

		for _, ds := range mc.DataProtection.DatabaseSecurity {
			dpModel.DatabaseSecurity = append(dpModel.DatabaseSecurity, DatabaseSecurityModel{
				Severity: sdkString(ds.Severity, ds.FieldPresence("severity")),
				Name:     types.StringValue(ds.Name),
				Action:   types.StringValue(ds.Action),
			})
		}

		if source := mc.DataProtection.SourceCodeDetection; source != nil {
			dpModel.SourceCodeDetection = &SourceCodeDetectionModel{Action: sdkString(string(source.Action), source.FieldPresence("action")), Severity: sdkString(source.Severity, source.FieldPresence("severity"))}
		}
		model.DataProtection = dpModel
	}

	if mc.AppProtection != nil {
		model.AppProtection = &AppProtectionModel{
			AlertURLCategory:    urlCategoryToList(ctx, mc.AppProtection.AlertURLCategory, diags),
			BlockURLCategory:    urlCategoryToList(ctx, mc.AppProtection.BlockURLCategory, diags),
			AllowURLCategory:    urlCategoryToList(ctx, mc.AppProtection.AllowURLCategory, diags),
			DefaultURLCategory:  urlCategoryToList(ctx, mc.AppProtection.DefaultURLCategory, diags),
			UrlDetectedSeverity: sdkString(mc.AppProtection.UrlDetectedSeverity, mc.AppProtection.FieldPresence("url-detected-severity")),
			UrlDetectedAction:   sdkString(mc.AppProtection.UrlDetectedAction, mc.AppProtection.FieldPresence("url-detected-action")),
		}
		if mc.AppProtection.MaliciousCodeProtection != nil {
			model.AppProtection.MaliciousCodeProtection = &MaliciousCodeProtectionModel{
				Severity: sdkString(mc.AppProtection.MaliciousCodeProtection.Severity, mc.AppProtection.MaliciousCodeProtection.FieldPresence("severity")),
				Name:     types.StringValue(mc.AppProtection.MaliciousCodeProtection.Name),
				Action:   types.StringValue(mc.AppProtection.MaliciousCodeProtection.Action),
			}
		}
	}

	for _, mp := range mc.ModelProtection {
		mpModel := ModelProtectionModel{
			Severity:             sdkString(mp.Severity, mp.FieldPresence("severity")),
			SeverityByConfidence: confidenceFromSDK(mp.SeverityByConfidence),
			Name:                 types.StringValue(mp.Name),
			Action:               types.StringValue(string(mp.Action)),
		}
		for _, tc := range mp.ToxicCategoryList {
			mpModel.ToxicCategories = append(mpModel.ToxicCategories, ToxicCategoryModel{
				SeverityByConfidence: confidenceFromSDK(tc.SeverityByConfidence),
				Category:             types.StringValue(tc.Category),
				Action:               types.StringValue(tc.Action),
			})
		}
		for _, tl := range mp.TopicList {
			tlModel := TopicListModel{
				Action: types.StringValue(string(tl.Action)),
			}
			if len(tl.Topic) > 0 {
				for _, t := range tl.Topic {
					tlModel.Topics = append(tlModel.Topics, TopicRefModel{
						Severity:  sdkString(t.Severity, t.FieldPresence("severity")),
						TopicName: types.StringValue(t.TopicName),
						TopicID:   types.StringValue(t.TopicID),
						Revision:  types.Int64Value(t.Revision),
					})
				}
			} else {
				// API may not return topics; preserve from prior state
				if tl.FieldPresence("topic") == airsruntime.JSONOmitted {
					tlModel.Topics = priorTopics(prior, mp.Name, string(tl.Action))
				}
			}
			mpModel.TopicLists = append(mpModel.TopicLists, tlModel)
		}
		model.ModelProtection = append(model.ModelProtection, mpModel)
	}

	for _, ap := range mc.AgentProtection {
		model.AgentProtection = append(model.AgentProtection, AgentProtectionModel{
			Severity: sdkString(ap.Severity, ap.FieldPresence("severity")),
			Name:     types.StringValue(ap.Name),
			Action:   types.StringValue(string(ap.Action)),
		})
	}
	return model
}

func setBoolPresence(p *airsruntime.ProfileJSON, name string, value types.Bool) {
	if value.IsNull() || value.IsUnknown() {
		p.SetFieldPresence(name, airsruntime.JSONOmitted)
	} else {
		p.SetFieldPresence(name, airsruntime.JSONPresent)
	}
}
func setStringPresence(p *airsruntime.ProfileJSON, name string, value types.String) {
	if value.IsNull() || value.IsUnknown() {
		p.SetFieldPresence(name, airsruntime.JSONOmitted)
	} else {
		p.SetFieldPresence(name, airsruntime.JSONPresent)
	}
}
func sdkString(value string, presence airsruntime.JSONPresence) types.String {
	if presence == airsruntime.JSONOmitted {
		return types.StringNull()
	}
	return types.StringValue(value)
}
func confidenceToSDK(m *SeverityByConfidenceModel) *airsruntime.SeverityByConfidence {
	if m == nil {
		return nil
	}
	v := &airsruntime.SeverityByConfidence{High: m.High.ValueString(), Moderate: m.Moderate.ValueString()}
	setStringPresence(&v.ProfileJSON, "high", m.High)
	setStringPresence(&v.ProfileJSON, "moderate", m.Moderate)
	return v
}
func confidenceFromSDK(m *airsruntime.SeverityByConfidence) *SeverityByConfidenceModel {
	if m == nil {
		return nil
	}
	return &SeverityByConfidenceModel{High: sdkString(m.High, m.FieldPresence("high")), Moderate: sdkString(m.Moderate, m.FieldPresence("moderate"))}
}
func directionsToSDK(ctx context.Context, m *ContentTypeConfigurationsModel, diags *diag.Diagnostics) *airsruntime.ContentTypeConfigurations {
	if m == nil {
		return nil
	}
	return &airsruntime.ContentTypeConfigurations{
		Prompt: protectionToSDK(ctx, m.Prompt, diags), Response: protectionToSDK(ctx, m.Response, diags),
		ToolCall: protectionToSDK(ctx, m.ToolCall, diags), ToolResponse: protectionToSDK(ctx, m.ToolResponse, diags),
	}
}
func directionsFromSDK(ctx context.Context, m *airsruntime.ContentTypeConfigurations, prior *ContentTypeConfigurationsModel, diags *diag.Diagnostics) *ContentTypeConfigurationsModel {
	if m == nil {
		return nil
	}
	if prior == nil {
		prior = &ContentTypeConfigurationsModel{}
	}
	return &ContentTypeConfigurationsModel{
		Prompt: protectionFromSDK(ctx, m.Prompt, prior.Prompt, diags), Response: protectionFromSDK(ctx, m.Response, prior.Response, diags),
		ToolCall: protectionFromSDK(ctx, m.ToolCall, prior.ToolCall, diags), ToolResponse: protectionFromSDK(ctx, m.ToolResponse, prior.ToolResponse, diags),
	}
}
func priorProfile(prior []AiSecurityProfileModel, profile airsruntime.AiSecurityProfileConfig) *AiSecurityProfileModel {
	var found *AiSecurityProfileModel
	for i := range prior {
		if prior[i].ModelType.ValueString() == profile.ModelType && prior[i].ContentType.ValueString() == profile.ContentType {
			if found != nil {
				return nil
			} // Ambiguous entries must never share references.
			found = &prior[i]
		}
	}
	return found
}
func priorProfileProtection(prior []AiSecurityProfileModel, profile airsruntime.AiSecurityProfileConfig) *ProtectionModel {
	if p := priorProfile(prior, profile); p != nil {
		return p.protection()
	}
	return nil
}
func priorDirections(prior []AiSecurityProfileModel, profile airsruntime.AiSecurityProfileConfig) *ContentTypeConfigurationsModel {
	if p := priorProfile(prior, profile); p != nil {
		return p.ContentTypeConfigurations
	}
	return nil
}
func priorTopics(prior *ProtectionModel, detector, action string) []TopicRefModel {
	if prior == nil {
		return nil
	}
	var found []TopicRefModel
	matches := 0
	for _, mp := range prior.ModelProtection {
		if mp.Name.ValueString() != detector {
			continue
		}
		for _, tl := range mp.TopicLists {
			if tl.Action.ValueString() == action {
				found = tl.Topics
				matches++
			}
		}
	}
	if matches != 1 {
		return nil
	}
	for _, t := range found {
		if t.TopicID.IsUnknown() || t.Revision.IsUnknown() {
			return nil
		}
	}
	return found
}

// Every protection layout participates in the same topic resolution pass.
func profileProtectionModels(plan *SecurityProfileResourceModel) []*ProtectionModel {
	var models []*ProtectionModel
	for i := range plan.AiSecurityProfiles {
		asp := &plan.AiSecurityProfiles[i]
		models = append(models, asp.protection())
		if d := asp.ContentTypeConfigurations; d != nil {
			for _, p := range []*ProtectionModel{d.Prompt, d.Response, d.ToolCall, d.ToolResponse} {
				if p != nil {
					models = append(models, p)
				}
			}
		}
	}
	return models
}
