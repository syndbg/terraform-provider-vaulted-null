package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Ingredient -
type Ingredient struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

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
	var client = &http.Client{Timeout: 10 * time.Second}

	coffeeID := d.Get("coffee_id").(int)
	cID := strconv.Itoa(coffeeID)

	log.Printf("==%+v", cID)

	r, err := client.Get(fmt.Sprintf("http://localhost:9090/coffees/%s/ingredients", cID))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	ingredients := make([]map[string]interface{}, 0)

	ings := []Ingredient{}
	err = json.NewDecoder(r.Body).Decode(&ings)
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
