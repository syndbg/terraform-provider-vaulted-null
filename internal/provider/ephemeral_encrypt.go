package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure EphemeralEncryptResource satisfies various ephemeral resource interfaces.
var (
	_ ephemeral.EphemeralResource              = &EphemeralEncryptResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &EphemeralEncryptResource{}
)

func NewEphemeralEncryptResource() ephemeral.EphemeralResource {
	return &EphemeralEncryptResource{}
}

// EphemeralEncryptResource defines the ephemeral resource implementation for encryption.
type EphemeralEncryptResource struct {
	client *MetaClient
}

// EphemeralEncryptResourceModel describes the ephemeral resource data model.
type EphemeralEncryptResourceModel struct {
	Plaintext types.String `tfsdk:"plaintext"`
	Encrypted types.String `tfsdk:"encrypted"`
	Id        types.String `tfsdk:"id"`
}

func (r *EphemeralEncryptResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt"
}

func (r *EphemeralEncryptResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ephemeral resource to encrypt plaintext content. The encrypted result is available during the Terraform run but is not stored in state for security.",

		Attributes: map[string]schema.Attribute{
			"plaintext": schema.StringAttribute{
				MarkdownDescription: "The plaintext content to encrypt",
				Required:            true,
				Sensitive:           true,
			},
			"encrypted": schema.StringAttribute{
				MarkdownDescription: "The encrypted content (vaulted format)",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier",
				Computed:            true,
			},
		},
	}
}

func (r *EphemeralEncryptResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *EphemeralEncryptResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data EphemeralEncryptResourceModel

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

	// Encrypt the plaintext content
	plaintext := data.Plaintext.ValueString()

	encrypted, err := r.client.EncryptValue(plaintext)
	if err != nil {
		resp.Diagnostics.AddError(
			"Encryption Error",
			"Unable to encrypt content: "+err.Error(),
		)
		return
	}

	// Set the encrypted value and ID
	data.Encrypted = types.StringValue(encrypted)
	data.Id = types.StringValue(plaintext) // Use plaintext as ID (this won't be stored in state)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func (r *EphemeralEncryptResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	// No cleanup needed for encryption
}
