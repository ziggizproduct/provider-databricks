package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/controller"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/databricks/terraform-provider-databricks/xpprovider"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	billingv1beta1 "github.com/glalanne/provider-databricks/apis/cluster/billing/v1beta1"
)

// TestBudgetPolicyObserve drives the real upjet framework Observe/Create path
// against the real databricks_budget_policy resource, with the account API
// served by a fake. It reproduces tenant-e2e run 39074653769: a new
// BudgetPolicy with no external-name must be reported as not existing so that
// Create runs, rather than failing with `cannot find "policy_id" in tfstate`.
func TestBudgetPolicyObserve(t *testing.T) {
	ctx := context.Background()
	const wantID = "11111111-2222-3333-4444-555555555555"

	var mu sync.Mutex
	var calls []string
	encode := func(w http.ResponseWriter, v map[string]string) {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("encode fake response: %v", err)
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		p := strings.TrimSuffix(r.URL.Path, "/")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(p, "/budget-policies"):
			encode(w, map[string]string{"policy_id": wantID, "policy_name": "p"})
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/budget-policies/"+wantID):
			encode(w, map[string]string{"policy_id": wantID, "policy_name": "p"})
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/budget-policies"):
			// GET with an empty policy id collapses to the list endpoint:
			// 200 and no policy_id, which is what makes the live Read succeed.
			encode(w, map[string]string{})
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error_code":"RESOURCE_DOES_NOT_EXIST","message":"nope"}`))
		}
	})
	// The SDK only treats https://accounts.* hosts as account hosts, so serve
	// TLS on loopback and reach it through accounts.localhost.
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	host := "https://accounts.localhost:" + strings.Split(srv.URL, ":")[2]

	fw, sdk, err := xpprovider.GetProvider(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pc, err := GetProvider(ctx, fw, sdk, false)
	if err != nil {
		t.Fatal(err)
	}
	cfg := pc.Resources["databricks_budget_policy"]
	setup := func(context.Context, client.Client, xpresource.Managed) (terraform.Setup, error) {
		return terraform.Setup{
			FrameworkProvider: fw,
			Configuration: map[string]any{
				"host":        host,
				"account_id":  "acc",
				"token":       "t",
				"skip_verify": true,
			},
		}, nil
	}
	newClient := func(mg *billingv1beta1.BudgetPolicy) managed.ExternalClient {
		c := controller.NewTerraformPluginFrameworkConnector(fake.NewClientBuilder().Build(), setup, cfg,
			controller.NewOperationStore(logging.NewNopLogger()),
			controller.WithTerraformPluginFrameworkLogger(logging.NewNopLogger()))
		ec, err := c.Connect(ctx, mg)
		if err != nil {
			t.Fatalf("connect: %v", err)
		}
		return ec
	}
	mg := &billingv1beta1.BudgetPolicy{}
	mg.SetName("tenant-e2e-c-none")
	mg.Spec.ForProvider.PolicyName = ptr.To("tenant-e2e-c-none")
	mg.Spec.ForProvider.BindingWorkspaceIds = []*float64{ptr.To(7474655817775321.0)}

	// First reconcile: empty external-name.
	ec := newClient(mg)
	obs, err := ec.Observe(ctx, mg)
	if err != nil {
		t.Fatalf("first observe must not error: %v", err)
	}
	if obs.ResourceExists {
		t.Fatal("first observe with empty external-name must report ResourceExists=false")
	}
	if _, err := ec.Create(ctx, mg); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := meta.GetExternalName(mg); got != wantID {
		t.Fatalf("external-name after create = %q, want %q", got, wantID)
	}

	// Pod restart: fresh tracker store, external-name now on the annotation.
	ec = newClient(mg)
	obs, err = ec.Observe(ctx, mg)
	if err != nil {
		t.Fatalf("observe after restart: %v", err)
	}
	if !obs.ResourceExists {
		t.Fatal("observe after restart must report ResourceExists=true")
	}
	t.Logf("api calls: %v", calls)
}
