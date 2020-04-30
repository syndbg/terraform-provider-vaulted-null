package hashicups

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"io/ioutil"
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
// )

// func resourceOrder() *schema.Resource {
// 	return &schema.Resource{
// 		Create: resourceOrderCreate,
// 		Read:   resourceOrderRead,
// 		Delete: resourceOrderDelete,
// 		Schema: map[string]*schema.Schema{
// 			"item": &schema.Schema{
// 				Type:     schema.TypeSet,
// 				Required: true,
// 				ForceNew: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"coffee_id": &schema.Schema{
// 							Type:     schema.TypeInt,
// 							Required: true,
// 							ForceNew: true,
// 						},
// 						"quantity": &schema.Schema{
// 							Type:     schema.TypeInt,
// 							Required: true,
// 							ForceNew: true,
// 						},
// 					},
// 				},
// 			},
// 			"order": &schema.Schema{
// 				Type:     schema.TypeList,
// 				MaxItems: 1,
// 				Computed: true,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"id": &schema.Schema{
// 							Type:     schema.TypeInt,
// 							Computed: true,
// 						},
// 						"items": &schema.Schema{
// 							Type:     schema.TypeList,
// 							MaxItems: 1,
// 							Computed: true,
// 							Elem: &schema.Resource{
// 								Schema: map[string]*schema.Schema{
// 									"coffee": &schema.Schema{
// 										Type:     schema.TypeMap,
// 										Computed: true,
// 										Elem: &schema.Schema{
// 											Type: schema.TypeString,
// 										},
// 									},
// 									"quantity": &schema.Schema{
// 										Type:     schema.TypeInt,
// 										Required: true,
// 										ForceNew: true,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// // Order -
// type Order struct {
// 	ID    int         `json:"id,omitempty"`
// 	Items []OrderItem `json:"items,omitempty"`
// }

// // OrderItem -
// type OrderItem struct {
// 	Coffee   Coffee `json:"coffee"`
// 	Quantity int    `json:"quantity"`
// }

// // Coffee -
// type Coffee struct {
// 	ID          int     `json:"id"`
// 	Name        string  `json:"name"`
// 	Teaser      string  `json:"teaser"`
// 	Description string  `json:"description"`
// 	Price       float64 `json:"price"`
// 	Image       string  `json:"image"`
// }

// func resourceOrderCreate(d *schema.ResourceData, m interface{}) error {
// 	config := m.(*Config)
// 	items := d.Get("item").([]interface{})
// 	ois := []OrderItem{}

// 	for _, item := range items {
// 		i := item.(map[string]interface{})

// 		oi := OrderItem{
// 			Coffee: Coffee{
// 				ID: i["coffee_id"].(int),
// 			},
// 			Quantity: i["quantity"].(int),
// 		}

// 		ois = append(ois, oi)
// 	}

// 	rb, err := json.Marshal(ois)
// 	if err != nil {
// 		return err
// 	}

// 	var client = &http.Client{Timeout: 10 * time.Second}
// 	req, err := http.NewRequest("POST", "http://localhost:9090/orders", strings.NewReader(string(rb)))
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Authorization", config.Token)

// 	r, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer r.Body.Close()

// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return err
// 	}

// 	// parse response body
// 	o := Order{}
// 	err = json.Unmarshal(body, &o)
// 	if err != nil {
// 		return err
// 	}

// 	d.SetId(strconv.Itoa(o.ID))

// 	resourceOrderRead(d, m)

// 	return nil
// }

// func resourceOrderRead(d *schema.ResourceData, m interface{}) error {
// 	oID := d.Id()

// 	order, err := getOrder(oID, m)
// 	if err != nil {
// 		return err
// 	}

// 	if err := d.Set("order", []interface{}{order}); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func getOrder(oID string, m interface{}) (map[string]interface{}, error) {
// 	config := m.(*Config)

// 	var client = &http.Client{Timeout: 10 * time.Second}
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:9090/orders/%s", oID), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Authorization", config.Token)

// 	r, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer r.Body.Close()

// 	// parse response body
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	o := Order{}
// 	err = json.Unmarshal(body, &o)
// 	if err != nil {
// 		return nil, err
// 	}

// 	order := make(map[string]interface{})
// 	if o.ID != 0 {
// 		order["id"] = o.ID
// 		order["items"] = flattenOrderItems(&o.Items)
// 	}

// 	return order, nil
// }

// func flattenOrderItems(ois *[]OrderItem) []interface{} {
// 	if ois != nil {
// 		orderItems := make([]interface{}, len(*ois), len(*ois))

// 		for i, oi := range *ois {
// 			orderItem := make(map[string]interface{})

// 			orderItem["coffee"] = flattenCoffee(oi.Coffee)
// 			orderItem["quantity"] = oi.Quantity

// 			orderItems[i] = orderItem
// 		}

// 		return orderItems
// 	}

// 	return make([]interface{}, 0)
// }

// func flattenCoffee(c Coffee) map[string]string {
// 	coffee := make(map[string]string)
// 	coffee["id"] = strconv.Itoa(c.ID)
// 	coffee["name"] = c.Name
// 	coffee["teaser"] = c.Teaser
// 	coffee["description"] = c.Description
// 	coffee["price"] = fmt.Sprintf("%f", c.Price)
// 	coffee["image"] = c.Image

// 	return coffee
// }

// func resourceOrderDelete(d *schema.ResourceData, m interface{}) error {
// 	d.SetId("")
// 	return nil
// }
