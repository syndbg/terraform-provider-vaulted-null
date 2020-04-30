package hashicups

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccHashicupsOrderBasic(t *testing.T) {
	var items []interface{}
	coffeeID := "1"
	quantity := "2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHashicupsOrderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHashicupsOrderConfigBasic(coffeeID, quantity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHashicupsOrderExists("hashicups_order.new"),
				),
			},
		},
	})
}

func testAccCheckHashicupsOrderDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hashicups_order" {
			continue
		}

		orderID := rs.Primary.ID

		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/orders/%s", HostURL, orderID), nil)
		if err != nil {
			return err
		}

		body, err := c.doRequest(req, false)
		if err != nil {
			return err
		}

		if string(body) != "Deleted order" {
			return errors.New(string(body))
		}
	}

	return nil
}

func testAccCheckHashicupsOrderConfigBasic(coffeeID, quantity string) string {
	return fmt.Sprintf(`
	resource "hashicups_order" "new" {
		items {
			coffee {
				id = %s
			}
    		quantity = %s
  		}
	}
	`, coffeeID, quantity)
}

func testAccCheckHashicupsOrderExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OrderID set")
		}

		return nil
	}
}
