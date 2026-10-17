package common

import (
	"context"
	"fmt"

	"github.com/crossplane/upjet/v2/pkg/config"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// IdentifierArgumentExternalName is an external-name configuration for
// plugin framework resources whose only identifier is the computed attribute
// field and which have no "id" attribute. The external name is the value of
// field. Because upjet only reaches the Terraform prior state through
// SetIdentifierArgumentFn when there is no "id", it also writes the external
// name back to field; without that, Read runs with no identifier and the state
// comes back empty.
func IdentifierArgumentExternalName(field string) config.ExternalName {
	en := config.NewExternalNameFrom(
		config.IdentifierFromProvider,
		config.WithSetIdentifierArgumentsFn(func(_ config.SetIdentifierArgumentsFn, base map[string]any, externalName string) {
			if externalName != "" {
				base[field] = externalName
			}
		}),
		config.WithGetExternalNameFn(func(_ config.GetExternalNameFn, tfState map[string]any) (string, error) {
			value, ok := tfState[field]
			if !ok || value == nil || fmt.Sprint(value) == "" {
				return "", fmt.Errorf("cannot find %q in tfstate", field)
			}
			return fmt.Sprint(value), nil
		}),
	)
	// field is optional+computed, so keep it out of the config sent to
	// Terraform; it is still set on the prior state via SetIdentifierArgumentFn.
	en.TFPluginFrameworkOptions.ComputedIdentifierAttributes = []string{field}
	return en
}

var BudgetPolicyExternalName = IdentifierArgumentExternalName("policy_id")

// StateEmptyWithoutIdentifier is a TerraformPluginFrameworkIsStateEmptyFn for
// resources whose only identifier is the computed attribute field. On the first
// observe the external name is empty, so the prior state is reconstructed from
// spec.forProvider without field. The provider Read is then called with no
// identifier and, instead of reporting not-found, can return a non-null state
// that still has no field. upjet would treat that as an existing resource and
// fail computing the external name; a state without field cannot be a real
// resource, so report it as empty and let Create run.
func StateEmptyWithoutIdentifier(field string) config.TerraformPluginFrameworkIsStateEmptyFn {
	return func(_ context.Context, state tftypes.Value, _ rschema.Schema) (bool, error) {
		attrs := map[string]tftypes.Value{}
		if err := state.As(&attrs); err != nil {
			return false, err
		}
		v, ok := attrs[field]
		if !ok || v.IsNull() || !v.IsKnown() {
			return true, nil
		}
		var s string
		if err := v.As(&s); err != nil {
			return false, err
		}
		return s == "", nil
	}
}
