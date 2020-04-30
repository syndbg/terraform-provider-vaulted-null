package main

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceOrder() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrderRead,
		Schema: map[string]*schema.Schema{
			"order_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
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

func dataSourceOrderRead(d *schema.ResourceData, m interface{}) error {
	orderID := strconv.Itoa(d.Get("order_id").(int))

	order, err := getOrder(orderID, m)
	if err != nil {
		return err
	}

	if err := d.Set("order", []interface{}{order}); err != nil {
		return err
	}

	d.SetId(orderID)

	return nil
}
