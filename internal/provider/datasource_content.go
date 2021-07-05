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
				Type:     schema.TypeString,
				Required: true,
			},
			"decrypted": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceContentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	metaClient, ok := m.(*MetaClient)
	if !ok {
		return diag.Errorf("unexpected meta client: %v", metaClient)
	}

	content, ok := d.Get("content").(string)
	if !ok {
		return diag.Errorf("unexpected `content`, must be string: %v", metaClient)
	}

	decryptedContent, err := metaClient.DecryptValue(content)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("decrypted", decryptedContent)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
