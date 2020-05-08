package hashicups

import (
	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HASHICUPS_USERNAME", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HASHICUPS_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hashicups_order": resourceOrder(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hashicups_coffees":     dataSourceCoffees(),
			"hashicups_ingredients": dataSourceIngredients(),
			"hashicups_order":       dataSourceOrder(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	if (username != "") && (password != "") {
		c, err := hashicups.NewClient(nil, &username, &password)
		if err != nil {
			return nil, err
		}

		return c, nil
	}

	c, err := hashicups.NewClient(nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}
