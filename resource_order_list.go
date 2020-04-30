package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
		Update: resourceOrderUpdate,
		Delete: resourceOrderDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"item": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coffee_id": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"order": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"items": &schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"coffee": &schema.Schema{
										Type:     schema.TypeList,
										MaxItems: 1,
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
											},
										},
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
				},
			},
		},
	}
}

// Order -
type Order struct {
	ID    int         `json:"id,omitempty"`
	Items []OrderItem `json:"items,omitempty"`
}

// OrderItem -
type OrderItem struct {
	Coffee   Coffee `json:"coffee"`
	Quantity int    `json:"quantity"`
}

// Coffee -
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
		return err
	}

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

	resourceOrderRead(d, m)

	return nil
}

func resourceOrderRead(d *schema.ResourceData, m interface{}) error {
	orderID := d.Id()

	order, err := getOrder(orderID, m)
	if err != nil {
		return err
	}

	if err := d.Set("order", []interface{}{order}); err != nil {
		return err
	}

	return nil
}

func getOrder(orderID string, m interface{}) (map[string]interface{}, error) {
	config := m.(*Config)

	var client = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:9090/orders/%s", orderID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", config.Token)

	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// parse response body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	order := Order{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}

	o := make(map[string]interface{})
	if order.ID != 0 {
		o["id"] = order.ID
		o["items"] = flattenOrderItems(&order.Items)
	}

	return o, nil
}

func flattenOrderItems(orderItems *[]OrderItem) []interface{} {
	if orderItems != nil {
		ois := make([]interface{}, len(*orderItems), len(*orderItems))

		for i, orderItem := range *orderItems {
			oi := make(map[string]interface{})

			oi["coffee"] = flattenCoffee(orderItem.Coffee)
			oi["quantity"] = orderItem.Quantity

			ois[i] = oi
		}

		return ois
	}

	return make([]interface{}, 0)
}

func flattenCoffee(coffee Coffee) []interface{} {
	c := make(map[string]interface{})
	c["id"] = coffee.ID
	c["name"] = coffee.Name
	c["teaser"] = coffee.Teaser
	c["description"] = coffee.Description
	c["price"] = coffee.Price
	c["image"] = coffee.Image

	return []interface{}{c}
}

func resourceOrderUpdate(d *schema.ResourceData, m interface{}) error {
	// Enable partial state mode
	d.Partial(true)

	if d.HasChange("item") {
		orderID := d.Id()

		// Try updating the order
		if err := updateOrder(orderID, d, m); err != nil {
			return err
		}

		d.SetPartial("last_updated")

		d.Set("last_updated", time.Now().Format(time.RFC850))

		d.Set("item", d.Get("item").([]interface{}))

		// return nil
	}

	d.Partial(false)

	d.Set("item", d.Get("item").([]interface{}))

	return resourceOrderRead(d, m)
}

func updateOrder(orderID string, d *schema.ResourceData, m interface{}) error {
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
		return err
	}

	var client = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:9090/orders/%s", orderID), strings.NewReader(string(rb)))
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

	return nil
}

func resourceOrderDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	orderID := d.Id()

	var client = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:9090/orders/%s", orderID), nil)
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

	if string(body) != "Deleted order" {
		return errors.New(string(body))
	}

	d.SetId("")
	return nil
}
