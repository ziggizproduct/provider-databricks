package group_member

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDropComputedProviderConfigDiff(t *testing.T) {
	state := &terraform.InstanceState{ID: "84349292178449|5895513711777485"}

	t.Run("computed provider_config on an existing member is dropped", func(t *testing.T) {
		diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{
			"provider_config.#": {NewComputed: true},
		}}
		got, err := dropComputedProviderConfigDiff(diff, state, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Attributes) != 0 {
			t.Fatalf("want an empty diff, got %v", got.Attributes)
		}
	})

	t.Run("a replacement still passes through", func(t *testing.T) {
		diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{
			"provider_config.#":              {NewComputed: true},
			"provider_config.0.workspace_id": {Old: "1", New: "2", RequiresNew: true},
			"member_id":                      {Old: "1", New: "2", RequiresNew: true},
		}}
		got, _ := dropComputedProviderConfigDiff(diff, state, nil)
		if _, ok := got.Attributes["provider_config.#"]; ok {
			t.Fatal("provider_config.# should be dropped")
		}
		for _, k := range []string{"provider_config.0.workspace_id", "member_id"} {
			if _, ok := got.Attributes[k]; !ok {
				t.Fatalf("%s requires replacement and must stay", k)
			}
		}
	})

	t.Run("create and destroy are untouched", func(t *testing.T) {
		create := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{"provider_config.#": {NewComputed: true}}}
		if got, _ := dropComputedProviderConfigDiff(create, &terraform.InstanceState{}, nil); len(got.Attributes) != 1 {
			t.Fatal("a create diff must be kept")
		}
		destroy := &terraform.InstanceDiff{Destroy: true, Attributes: map[string]*terraform.ResourceAttrDiff{"provider_config.#": {NewComputed: true}}}
		if got, _ := dropComputedProviderConfigDiff(destroy, state, nil); len(got.Attributes) != 1 {
			t.Fatal("a destroy diff must be kept")
		}
	})
}
