terraform {
  required_providers {
    vaulted-null = {
      source = "registry.terraform.io/syndbg/vaulted-null"
      # No version constraint needed with dev_overrides
    }
  }
}

provider "vaulted-null" {
  private_key_path = "./private.pem"
  public_key_path  = "./public.pem"
}

data "vaulted-null_content" "example" {
  content = "$VED;1.0::ldZuDi3YIqYqMjp1v2gNRrLzKSECydQwMJDeUQRTNElDY6xRQsaPTv6Z+alKi4srQ4chRVPoNhsSe9NoxviofxgvjCkTlpJ6lybaBUABtwBUtQ9mMmT1dnMvhfFJIaSqHjBKoh0E+SnNZC9e06dWtmBmHh5m+qlJn7hswVZ8PmY90wjfN1xGmIe9NdY2pm7eLQaI7YMZiVAgnXozbwvAvdInFkPmlyBHjTWJvX+zf3R8tgDUe1VWSxXOo67Sg9+ejVZvE+/qmgLAsbCfhLivtbvjlhsMiK6BcsaAZS6JOJMuguwF4t/iX9Asm//6tnSWX9hu1fnVkd2AKzaNJ/mPLw==::2+6jqW7y9eZ23ztQ02FV/us25shwLFYQx1AcWm46gw1aTfTT"
}

resource "vaulted-null_encrypt_content" "example" {
  plaintext = "example3"
}

output "example" {
  value     = data.vaulted-null_content.example.decrypted
  sensitive = true
}

output "new-encrypted-example" {
  value = vaulted-null_encrypt_content.example.encrypted
}

# Ephemeral resources examples
ephemeral "vaulted-null_encrypt" "sensitive_data" {
  plaintext = "This is sensitive data that should be encrypted"
}

# Demonstrate how to work with ephemeral resources
# Note: Ephemeral values can only be used in write-only attributes or passed to other ephemeral resources

# Example: Using ephemeral decrypt to get decrypted content back
ephemeral "vaulted-null_decrypt" "verify_encryption" {
  content = ephemeral.vaulted-null_encrypt.sensitive_data.encrypted
}

# Note: Ephemeral values cannot be output directly in root modules
# They are only available during the Terraform run and can be passed to other resources
