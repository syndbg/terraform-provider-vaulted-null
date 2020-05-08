package hashicups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceCoffees() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCoffeesRead,
		Schema: map[string]*schema.Schema{
			"coffees": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"teaser": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"image": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"ingredients": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ingredient_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCoffeesRead(d *schema.ResourceData, m interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/coffees", "http://localhost:19090"), nil)
	if err != nil {
		return err
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	coffees := make([]map[string]interface{}, 0)
	err = json.NewDecoder(r.Body).Decode(&coffees)
	if err != nil {
		return err
	}

	if err := d.Set("coffees", coffees); err != nil {
		return err
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
