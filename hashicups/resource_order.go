package hashicups

import (
	"strconv"
	"time"

	hc "github.com/hashicorp-demoapp/hashicups-client-go"
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

func resourceOrderCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*hc.Client)

	items := d.Get("items").([]interface{})
	ois := []hc.OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		co := i["coffee"].([]interface{})[0]
		coffee := co.(map[string]interface{})

		oi := hc.OrderItem{
			Coffee: hc.Coffee{
				ID: coffee["id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	o, err := c.CreateOrder(ois)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(o.ID))

	resourceOrderRead(d, m)

	return nil
}

func resourceOrderRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*hc.Client)

	orderID := d.Id()

	order, err := c.GetOrder(orderID)
	if err != nil {
		return err
	}

	orderItems := flattenOrderItems(&order.Items)
	if err := d.Set("items", orderItems); err != nil {
		return err
	}

	return nil
}

func flattenOrderItems(orderItems *[]hc.OrderItem) []interface{} {
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

func flattenCoffee(coffee hc.Coffee) []interface{} {
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
	c := m.(*hc.Client)

	orderID := d.Id()

	// enable partial state mode
	d.Partial(true)

	if d.HasChange("items") {
		items := d.Get("items").([]interface{})
		ois := []hc.OrderItem{}

		for _, item := range items {
			i := item.(map[string]interface{})

			co := i["coffee"].([]interface{})[0]
			coffee := co.(map[string]interface{})

			oi := hc.OrderItem{
				Coffee: hc.Coffee{
					ID: coffee["id"].(int),
				},
				Quantity: i["quantity"].(int),
			}

			ois = append(ois, oi)
		}

		_, err := c.UpdateOrder(orderID, ois)
		if err != nil {
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

func resourceOrderDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*hc.Client)

	orderID := d.Id()

	err := c.DeleteOrder(orderID)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
