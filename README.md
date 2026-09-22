# AWS SecretsManager Plugin

This is the `awssecretsmanager` implementation plugin built using the go-plugin architecture from hashicorp, it is used by the [ConfigManager](https://github.com/DevLabFoundry/configmanager) service.

## Token Prefix

This plugin uses the `AWSSECRETS` token prefix.

## Configuration

The plugin supports the following metadata configuration:

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `version` | string | The version stage of the secret to retrieve | `AWSCURRENT` |

### Example Usage

```yaml
# Basic secret retrieval
MY_SECRET: "AWSSECRETS#/my/secret/path"

# With specific version
MY_SECRET_VERSIONED: "AWSSECRETS#/my/secret/path[version=AWSPREVIOUS]"
```

## AWS Authentication

The plugin uses the default AWS SDK credential chain. Ensure your environment has appropriate AWS credentials configured via:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- Shared credentials file (`~/.aws/credentials`)
- IAM role (when running on AWS infrastructure)
- AWS SSO configuration
