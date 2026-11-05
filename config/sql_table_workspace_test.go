package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/databricks/databricks-sdk-go"
	sdkclient "github.com/databricks/databricks-sdk-go/client"
	sdkconfig "github.com/databricks/databricks-sdk-go/config"
	tfcommon "github.com/databricks/terraform-provider-databricks/common"
	"github.com/databricks/terraform-provider-databricks/xpprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const sqlTableJSON = `{
  "name": "deliveries", "catalog_name": "tenant", "schema_name": "courier",
  "table_type": "MANAGED", "data_source_format": "DELTA",
  "comment": "delivered messages",
  "properties": {"delta.enableChangeDataFeed": "true"},
  "columns": [
    {"name": "id", "type_text": "string", "position": 0, "nullable": false, "comment": "delivery id"},
    {"name": "at", "type_text": "timestamp", "position": 1, "nullable": true}
  ]
}`

type recorder struct {
	mu    sync.Mutex
	paths []string
}

func (r *recorder) add(p string) { r.mu.Lock(); r.paths = append(r.paths, p); r.mu.Unlock() }
func (r *recorder) tableReads() (n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.paths {
		if strings.Contains(p, "/unity-catalog/tables/") {
			n++
		}
	}
	return n
}

// TestSQLTableReadHonoursProviderConfigWorkspaceID reproduces the live failure
// on the tenant-e2e SQLTables: the provider is account-level and the resource
// sets provider_config.workspace_id. The table read must go to the workspace
// host. Unpatched, databricks_sql_table does a raw GET on the provider's own
// host, here the account host, which answers with HTML, as the account console
// does.
func TestSQLTableReadHonoursProviderConfigWorkspaceID(t *testing.T) {
	ctx := context.Background()
	const wsID = "7474655817775321"

	account, workspace := &recorder{}, &recorder{}
	accountSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account.add(r.URL.Path)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<!doctype html><html>account console</html>"))
	}))
	t.Cleanup(accountSrv.Close)
	workspaceSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspace.add(r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/unity-catalog/tables/tenant.courier.deliveries") {
			_, _ = w.Write([]byte(sqlTableJSON))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error_code":"RESOURCE_DOES_NOT_EXIST","message":"nope"}`))
	}))
	t.Cleanup(workspaceSrv.Close)

	fw, sdk, err := xpprovider.GetProvider(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Applies the fork's resource configurators, including sql_table's.
	if _, err := GetProvider(ctx, fw, sdk, false); err != nil {
		t.Fatal(err)
	}
	res := sdk.ResourcesMap["databricks_sql_table"]

	accountHost := "https://accounts.localhost:" + strings.Split(accountSrv.URL, ":")[2]
	cfg := &sdkconfig.Config{Host: accountHost, AccountID: "acc", Token: "t", InsecureSkipVerify: true}
	low, err := sdkclient.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	client := &tfcommon.DatabricksClient{DatabricksClient: low}
	// What the provider resolves for provider_config.workspace_id on an
	// account host: a client for that workspace's own host.
	ws, err := databricks.NewWorkspaceClient(&databricks.Config{Host: workspaceSrv.URL, Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	client.SetWorkspaceClientForWorkspace(wsID, ws)

	// The attributes a SQLTable's forProvider sets; Read must fill in the rest.
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]any{
		"name":            "deliveries",
		"catalog_name":    "tenant",
		"schema_name":     "courier",
		"table_type":      "MANAGED",
		"comment":         "delivered messages",
		"provider_config": []any{map[string]any{"workspace_id": wsID}},
		"column": []any{
			map[string]any{"name": "id", "type": "string", "nullable": false, "comment": "delivery id"},
			map[string]any{"name": "at", "type": "timestamp", "nullable": true},
		},
	})
	d.SetId("tenant.courier.deliveries")
	if diags := res.ReadContext(ctx, d, client); diags.HasError() {
		t.Fatalf("read failed: %v", diags)
	}

	if n := account.tableReads(); n != 0 {
		t.Errorf("table read went to the account host %d time(s), want 0 (paths: %v)", n, account.paths)
	}
	if n := workspace.tableReads(); n == 0 {
		t.Errorf("table was never read from the workspace host (paths: %v)", workspace.paths)
	}
	assertSQLTableState(t, d)
}

func assertSQLTableState(t *testing.T, d *schema.ResourceData) {
	t.Helper()
	if got := d.Get("comment"); got != "delivered messages" {
		t.Errorf("comment = %v", got)
	}
	if got := d.Get("table_type"); got != "MANAGED" {
		t.Errorf("table_type = %v", got)
	}
	if got := d.Get("column.#"); got != 2 {
		t.Fatalf("column count = %v, want 2", got)
	}
	if got := d.Get("column.0.name"); got != "id" {
		t.Errorf("column.0.name = %v", got)
	}
	if got := d.Get("column.0.type"); got != "string" {
		t.Errorf("column.0.type = %v", got)
	}
	if got := d.Get("column.0.comment"); got != "delivery id" {
		t.Errorf("column.0.comment = %v", got)
	}
	if got := d.Get("column.1.nullable"); got != true {
		t.Errorf("column.1.nullable = %v", got)
	}
	props, _ := d.Get("effective_properties").(map[string]any)
	if props["delta.enableChangeDataFeed"] != "true" {
		t.Errorf("effective_properties = %v", props)
	}
}
