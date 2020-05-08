package hashicups

// func flattenOrderItems(orderItems *[]OrderItem) []interface{} {
// 	if orderItems != nil {
// 		ois := make([]interface{}, len(*orderItems), len(*orderItems))

// 		for i, orderItem := range *orderItems {
// 			oi := make(map[string]interface{})

// 			oi["coffee"] = flattenCoffee(orderItem.Coffee)
// 			oi["quantity"] = orderItem.Quantity

// 			ois[i] = oi
// 		}

// 		return ois
// 	}

// 	return make([]interface{}, 0)
// }

// func flattenCoffee(coffee Coffee) []interface{} {
// 	c := make(map[string]interface{})
// 	c["id"] = coffee.ID
// 	c["name"] = coffee.Name
// 	c["teaser"] = coffee.Teaser
// 	c["description"] = coffee.Description
// 	c["price"] = coffee.Price
// 	c["image"] = coffee.Image

// 	return []interface{}{c}
// }
