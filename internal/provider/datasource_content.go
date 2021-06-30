package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceContent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceContentRead,
		Schema: map[string]*schema.Schema{
			"content": {
				Type: schema.TypeString,
				Required: true,
			},
			"decrypted": {
				Type: schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceContentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	metaClient := m.(*MetaClient)

	content := d.Get("content").(string)
	decryptedContent, err := metaClient.DecryptValue(content)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("decrypted", decryptedContent)
	return nil
}