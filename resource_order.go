package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrderCreate,
		Read:   resourceOrderRead,
		Delete: resourceOrderDelete,
		Schema: map[string]*schema.Schema{
			"item": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coffee_id": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

// Order -
type Order struct {
	ID    int         `db:"id" json:"id,omitempty"`
	Items []OrderItem `json:"items,omitempty"`
}

// OrderItem -
type OrderItem struct {
	Coffee   Coffee `json:"coffee"`
	Quantity int    `json:"quantity"`
}

// Coffee
type Coffee struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Teaser      string  `json:"teaser"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

func resourceOrderCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	items := d.Get("item").([]interface{})
	ois := []OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		oi := OrderItem{
			Coffee: Coffee{
				ID: i["coffee_id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	rb, err := json.Marshal(ois)
	if err != nil {
		panic(err)
	}

	log.Printf("==== %+v", string(rb))

	var client = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:9090/orders", strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", config.Token)

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// parse response body
	o := Order{}
	err = json.Unmarshal(body, &o)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(o.ID))

	return nil
}

func resourceOrderRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func getOrder(orderID string) (map[string]interface{}, error) {
	var client = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:9090/orders/%s", orderID), nil)
	if err != nil {
		return nil, err
	}

	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	order := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func resourceOrderDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
