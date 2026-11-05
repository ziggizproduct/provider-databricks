package sql_table

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/glalanne/provider-databricks/config/common"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("databricks_sql_table", func(r *config.Resource) {
		r.ShortGroup = "unity"
		// EXIT: remove when upstream reads sql_table through the unified-provider client.
		common.WorkspaceScopedSDKv2Resource(r.TerraformResource)
		r.References["cluster_id"] = config.Reference{
			TerraformName: "databricks_cluster",
		}
		r.References["warehouse_id"] = config.Reference{
			TerraformName: "databricks_sql_endpoint",
		}
	})
}
