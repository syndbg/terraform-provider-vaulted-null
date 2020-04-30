package hashicups

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
			"items": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coffee": &schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": &schema.Schema{
										Type:     schema.TypeInt,
										Required: true,
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
	c := m.(*Config)
	items := d.Get("items").([]interface{})
	ois := []OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		co := i["coffee"].([]interface{})[0]
		coffee := co.(map[string]interface{})

		oi := OrderItem{
			Coffee: Coffee{
				ID: coffee["id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	rb, err := json.Marshal(ois)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/orders", c.Host), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.Token)

	r, err := c.Client.Do(req)
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

	items, err := getOrderItems(orderID, m)
	if err != nil {
		return err
	}

	if err := d.Set("items", items); err != nil {
		return err
	}

	return nil
}

func getOrderItems(orderID string, m interface{}) ([]interface{}, error) {
	c := m.(*Config)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/orders/%s", c.Host, orderID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.Token)

	r, err := c.Client.Do(req)
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

	return flattenOrderItems(&order.Items), nil
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
	orderID := d.Id()

	// enable partial state mode
	d.Partial(true)

	if d.HasChange("items") {
		if err := updateOrder(orderID, d, m); err != nil {
			return err
		}

		d.SetPartial("last_updated")
		d.Set("last_updated", time.Now().Format(time.RFC850))

		// if err := resourceOrderRead(d, m); err != nil {
		// 	return err
		// }

		// return nil
	}
	d.Partial(false)

	return resourceOrderRead(d, m)
}

func updateOrder(orderID string, d *schema.ResourceData, m interface{}) error {
	c := m.(*Config)

	items := d.Get("items").([]interface{})
	ois := []OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		co := i["coffee"].([]interface{})[0]
		coffee := co.(map[string]interface{})

		oi := OrderItem{
			Coffee: Coffee{
				ID: coffee["id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	rb, err := json.Marshal(ois)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/orders/%s", c.Host, orderID), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.Token)

	r, err := c.Client.Do(req)
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
	c := m.(*Config)
	orderID := d.Id()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/orders/%s", c.Host, orderID), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.Token)

	r, err := c.Client.Do(req)
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
