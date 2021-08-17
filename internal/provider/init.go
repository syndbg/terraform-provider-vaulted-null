package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// nolint:gochecknoinits
func init() {
	// NOTE: Part of TF registry docs generation
	schema.DescriptionKind = schema.StringMarkdown
}
