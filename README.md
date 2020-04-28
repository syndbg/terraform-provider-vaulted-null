# Terraform Provider Hashicups

Run the following command to build the provider

```shell
go build -o terraform-provider-hashicups
```

## Test sample configuration

First, build the provider in `test` directory.

```shell
go build -o test/terraform-provider-hashicups
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```