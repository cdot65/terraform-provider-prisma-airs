package redteam

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var targetFamilies = []string{"openai", "hugging_face", "databricks", "bedrock", "custom", "rest", "streaming", "adapter"}
var targetAuthBlocks = []string{"headers_auth", "basic_auth", "oauth2_auth"}

func targetString(sensitive bool, description string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Sensitive: sensitive, Description: description,
		Validators: []validator.String{stringvalidator.LengthAtLeast(1)}}
}

func targetBlocks() map[string]schema.Block {
	blocks := map[string]schema.Block{}
	for _, name := range targetFamilies {
		attrs := map[string]schema.Attribute{}
		switch name {
		case "adapter":
			attrs["uuid"] = targetString(false, "Existing or Terraform-managed adapter UUID.")
			attrs["variable_overrides"] = schema.MapNestedAttribute{Optional: true, Sensitive: true, Description: "Complete target override key set. New overrides require values; redacted imported overrides cannot be written until supplied.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"type":  schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("VAR", "SECRET")}, Description: "VAR or SECRET."},
				"value": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Override value. Unavailable imported SECRET values remain null for read-only adoption."},
			}}}
		case "openai", "hugging_face":
			attrs["api_key"] = targetString(true, "Provider credential. Unavailable on import; supply the desired value.")
			attrs["model_name"] = targetString(false, "Provider model name.")
		case "bedrock":
			for _, field := range []string{"access_id", "access_secret", "region", "model_id"} {
				attrs[field] = targetString(strings.HasPrefix(field, "access_"), "AWS Bedrock "+field+".")
			}
			attrs["session_token"] = targetString(true, "Optional AWS session token.")
		case "databricks":
			attrs["response_stop_key"] = schema.StringAttribute{Optional: true, Description: "Databricks streaming completion field. Imported empty values permit read-only adoption; writes require a nonempty value."}
			attrs["response_stop_value"] = schema.StringAttribute{Optional: true, Description: "Databricks streaming completion value. Imported empty values permit read-only adoption; writes require a nonempty value."}
			attrs["workspace_url"] = targetString(false, "Databricks workspace URL.")
			attrs["model_name"] = targetString(false, "Serving model name.")
			for _, field := range []string{"access_token", "client_id", "secret"} {
				attrs[field] = targetString(true, "Use access_token, or client_id and secret for OAuth; never combine them.")
			}
		default:
			attrs["api_endpoint"] = targetString(false, "Target API URL; management does not execute inference.")
			attrs["request_headers"] = schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, Description: "Desired request headers. Put credentials in an authentication block."}
			attrs["request_body"] = schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Desired native HCL request object. Nested lists and nulls are supported; JSON serialization is internal."}
			attrs["response_body"] = schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Desired native HCL response object. Read-back cannot reliably detect changes inside this payload."}
			attrs["response_key"] = targetString(false, "Path to the model response.")
			if name == "streaming" {
				attrs["response_stop_key"] = schema.StringAttribute{Optional: true, Description: "Streaming completion field. Imported empty values permit read-only adoption; writes require a nonempty value."}
				attrs["response_stop_value"] = schema.StringAttribute{Optional: true, Description: "Streaming completion value. Imported empty values permit read-only adoption; writes require a nonempty value."}
			}
		}
		if name == "openai" || name == "hugging_face" || name == "databricks" || name == "bedrock" {
			attrs["api_endpoint"] = targetString(false, "Optional provider endpoint override.")
			attrs["request_body"] = schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Native HCL request object containing {INPUT}; required for text targets."}
			attrs["response_body"] = schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Desired native HCL response object."}
			attrs["response_key"] = targetString(false, "Desired response path. Native provider read-back may omit this value; unavailable on import.")
			attrs["request_headers"] = schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, Description: "Desired provider request headers."}
		}

		blocks[name] = schema.SingleNestedBlock{Attributes: attrs, Description: "Native " + strings.ToUpper(name) + " connection. Exactly one connection block is required."}
	}
	blocks["headers_auth"] = schema.SingleNestedBlock{Description: "Static authentication headers.", Attributes: map[string]schema.Attribute{
		"headers": schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, Description: "Desired secret header values; unavailable on import."},
	}}
	blocks["basic_auth"] = schema.SingleNestedBlock{Description: "HTTP Basic authentication.", Attributes: map[string]schema.Attribute{
		"username": targetString(true, "Basic authentication username."),
		"password": targetString(true, "Basic authentication password; unavailable on import."),
		"location": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("HEADER"), Description: "Credential location.", Validators: []validator.String{stringvalidator.OneOf("HEADER", "PAYLOAD")}},
	}}
	blocks["oauth2_auth"] = schema.SingleNestedBlock{Description: "OAuth2 authentication.", Attributes: map[string]schema.Attribute{
		"token_url":      targetString(false, "Token endpoint URL."),
		"headers":        schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Desired native HCL token request headers."},
		"body":           schema.DynamicAttribute{Optional: true, Sensitive: true, Description: "Desired native HCL token request body."},
		"expiry_minutes": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(60), Validators: []validator.Int64{int64validator.AtLeast(0)}, Description: "Token validity in minutes; defaults to 60."},
		"response_key":   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("access_token"), Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, Description: "Token response path; defaults to access_token."},
		"inject_header":  schema.MapAttribute{Optional: true, Sensitive: true, ElementType: types.StringType, Description: "Header templates containing {TOKEN}."},
	}}
	return blocks
}

func targetObjectTypes(name string) map[string]attr.Type {
	return targetBlocks()[name].(schema.SingleNestedBlock).Type().(types.ObjectType).AttrTypes
}
