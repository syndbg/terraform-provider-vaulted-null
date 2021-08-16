# terraform-provider-vaulted-null

Terraform provider utilizing [sumup-oss/vaulted](https://github.com/sumup-oss/vaulted) to provide a data source
 able to decrypt a vaulted encrypted payload. 

Are you using HashiCorp Vault? Perhaps [terraform-provider-vaulted](https://github.com/sumup-oss/terraform-provider-vaulted)
 is going to be useful to you.

Which one to use?

* terraform-provider-vaulted-null is meant to be used with remote/non-local encryption-at-transit Terraform state providers like Terraform Cloud. 
  Perfect for Terraform Cloud workspace agents/executors and trusted CI environments.
  The encrypted payload is decrypted via the data source, therefore it is stored in **plaintext in the Terraform State**.
* terraform-provider-vaulted is meant for less secure CI environments. E.g "public cloud" CI agents/executors. 
  It provides Terraform resources provisioning HashiCorp Vault with a vaulted encrypted payload. 
  The encrypted payload **is never stored in plaintext in the Terraform State**.

## Usage

Check out the [examples' main.tf](./examples/main.tf).

## Contributing

### Build

Run the following command to build the provider

```shell
go build 
```

