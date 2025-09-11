# terraform-provider-vaulted-null

[![CI Status](https://github.com/syndbg/terraform-provider-vaulted-null/workflows/CI/badge.svg)](https://github.com/syndbg/terraform-provider-vaulted-null/actions)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/syndbg/vaulted-null/latest)

A Terraform provider for encrypting and decrypting secrets using RSA keys or AWS KMS, perfect for secure secret management in your infrastructure as code.

## 🔐 What does this provider do?

This provider allows you to:
- **Encrypt plaintext secrets** into secure, encrypted payloads
- **Decrypt encrypted payloads** back to plaintext for use in your Terraform configurations
- Use **RSA key pairs** or **AWS KMS** for encryption/decryption
- Securely manage secrets in Terraform Cloud, CI/CD pipelines, and other environments

## 🆚 Which provider should I use?

| Provider | Use Case | State Storage |
|----------|----------|---------------|
| **terraform-provider-vaulted-null** | Terraform Cloud, trusted CI environments | ⚠️ Decrypted secrets stored in state |
| [terraform-provider-vaulted](https://github.com/sumup-oss/terraform-provider-vaulted) | Public cloud CI, less secure environments | ✅ Secrets never stored in plaintext |

**Choose vaulted-null when:** You use Terraform Cloud or have secure state storage and want simple secret decryption.

**Choose vaulted when:** You need maximum security and can't have plaintext secrets in your Terraform state.

## 🚀 Quick Start

### 1. Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    vaulted-null = {
      source  = "syndbg/vaulted-null"
      version = "~> 0.3"
    }
  }
}
```

### 2. Generate RSA Keys

```bash
# Generate private key
openssl genrsa -out private.pem 2048

# Extract public key
openssl rsa -in private.pem -pubout -out public.pem
```

### 3. Configure the Provider

```hcl
provider "vaulted-null" {
  private_key_path = "./private.pem"
  public_key_path  = "./public.pem"
}
```

### 4. Encrypt a Secret

```hcl
resource "vaulted-null_encrypt_content" "api_key" {
  plaintext = "super-secret-api-key-12345"
}

output "encrypted_api_key" {
  value = vaulted-null_encrypt_content.api_key.encrypted
}
```

### 5. Decrypt a Secret

```hcl
data "vaulted-null_content" "api_key" {
  content = "$VED;1.0::your-encrypted-payload-here..."
}

output "decrypted_api_key" {
  value     = data.vaulted-null_content.api_key.decrypted
  sensitive = true
}
```

## 📋 Configuration Options

### Provider Configuration

#### RSA Key Authentication

```hcl
provider "vaulted-null" {
  # Option 1: File paths
  private_key_path = "./private.pem"
  public_key_path  = "./public.pem"

  # Option 2: Key content directly
  private_key_content = var.private_key
  public_key_content  = var.public_key
}
```

#### AWS KMS Authentication

```hcl
provider "vaulted-null" {
  aws_kms_key_id = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  aws_region     = "us-east-1"
  aws_profile    = "my-profile"  # optional
}
```

### Environment Variables

You can also configure the provider using environment variables:

```bash
export VAULTED_PRIVATE_KEY_PATH="./private.pem"
export VAULTED_PUBLIC_KEY_PATH="./public.pem"
export VAULTED_AWS_KMS_KEY_ID="your-kms-key-id"
export AWS_REGION="us-east-1"
export AWS_PROFILE="your-profile"
```

## 💡 Usage Examples

### Example 1: Database Password Management

```hcl
# Encrypt your database password
resource "vaulted-null_encrypt_content" "db_password" {
  plaintext = "MySecureDBPassword123!"
}

# Store the encrypted password in your configuration
locals {
  encrypted_db_password = "$VED;1.0::IhQnapipQSRi9gR9OD2Bph3souOeg2DzyyhciRNOasXItG05..."
}

# Decrypt when needed
data "vaulted-null_content" "db_password" {
  content = local.encrypted_db_password
}

# Use in your resources
resource "aws_db_instance" "main" {
  identifier = "my-database"
  engine     = "mysql"
  password   = data.vaulted-null_content.db_password.decrypted
  # ... other configuration
}
```

### Example 2: API Keys for External Services

```hcl
# Multiple encrypted secrets
data "vaulted-null_content" "github_token" {
  content = var.encrypted_github_token
}

