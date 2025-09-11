package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure EphemeralDecryptResource satisfies various ephemeral resource interfaces.
var (
	_ ephemeral.EphemeralResource              = &EphemeralDecryptResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &EphemeralDecryptResource{}
)

func NewEphemeralDecryptResource() ephemeral.EphemeralResource {
	return &EphemeralDecryptResource{}
}

// EphemeralDecryptResource defines the ephemeral resource implementation for decryption.
type EphemeralDecryptResource struct {
	client *MetaClient
}

// EphemeralDecryptResourceModel describes the ephemeral resource data model.
type EphemeralDecryptResourceModel struct {
	Content   types.String `tfsdk:"content"`
	Decrypted types.String `tfsdk:"decrypted"`
	Id        types.String `tfsdk:"id"`
}

func (r *EphemeralDecryptResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_decrypt"
}

func (r *EphemeralDecryptResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ephemeral resource to decrypt vaulted content. The decrypted result is available during the Terraform run but is not stored in state for security.",

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
				MarkdownDescription: "Resource identifier",
				Computed:            true,
			},
		},
	}
}

func (r *EphemeralDecryptResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*MetaClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			"Expected *MetaClient, got something else. Please report this issue to the provider developers.",
		)

		return
	}

	r.client = client
}

func (r *EphemeralDecryptResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data EphemeralDecryptResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Provider Configuration Error",
			"Provider must be configured before using ephemeral resources.",
		)
		return
	}

	// Decrypt the content
	content := data.Content.ValueString()

	decrypted, err := r.client.DecryptValue(content)
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
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func (r *EphemeralDecryptResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	// No cleanup needed for decryption
}
