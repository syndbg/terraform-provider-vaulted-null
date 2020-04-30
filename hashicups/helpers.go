package hashicups

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (c *Config) doRequest(req *http.Request, auth bool) ([]byte, error) {
	if auth {
		req.Header.Set("Authorization", c.Token)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

func getOrderItems(orderID string, m interface{}) ([]interface{}, error) {
	c := m.(*Config)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/orders/%s", c.Host, orderID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, true)
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