data "vaulted-null_content" "slack_webhook" {
  content = var.encrypted_slack_webhook
}

# Use in webhook configuration
resource "github_repository_webhook" "deploy_hook" {
  repository = "my-repo"

  configuration {
    url          = data.vaulted-null_content.slack_webhook.decrypted
    content_type = "json"
  }
}
```

### Example 3: Terraform Cloud with Encrypted Variables

```hcl
# In your Terraform Cloud workspace, set variables:
# TF_VAR_encrypted_secret = "your-encrypted-payload"

variable "encrypted_secret" {
  description = "Encrypted secret payload"
  type        = string
}

data "vaulted-null_content" "secret" {
  content = var.encrypted_secret
}

# Use the decrypted secret
resource "kubernetes_secret" "app_secret" {
  metadata {
    name = "app-secret"
  }

  data = {
    api_key = data.vaulted-null_content.secret.decrypted
  }
}
```

## 🔧 Resources and Data Sources

### Traditional Resources (Stored in State)

#### `vaulted-null_encrypt_content` Resource

Encrypts plaintext content and stores the result in state.

```hcl
resource "vaulted-null_encrypt_content" "example" {
  plaintext = "your-secret-here"
}
```

**Attributes:**
- `plaintext` (string, required) - The plaintext content to encrypt
- `encrypted` (string, computed) - The resulting encrypted payload
- `id` (string, computed) - Resource identifier

#### `vaulted-null_content` Data Source

Decrypts encrypted content (⚠️ plaintext stored in state).

```hcl
data "vaulted-null_content" "example" {
  content = "$VED;1.0::encrypted-payload..."
}
```

**Attributes:**
- `content` (string, required) - The encrypted payload to decrypt
- `decrypted` (string, computed) - The decrypted plaintext content
- `id` (string, computed) - Data source identifier

### 🔥 NEW: Ephemeral Resources (NOT Stored in State)

**Requires Terraform 1.10+**

Ephemeral resources provide maximum security by never storing sensitive data in Terraform state.

#### `vaulted-null_encrypt` Ephemeral Resource

Encrypts plaintext content temporarily during Terraform execution.

```hcl
ephemeral "vaulted-null_encrypt" "api_key" {
  plaintext = "super-secret-api-key"
}

# Use immediately in other resources
resource "kubernetes_secret" "app_secret" {
  data = {
    encrypted_key = ephemeral.vaulted-null_encrypt.api_key.encrypted
  }
}
```

**Attributes:**
- `plaintext` (string, required, sensitive) - The plaintext content to encrypt
- `encrypted` (string, computed) - The resulting encrypted payload
- `id` (string, computed) - Resource identifier

#### `vaulted-null_decrypt` Ephemeral Resource

Decrypts content temporarily without storing plaintext in state.

```hcl
ephemeral "vaulted-null_decrypt" "db_password" {
  content = var.encrypted_db_password
}

# Use immediately in provider configuration
provider "postgresql" {
  password = ephemeral.vaulted-null_decrypt.db_password.decrypted
}
```

**Attributes:**
- `content` (string, required) - The encrypted payload to decrypt
- `decrypted` (string, computed, sensitive) - The decrypted plaintext content
- `id` (string, computed) - Resource identifier

### 🛡️ Security Comparison

| Resource Type | Plaintext in State | Encrypted in State | Use Case |
|---------------|-------------------|-------------------|----------|
| **Traditional Resources** | ⚠️ Yes (data sources) | ✅ Yes | Basic encryption, less sensitive data |
| **Ephemeral Resources** | ❌ Never | ❌ No (temporary only) | Maximum security, highly sensitive data |

### 🔄 How Diffing Works

Understanding how Terraform tracks changes is crucial for choosing the right resource type:

#### Traditional Resources: State-Based Diffing
```bash
# When you change a secret:
~ resource "vaulted-null_encrypt_content" "example" {
    ~ plaintext = "old-secret" → "new-secret"
    ~ encrypted = "old-encrypted" → (known after apply)
  }
```

**Behavior:**
- ✅ **Shows clear diffs** when inputs change
- ✅ **Tracks history** in state file
- ⚠️ **Stores sensitive data** in state

#### Ephemeral Resources: No State Tracking
```bash
# When you change a secret:
# (No diff shown for ephemeral resource itself)

