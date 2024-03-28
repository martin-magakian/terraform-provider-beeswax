package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martin-magakian/terraform-provider-beeswax/internal/beeswax-client"
)

func defaultConfiguration(providerData any, diagnostics diag.Diagnostics) *beeswax.Client {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*beeswax.Client)
	if !ok {
		diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *beeswax.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}
	return client
}

func convertListInt(list []types.Int64) []int64 {
	result := []int64{}
	for _, item := range list {
		result = append(result, item.ValueInt64())
	}
	return result
}
