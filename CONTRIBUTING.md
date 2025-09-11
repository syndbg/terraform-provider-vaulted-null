#  Contributing

## Prerequisites

* Golang 1.25.x or use https://github.com/syndbg/goenv to use correct go version.
* (To run linter) https://github.com/golangci/golangci-lint in `$PATH`.

## Common commands

### Building the provider

```shell
> go build -o terraform-provider-vaulted-null
```

### Running only unit tests with coverage

```shell
> go run mage.go -v unitTests
```

### Running the linter

```shell
> go run mage.go -v lint
```

## Local Development and Testing

### Setting up local provider development

For active development, use development overrides to avoid copying binaries:

1. **Create a development Terraform configuration file** (e.g., `examples/.terraformrc`):
   ```hcl
   provider_installation {
     # Development overrides - points to directory containing your built binary
     dev_overrides {
       "registry.terraform.io/syndbg/vaulted-null" = "/absolute/path/to/project-directory"
     }

     # For all other providers, install them directly from their origin provider registries
     direct {}
   }
   ```

2. **Build and test workflow**:
   ```shell
   # Build the provider
   go build -o terraform-provider-vaulted-null

   # Set configuration and test (skip terraform init with dev_overrides)
   export TF_CLI_CONFIG_FILE="$(pwd)/examples/.terraformrc"
   cd examples
   terraform plan
   terraform apply
   ```

### Using the example configuration

The repository includes working examples with pre-generated RSA keys:
- `examples/private.pem` - RSA private key for decryption
- `examples/public.pem` - RSA public key for encryption
- `examples/main.tf` - Example configuration with encrypted content

The example includes a pre-encrypted payload that will decrypt to "example3" when using the included keys.

### Alternative: Environment variables

You can also use environment variables instead of key files:
```shell
export VAULTED_PRIVATE_KEY_CONTENT="$(cat examples/private.pem)"
export VAULTED_PUBLIC_KEY_CONTENT="$(cat examples/public.pem)"
```

### Development tips

- With `dev_overrides`, **skip `terraform init`** - Terraform will warn you about this
- Use absolute paths in `.terraformrc` (avoid `~` which may not expand correctly)
- `dev_overrides` path should point to the **directory** containing the binary, not the binary file itself
- Remove any existing `.terraform/` directory and `.terraform.lock.hcl` when switching to dev_overrides
- The development warning in terraform output is normal and expected
- After code changes, just rebuild and run `terraform plan/apply` directly

## Workflows

### Submitting an issue

1. Check existing issues and verify that your issue is not already submitted.
   If it is, it's highly recommended to add  to that issue with your reports.
1. Open issue
1. Be as detailed as possible - `go` version, what did you do,
   what did you expect to happen, what actually happened.

### Submitting a PR

1. Find an existing issue to work on or follow `Submitting an issue` to create one
   that you're also going to fix.
   Make sure to notify that you're working on a fix for the issue you picked.
1. Branch out from latest `master`.
1. Code, add, commit and push your changes in your branch.
1. Make sure that tests and linter(s) pass locally for you.
1. Submit a PR.
1. Collaborate with the codeowners/reviewers to merge this in `master`.

### Releasing

GitHub Actions are used as a backbone to get releases going.

#### Rules

1. Releases are only created from `main`.
1. `main` is meant to be stable, so before tagging and pushing a tag, make sure that the CI checks pass.
1. Releases are GitHub releases.
1. Releases are following *semantic versioning*.
1. Releases are to be named in pattern of `vX.Y.Z`. The produced binary artifacts contain the `vX.Y.Z` in their names.
1. Changelog must up-to-date with what's going to be released. Check [CHANGELOG](./CHANGELOG.md).

#### Flow

1. Create a new GitHub a new tag from `main`
1. Push it to the remote git repository.
1. Wait for GitHub action workflow to finish
