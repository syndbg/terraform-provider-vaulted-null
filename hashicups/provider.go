package hashicups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// HostURL -
const HostURL string = "http://localhost:9090"

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hashicups_order": resourceOrder(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hashicups_coffees":     dataSourceCoffees(),
			"hashicups_ingredients": dataSourceIngredients(),
			"hashicups_order":       dataSourceOrder(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	c := Config{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	if (username != "") && (password != "") {
		// form request body
		rb, err := json.Marshal(AuthStruct{
			Username: username,
			Password: password,
		})
		if err != nil {
			return nil, err
		}

		// authenticate
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/signin", HostURL), strings.NewReader(string(rb)))
		if err != nil {
			return nil, err
		}

		body, err := c.doRequest(req, false)
		if err != nil {
			return nil, err
		}

		// parse response body
		ar := AuthResponse{}
		err = json.Unmarshal(body, &ar)
		if err != nil {
			return nil, err
		}

		c.UserID = strconv.Itoa(ar.UserID)
		c.Username = username
		c.Token = ar.Token

		return &c, nil
	}

	return &c, nil
}
