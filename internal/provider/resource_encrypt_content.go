package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEncryptContent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"plaintext": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func generateHashedTimestamp(unixTimestamp int64) string {
	timestampString := strconv.FormatInt(unixTimestamp, 10)
	timestampBytes := []byte(timestampString)

	hasher := sha256.New()
	hasher.Write(timestampBytes)
	hashedBytes := hasher.Sum(nil)

	// Convert the hashed bytes to a hexadecimal string
	hashedTimestamp := hex.EncodeToString(hashedBytes)

	return hashedTimestamp
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
	currentTimeUnix := time.Now().Unix()
	hashedTimestamp := generateHashedTimestamp(currentTimeUnix)
	d.SetId(string(hashedTimestamp))

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