~ resource "kubernetes_secret" "app" {
    ~ data = {
        password = (sensitive value)
      }
  }
```

**Behavior:**
- ❌ **No diffs shown** for ephemeral resources
- ✅ **Downstream resources** show diffs
- ✅ **Maximum security** - no state persistence
- 🔄 **Recreated every run** regardless of changes

#### When to Use Each

**Use Traditional Resources when:**
- You need to track changes over time
- Security of state storage is acceptable
- Working with less sensitive data
- Debugging requires state inspection

**Use Ephemeral Resources when:**
- Maximum security is required
- Secrets should never touch state files
- Working with highly sensitive data (passwords, tokens)
- State files might be compromised or shared

## 🔐 Security Best Practices

### 1. Key Management

- **Store private keys securely** - Never commit private keys to version control
- **Use strong RSA keys** - Minimum 2048 bits, preferably 4096 bits
- **Rotate keys regularly** - Establish a key rotation schedule
- **Limit key access** - Only authorized personnel should access private keys

### 2. Environment Security

- **Use Terraform Cloud** - For secure state storage when using this provider
- **Encrypt state storage** - Ensure your Terraform state is encrypted at rest
- **Audit access** - Monitor who accesses encrypted secrets
- **Use least privilege** - Grant minimal necessary permissions

### 3. AWS KMS Best Practices

- **Use customer managed keys** - Don't rely on AWS managed keys for sensitive data
- **Set up key policies** - Restrict key usage to specific roles/users
- **Enable CloudTrail** - Monitor key usage for auditing
- **Use different keys** - Separate keys for different environments

## ❗ Troubleshooting

### Common Issues

#### "Failed to read private key"
```
Error: failed to read private key from path ./private.pem
```

**Solution:** Ensure your private key is in the correct RSA format:
```bash
# Convert to traditional RSA format if needed
openssl rsa -in private.pem -out private_rsa.pem -traditional
```

#### "Unable to decrypt value"
```
Error: unable to decrypt `value`
```

**Possible causes:**
1. Wrong private key for the encrypted payload
2. Corrupted encrypted payload
3. Key format mismatch

**Solution:** Verify you're using the correct key pair that was used for encryption.

#### "Authentication signature from unknown issuer"
```
Error: authentication signature from unknown issuer
```

**Solution:** You're trying to use the public registry. For local development, see the [development guide](./CONTRIBUTING.md).

### Getting Help

1. Check the [examples directory](./examples/) for working configurations
2. Review the [contributing guide](./CONTRIBUTING.md) for development setup
3. [Open an issue](https://github.com/syndbg/terraform-provider-vaulted-null/issues) if you find a bug

## 🧪 Local Development

Want to contribute or test locally? See our [development guide](./CONTRIBUTING.md) for:
- Setting up the development environment
- Building and testing the provider
- Running examples locally
- Submitting contributions

## 📚 Advanced Topics

### Working with Multiple Environments

```hcl
# Development environment
provider "vaulted-null" {
  alias           = "dev"
  private_key_path = "./keys/dev-private.pem"
  public_key_path  = "./keys/dev-public.pem"
}

# Production environment
provider "vaulted-null" {
  alias          = "prod"
  aws_kms_key_id = var.prod_kms_key_id
  aws_region     = "us-west-2"
}

# Use different providers for different secrets
data "vaulted-null_content" "dev_secret" {
  provider = vaulted-null.dev
  content  = var.dev_encrypted_secret
}

data "vaulted-null_content" "prod_secret" {
  provider = vaulted-null.prod
  content  = var.prod_encrypted_secret
}
```

### Integration with External Tools

The provider works well with:
- **Terraform Cloud/Enterprise** - For secure state management
- **GitHub Actions** - For CI/CD secret management
- **AWS Parameter Store** - Store encrypted payloads as parameters
- **Kubernetes Secrets** - Create secrets from decrypted content

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details on:
- Development setup
- Code style and standards
- Submitting pull requests
- Reporting issues

## 🔗 Related Projects

- [sumup-oss/vaulted](https://github.com/sumup-oss/vaulted) - The underlying encryption library
- [terraform-provider-vaulted](https://github.com/sumup-oss/terraform-provider-vaulted) - Alternative provider for HashiCorp Vault integration

