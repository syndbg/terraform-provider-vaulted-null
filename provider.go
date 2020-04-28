package main

import (
	"encoding/json"
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
		DataSourcesMap: map[string]*schema.Resource{
			"hashicups_coffees":     dataSourceCoffees(),
			"hashicups_ingredients": dataSourceIngredients(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// Config - contains provider configuration (Hashicups auth)
type Config struct {
	UserID   string
	Username string
	Token    string
}

// AuthStruct -
type AuthStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse -
type AuthResponse struct {
	UserID   int    `json:"user_id,omitempty`
	Username string `json:"username,omitempty`
	Token    string `json:"token,omitempty"`
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	if (username != "") && (password != "") {
		// form request body
		rb, err := json.Marshal(AuthStruct{
			Username: "dos",
			Password: "test123",
		})
		if err != nil {
			panic(err)
		}

		// authenticate
		var client = &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("POST", "http://localhost:9090/signin", strings.NewReader(string(rb)))
		if err != nil {
			return nil, err
		}

		r, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		// parse response body
		ar := AuthResponse{}
		err = json.Unmarshal(body, &ar)
		if err != nil {
			return nil, err
		}

		config := Config{
			UserID:   strconv.Itoa(ar.UserID),
			Username: username,
			Token:    ar.Token,
		}

		return &config, nil
	}

	return nil, nil
}
