package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"hashicups_coffees":     dataSourceCoffees(),
			"hashicups_ingredients": dataSourceIngredients(),
		},
	}
}
