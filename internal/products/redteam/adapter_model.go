package redteam

import (
	"context"
	"encoding/base64"
	"sort"

	"github.com/cdot65/prisma-airs-go/aisec"
	rtschema "github.com/cdot65/prisma-airs-go/aisec/redteam/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var adapterVariableTypes = map[string]attr.Type{"type": types.StringType, "value": types.StringType}

type adapterModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Script                   types.String `tfsdk:"script"`
	NetworkBrokerChannelUUID types.String `tfsdk:"network_broker_channel_uuid"`
	Variables                types.Map    `tfsdk:"variables"`
	Validate                 types.Bool   `tfsdk:"validate"`
	ValidationPrompt         types.String `tfsdk:"validation_prompt"`
	Status                   types.String `tfsdk:"status"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

func adapterVariables(plan types.Map, prior types.Map, diags *diag.Diagnostics) []rtschema.AdapterVarBase {
	result := []rtschema.AdapterVarBase{}
	if plan.IsUnknown() {
		diags.AddError("Unknown adapter variables", "Variables must be known before writing.")
		return result
	}
	keys := []string{}
	for key := range plan.Elements() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		candidate := plan.Elements()[key]
		if candidate.IsNull() || candidate.IsUnknown() {
			diags.AddError("Invalid adapter variable", "Every variable requires a known type and a value or a preserved secret.")
			continue
		}
		fields := candidate.(types.Object).Attributes()
		kind := fields["type"].(types.String)
		value := fields["value"].(types.String)
		if kind.IsNull() || kind.IsUnknown() || (kind.ValueString() != "VAR" && kind.ValueString() != "SECRET") || value.IsUnknown() {
			diags.AddError("Invalid adapter variable", "Every variable requires VAR or SECRET and a known value.")
			continue
		}
		wire := rtschema.AdapterVarBase{Key: key, Type: rtschema.AdapterVarType(kind.ValueString()), Value: aisec.Null[string]()}
		if value.IsNull() {
			before, exists := prior.Elements()[key]
			if kind.ValueString() != "SECRET" || !exists || before.IsNull() || before.IsUnknown() || !before.(types.Object).Attributes()["type"].Equal(kind) {
				diags.AddError("Missing adapter variable value", "Null can preserve only an existing SECRET with the same key and type. Supply values for new variables.")
				continue
			}
		} else if targetContainsMask(value) {
			diags.AddError("Masked adapter secret", "A redacted placeholder is not a credential. Use null for an unchanged imported SECRET or supply its original value.")
			continue
		} else {
			wire.Value = aisec.Value(value.ValueString())
		}
		result = append(result, wire)
	}
	return result
}

func adapterRequest(plan *adapterModel, prior *adapterModel, diags *diag.Diagnostics) rtschema.CustomTargetAdapterCreateRequest {
	before := types.MapNull(types.ObjectType{AttrTypes: adapterVariableTypes})
	if prior != nil {
		before = prior.Variables
	}
	variables := adapterVariables(plan.Variables, before, diags)
	if plan.Validate.ValueBool() && (plan.NetworkBrokerChannelUUID.IsNull() || plan.NetworkBrokerChannelUUID.IsUnknown()) {
		diags.AddError("Adapter execution requires Network Broker", "validate = true executes the script. Supply an existing online network_broker_channel_uuid.")
	}
	return rtschema.CustomTargetAdapterCreateRequest{Name: plan.Name.ValueString(), Description: aisec.Value(plan.Description.ValueString()), ScriptB64: base64.StdEncoding.EncodeToString([]byte(plan.Script.ValueString())), NetworkBrokerChannelUUID: adapterOptionalString(plan.NetworkBrokerChannelUUID), Variables: &variables, Prompt: plan.ValidationPrompt.ValueString()}
}

func adapterOptionalString(v types.String) aisec.Optional[string] {
	if v.IsNull() {
		return aisec.Null[string]()
	}
	return aisec.Value(v.ValueString())
}

func mapAdapter(ctx context.Context, remote *rtschema.CustomTargetAdapter, state *adapterModel, diags *diag.Diagnostics) {
	script, err := base64.StdEncoding.DecodeString(remote.ScriptB64)
	if err != nil {
		diags.AddError("Invalid adapter script", "The service returned an invalid base64 script.")
		return
	}
	state.ID = types.StringValue(remote.UUID)
	state.Name = types.StringValue(remote.Name)
	description, _ := remote.Description.Get()
	state.Description = types.StringValue(description)
	state.Script = types.StringValue(string(script))
	state.Status = types.StringValue(string(remote.Status))
	state.NetworkBrokerChannelUUID = types.StringNull()
	if channel, ok := remote.NetworkBrokerChannelUUID.Get(); ok {
		state.NetworkBrokerChannelUUID = types.StringValue(channel)
	}
	state.CreatedAt = types.StringNull()
	state.UpdatedAt = types.StringNull()
	if value, ok := remote.CreatedAt.Get(); ok {
		state.CreatedAt = types.StringValue(value)
	}
	if value, ok := remote.UpdatedAt.Get(); ok {
		state.UpdatedAt = types.StringValue(value)
	}
	mapAdapterVariables(ctx, remote, state, diags)
}

func mapAdapterVariables(ctx context.Context, remote *rtschema.CustomTargetAdapter, state *adapterModel, diags *diag.Diagnostics) {
	values := map[string]attr.Value{}
	if remote.Variables != nil {
		for _, variable := range *remote.Variables {
			if variable.Key == "" {
				diags.AddError("Invalid adapter variable identity", "The service returned an empty variable key.")
				return
			}
			if _, duplicate := values[variable.Key]; duplicate {
				diags.AddError("Invalid adapter variable identity", "The service returned duplicate variable keys.")
				return
			}
			value := types.StringNull()
			redacted := variable.IsRedacted != nil && *variable.IsRedacted
			if actual, ok := variable.Value.Get(); ok && !redacted {
				value = types.StringValue(actual)
				if targetContainsMask(value) {
					value = types.StringNull()
				}
			}
			if variable.Type == rtschema.AdapterVarTypeSecret && (redacted || value.IsNull()) {
				if prior, exists := state.Variables.Elements()[variable.Key]; exists && !prior.IsNull() && !prior.IsUnknown() {
					attrs := prior.(types.Object).Attributes()
					if attrs["type"].Equal(types.StringValue("SECRET")) {
						value = attrs["value"].(types.String)
					}
				}
			}
			values[variable.Key] = types.ObjectValueMust(adapterVariableTypes, map[string]attr.Value{"type": types.StringValue(string(variable.Type)), "value": value})
		}
	}
	mapped, d := types.MapValue(types.ObjectType{AttrTypes: adapterVariableTypes}, values)
	diags.Append(d...)
	state.Variables = mapped
}
