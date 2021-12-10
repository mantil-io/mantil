
# mantil aws uninstall

Uninstalls Mantil node from AWS account

Command will remove node from AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run option which will show you what credentials will be used
and what account will be managed by command.

By default you will be asked to confirm the destruction.
This behaviour can be disabled using the --yes option.

### USAGE
<pre>
  mantil aws uninstall [node-name] [options]
</pre>
### ARGUMENTS
<pre>
  [node-name]  Mantil node name.
               If not provided default name dev will be used for destruction.
</pre>
### OPTIONS
<pre>
      --aws-access-key-id string       Access key ID for the AWS account, must be used with the
                                       aws-secret-access-key and aws-region options
      --aws-env                        Use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION
                                       environment variables for AWS authentication
      --aws-profile string             Use the given profile for AWS authentication
      --aws-region string              Region for the AWS account, must be used with and aws-access-key-id and
                                       aws-secret-access-key options
      --aws-secret-access-key string   Secret access key for the AWS account, must be used with the
                                       aws-access-key-id and aws-region options
      --dry-run                        Don't start install/uninstall just show what credentials will be used
      --force                          Force uninstall even if it will result in orphaned stages
  -y, --yes                            Assume 'yes' as answer to all prompts
</pre>
### EXAMPLES
<pre>
  You must provide credentials for Mantil to access your AWS account.
  There are three ways to provide credentials.

  ==&gt; specifiy access keys as arguments:
  $ mantil aws uninstall --aws-access-key-id=AKIAIOSFODNN7EXAMPLE \
                       --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
                       --aws-region=us-east-1

  ==&gt; read access keys from environment variables:
  $ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
  $ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  $ export AWS_DEFAULT_REGION=us-east-1
  $ mantil aws uninstall --aws-env

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

  ==&gt; use your named AWS profile form ~/.aws/config
  $ mantil aws uninstall --aws-profile=my-named-profile

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html
</pre>
### GLOBAL OPTIONS
<pre>
      --help       Show command help
      --no-color   Don't use colors in output
</pre>
### LEARN MORE
<pre>
  Visit https://github.com/mantil-io/docs to learn more.
  For further support contact us at support@mantil.com.
</pre>
