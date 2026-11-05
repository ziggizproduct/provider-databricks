package common

import (
	"context"
	"strconv"
	"sync"

	"github.com/databricks/databricks-sdk-go/config"
	tfcommon "github.com/databricks/terraform-provider-databricks/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EXIT: remove this file, and its callers in config/{cluster,namespaced}/sql_table,
// when terraform-provider-databricks reads databricks_sql_table through the
// unified-provider client (see catalog/resource_sql_table.go: Read and Update
// call SqlTablesAPI.getTable with the provider client, whose host is the
// account host, instead of c.DatabricksClientForUnifiedProvider(ctx, d)).

var scopedResources sync.Map // *schema.Resource -> struct{}; the SDK provider is shared by the cluster and namespaced configs

// WorkspaceScopedSDKv2Resource makes the Create, Read and Update of an SDKv2
// resource run against a client for the workspace named in
// provider_config.workspace_id when the provider is configured at account
// level. Upstream databricks_sql_table does a raw GET against the provider's
// own host, which for an account ProviderConfig is the account console and
// answers with HTML. Create already uses the unified-provider client, but the
// Read it runs after creating does not, so the table is created and the
// resource is then reported as failed. Wrapping is idempotent.
func WorkspaceScopedSDKv2Resource(r *schema.Resource) {
	if r == nil {
		return
	}
	if _, done := scopedResources.LoadOrStore(r, struct{}{}); done {
		return
	}
	wrap := func(f func(context.Context, *schema.ResourceData, any) diag.Diagnostics) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
		if f == nil {
			return nil
		}
		return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			wc, err := workspaceScopedClient(ctx, d, meta)
			if err != nil {
				return diag.FromErr(err)
			}
			return f(ctx, d, wc)
		}
	}
	r.CreateContext = wrap(r.CreateContext)
	r.ReadContext = wrap(r.ReadContext)
	r.UpdateContext = wrap(r.UpdateContext)
}

// workspaceScopedClient returns meta unchanged unless the provider is
// account-level, in which case it returns the client for the resource's
// workspace.
func workspaceScopedClient(ctx context.Context, d *schema.ResourceData, meta any) (any, error) {
	c, ok := meta.(*tfcommon.DatabricksClient)
	// A workspace-level provider is already scoped to its workspace.
	if !ok || c == nil || c.Config == nil || c.HostTypeForTerraform() == config.WorkspaceHost {
		return meta, nil
	}
	wc, err := c.DatabricksClientForUnifiedProvider(ctx, d)
	if err != nil {
		return nil, err
	}
	if wc == c {
		return meta, nil
	}
	// The client is derived from this workspace, so skip the /Me round trip
	// that would otherwise confirm the id.
	id, _ := d.Get("provider_config.0.workspace_id").(string)
	if id == "" {
		id = c.Config.WorkspaceID
	}
	if n, perr := strconv.ParseInt(id, 10, 64); perr == nil {
		wc.SetCachedWorkspaceID(n)
	}
	return wc, nil
}
