package group_member

import (
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("databricks_group_member", func(r *config.Resource) {
		r.ShortGroup = "security"
		r.References["group_id"] = config.Reference{
			TerraformName: "databricks_group",
		}
		r.TerraformCustomDiff = dropComputedProviderConfigDiff
	})
}

// dropComputedProviderConfigDiff removes the provider_config diff that an
// account-level group member always shows after creation.
//
// databricks_group_member is a dual (account and workspace) resource with an
// Optional+Computed provider_config block, but the Terraform provider skips
// populating provider_config in state when the resource is at account level
// (common/resource.go in terraform-provider-databricks). The plan then reports
// provider_config.# as NewComputed, an in-place change, and the resource has no
// Update, so every reconcile fails with "doesn't support update". A change that
// requires replacement still passes through: at workspace level the provider's
// CustomizeDiffDualResources marks a provider_config.workspace_id switch
// ForceNew, and at account level provider_config carries nothing. Every other
// attribute of this resource is ForceNew or Computed, so no real in-place
// change can be hidden.
//
// Exit: remove when terraform-provider-databricks fills provider_config for
// account-level dual resources or gives databricks_group_member an Update.
// databricks_group_role and databricks_group_instance_profile have the same
// shape and need this diff if they are ever composed at account level.
func dropComputedProviderConfigDiff(
	diff *terraform.InstanceDiff,
	state *terraform.InstanceState,
	_ *terraform.ResourceConfig,
) (*terraform.InstanceDiff, error) {
	if diff == nil || diff.Destroy || state == nil || state.ID == "" {
		return diff, nil
	}
	for k, d := range diff.Attributes {
		if (k == "provider_config.#" || strings.HasPrefix(k, "provider_config.")) && (d == nil || !d.RequiresNew) {
			delete(diff.Attributes, k)
		}
	}
	return diff, nil
}
