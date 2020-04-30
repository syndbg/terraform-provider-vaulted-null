package hashicups

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
		Host:   "http://localhost:9090",
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	if (username != "") && (password != "") {
		// form request body
		rb, err := json.Marshal(AuthStruct{
			Username: "dos",
			Password: "test123",
		})
		if err != nil {
			return nil, err
		}

		// authenticate
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/signin", c.Host), strings.NewReader(string(rb)))
		if err != nil {
			return nil, err
		}

		r, err := c.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		// parse response body
		ar := AuthResponse{}
		err = json.Unmarshal(body, &ar)
		if err != nil {
			return nil, err
		}

		c = Config{
			UserID:   strconv.Itoa(ar.UserID),
			Username: username,
			Token:    ar.Token,
			Host:     c.Host,
			Client:   c.Client,
		}

		return &c, nil
	}

	return &c, nil
}
