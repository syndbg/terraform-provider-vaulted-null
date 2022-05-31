package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEncryptContent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"plaintext": {
				Type:     schema.TypeString,
				Required: true,
			},
			"encrypted": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		ReadContext:   schema.NoopContext,
		CreateContext: resourceEncryptContentCreate,
		DeleteContext: resourceEncryptContentDelete,
	}
}

func resourceEncryptContentCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	metaClient, ok := m.(*MetaClient)
	if !ok {
		return diag.Errorf("unexpected meta client: %v", metaClient)
	}

	plaintext, ok := d.Get("plaintext").(string)
	if !ok {
		return diag.Errorf("unexpected `plaintext`, must be string: %v", metaClient)
	}

	encryptedContent, err := metaClient.EncryptValue(plaintext)
	if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: Set to something unique yet changing.
	// Since this is just a resource that generates a computed value, there's no need for comparison.
	d.SetId(plaintext)

	err = d.Set("encrypted", encryptedContent)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceEncryptContentDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// NOTE: Just deletes it from the state
	d.SetId("")
	return nil
}
