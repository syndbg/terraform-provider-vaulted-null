package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure EncryptContentResource satisfies various resource interfaces.
var (
	_ resource.Resource              = &EncryptContentResource{}
	_ resource.ResourceWithConfigure = &EncryptContentResource{}
)

func NewEncryptContentResource() resource.Resource {
	return &EncryptContentResource{}
}

// EncryptContentResource defines the resource implementation.
type EncryptContentResource struct {
	client *MetaClient
}

// EncryptContentResourceModel describes the resource data model.
type EncryptContentResourceModel struct {
	Plaintext types.String `tfsdk:"plaintext"`
	Encrypted types.String `tfsdk:"encrypted"`
	Id        types.String `tfsdk:"id"`
}

func (r *EncryptContentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt_content"
}

func (r *EncryptContentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to encrypt plaintext content and store the encrypted result in state",

		Attributes: map[string]schema.Attribute{
			"plaintext": schema.StringAttribute{
				MarkdownDescription: "The plaintext content to encrypt",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *EncryptContentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*MetaClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *MetaClient, got something else. Please report this issue to the provider developers.",
		)

		return
	}

	r.client = client
}

func (r *EncryptContentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EncryptContentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if client is configured
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Provider Configuration Error",
			"Provider must be configured before using resources.",
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
	data.Id = types.StringValue(plaintext)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptContentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EncryptContentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The encrypted content is already computed and stored in state, so we just need to keep it
	// No need to re-encrypt as this is a deterministic operation

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptContentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource uses RequiresReplace for the plaintext attribute, so Update should not be called
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"This resource should require replacement when plaintext changes",
	)
}

func (r *EncryptContentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do on delete - just remove from state
}
