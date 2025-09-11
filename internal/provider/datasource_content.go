package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ContentDataSource satisfies various data source interfaces.
var (
	_ datasource.DataSource              = &ContentDataSource{}
	_ datasource.DataSourceWithConfigure = &ContentDataSource{}
)

func NewContentDataSource() datasource.DataSource {
	return &ContentDataSource{}
}

// ContentDataSource defines the data source implementation.
type ContentDataSource struct {
	client *MetaClient
}

// ContentDataSourceModel describes the data source data model.
type ContentDataSourceModel struct {
	Content   types.String `tfsdk:"content"`
	Decrypted types.String `tfsdk:"decrypted"`
	Id        types.String `tfsdk:"id"`
}

func (d *ContentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content"
}

func (d *ContentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to decrypt vaulted content",

		Attributes: map[string]schema.Attribute{
			"content": schema.StringAttribute{
				MarkdownDescription: "The encrypted content (vaulted format) to decrypt",
				Required:            true,
			},
			"decrypted": schema.StringAttribute{
				MarkdownDescription: "The decrypted plaintext content",
				Computed:            true,
				Sensitive:           true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
		},
	}
}

func (d *ContentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*MetaClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *MetaClient, got something else. Please report this issue to the provider developers.",
		)

		return
	}

	d.client = client
}

func (d *ContentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if client is configured
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider Configuration Error",
			"Provider must be configured before using data sources.",
		)
		return
	}

	// Decrypt the content
	content := data.Content.ValueString()

	decrypted, err := d.client.DecryptValue(content)
	if err != nil {
		resp.Diagnostics.AddError(
			"Decryption Error",
			"Unable to decrypt content: "+err.Error(),
		)
		return
	}

	data.Decrypted = types.StringValue(decrypted)
	data.Id = types.StringValue(content)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
