package hashicups

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"hashicups": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("HASHICUPS_USERNAME"); err == "" {
		t.Fatal("HASHICUPS_USERNAME must be set for acceptance tests")
	}
	if err := os.Getenv("HASHICUPS_PASSWORD"); err == "" {
		t.Fatal("HASHICUPS_PASSWORD must be set for acceptance tests")
	}
}
