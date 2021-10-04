## mantil aws uninstall

Uninstall Mantil from AWS account

### Synopsis

Uninstall Mantil from AWS account

Command will remove backend services from AWS account.
You must provide credentials for Mantil to access your AWS account.
There are few ways to provide credentials:

1. specifiy access keys as arguments:
   $ mantil aws install --aws-access-key-id=AKIAIOSFODNN7EXAMPLE --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --aws-region=us-east-1

2. read access keys from environment variables:
   $ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
   $ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
   $ export AWS_DEFAULT_REGION=us-east-1
   $ mantil aws install --aws-env

reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

3. use your named AWS profile form ~/.aws/config
   $ mantil aws install --aws-profile=my-named-profile

reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html

Argument account-name is Mantil account reference.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.


```
mantil aws uninstall [account-name] [flags]
```

### Options

```
      --aws-access-key-id string       access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags
      --aws-env                        use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication
      --aws-profile string             use the given profile for AWS authentication
      --aws-region string              region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags
      --aws-secret-access-key string   secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags
      --dry-run                        don't start install/uninstall just show what credentials will be used
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil aws](mantil_aws.md)	 - AWS account subcommand

