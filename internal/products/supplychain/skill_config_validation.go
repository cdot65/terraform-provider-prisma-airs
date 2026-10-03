package supplychain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithValidateConfig = &skillDataSource{}

func (d *skillDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if d.kind != "scans" {
		return
	}
	var config types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	v := config.Attributes()
	if v["start_time"].IsUnknown() || v["end_time"].IsUnknown() {
		return
	}
	if err := validateTimeRange(skillString(v, "start_time"), skillString(v, "end_time")); err != nil {
		resp.Diagnostics.AddError("Invalid Skill Scanning time range", err.Error())
	}
}

var _ resource.ResourceWithValidateConfig = &skillResource{}

func (r *skillResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.kind != "instance" {
		return
	}
	var config types.Object
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value := config.Attributes()["registration_details"]
	if skillHasUnknown(value) {
		return
	}
	if _, e := skillRegistration(value); e != nil {
		resp.Diagnostics.AddAttributeError(path.Root("registration_details"), "Invalid native registration", skillError(e))
	}
}
