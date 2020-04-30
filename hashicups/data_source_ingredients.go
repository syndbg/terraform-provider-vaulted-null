package hashicups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceIngredients() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIngredientsRead,
		Schema: map[string]*schema.Schema{
			"coffee_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"ingredients": &schema.Schema{
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
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unit": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIngredientsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)
	coffeeID := d.Get("coffee_id").(int)
	cID := strconv.Itoa(coffeeID)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/coffees/%s/ingredients", HostURL, cID), nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req, false)
	if err != nil {
		return err
	}

	ingredients := make([]map[string]interface{}, 0)

	ings := []Ingredient{}
	err = json.Unmarshal(body, &ings)
	if err != nil {
		return err
	}

	for _, v := range ings {
		ingredient := make(map[string]interface{})

		ingredient["id"] = v.ID
		ingredient["name"] = fmt.Sprintf("ingredient - %+v", v.Name)
		ingredient["quantity"] = v.Quantity
		ingredient["unit"] = v.Unit

		ingredients = append(ingredients, ingredient)
	}

	if err := d.Set("ingredients", ingredients); err != nil {
		return err
	}

	d.SetId(cID)

	return nil
}
