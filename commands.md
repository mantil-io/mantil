## mantil --version
```
mantil version v0.1.10-52-g3221e7f
```

## mantil --help
```
Makes serverless development with Go and AWS Lambda joyful

Usage:
  mantil [command]

Available Commands:
  aws         AWS account subcommand
  completion  generate the autocompletion script for the specified shell
  deploy      Deploys updates to stages
  env         Show project environment variables
  generate    Automatically generate code in the project
  help        Help about any command
  invoke      Invoke function methods through the project's API Gateway
  logs        Fetch logs for a specific function/api
  new         Initializes a new Mantil project
  stage       Manage project stages
  test        Run project integration tests
  watch       Watch for file changes and automatically deploy them

Flags:
      --help       show command help
      --no-color   don't use colors in output
      --version    show mantil version

Use "mantil [command] --help" for more information about a command.
```

## mantil aws --help
```
AWS account subcommand

Usage:
  mantil aws [command]

Available Commands:
  install     Install Mantil into AWS account
  uninstall   Uninstall Mantil from AWS account

Global Flags:
      --help       show command help
      --no-color   don't use colors in output

Use "mantil aws [command] --help" for more information about a command.
```

## mantil aws install --help
```
Install Mantil into AWS account

Command will install backend services into AWS account.
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

Argument account-name is for referencing that account in Mantil.
If not provided default name dev will be used for the first account.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.

Usage:
  mantil aws install [account-name] [flags]

Flags:
      --aws-access-key-id string       access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags
      --aws-env                        use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication
      --aws-profile string             use the given profile for AWS authentication
      --aws-region string              region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags
      --aws-secret-access-key string   secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags
      --dry-run                        don't start install/uninstall just show what credentials will be used
      --override                       force override access tokens on already installed account

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil aws uninstall --help
```
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

Usage:
  mantil aws uninstall [account-name] [flags]

Flags:
      --aws-access-key-id string       access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags
      --aws-env                        use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication
      --aws-profile string             use the given profile for AWS authentication
      --aws-region string              region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags
      --aws-secret-access-key string   secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags
      --dry-run                        don't start install/uninstall just show what credentials will be used

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil stage --help
```
Manage project stages

A stage represents a named deployment of the project. Each stage creates a set of resources
which can be managed and configured separately.

Stages can be deployed to any account in the workspace.

Usage:
  mantil stage [command]

Available Commands:
  destroy     Destroy a stage
  new         Create a new stage

Global Flags:
      --help       show command help
      --no-color   don't use colors in output

Use "mantil stage [command] --help" for more information about a command.
```

## mantil stage new --help
```
Create a new stage

This command will create a new stage with the given name. If the name is left empty it will default to "dev".

If only one account is set up in the workspace, the stage will be deployed to that account by default.
Otherwise, you will be asked to pick an account. The account can also be specified via the --account flag.

Usage:
  mantil stage new <name> [flags]

Flags:
  -a, --account string   account in which the stage will be created

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil stage destroy --help
```
Destroy a stage

This command will destroy all resources belonging to a stage.
Optionally, you can set the --all flag to destroy all stages.

By default you will be asked to confirm the destruction by typing in the project name.
This behavior can be disabled using the --force flag.

Usage:
  mantil stage destroy <name> [flags]

Flags:
      --all     destroy all stages
      --force   don't ask for confirmation

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil new --help
```
Initializes a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from flag.
The source can either be an existing git repository or one of the predefined templates:
ping - https://github.com/mantil-io/go-mantil-template
excuses - https://github.com/mantil-io/template-excuses

If no source is provided it will default to the template "ping".

By default, the go module name of the initialized project will be the project name.
This can be changed by setting the --module-name flag.

Usage:
  mantil new <project> [flags]

Flags:
      --from string          name of the template or URL of the repository that will be used as one
      --module-name string   replace module name and import paths

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil deploy --help
```
Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage flag accepts any existing stage and defaults to the default stage if omitted.

Usage:
  mantil deploy [flags]

Flags:
  -s, --stage string   the name of the stage to deploy to

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil env --help
```
Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)

Usage:
  mantil env [flags]

Flags:
  -s, --stage string   stage name
  -u, --url            show only project api url

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil invoke --help
```
Invoke function methods through the project's API Gateway

This is a convenience method and provides similar output to calling:
curl -X POST https://<stage_api_url>/<function>[/method] [-d '<data>'] [-I]

Additionally, you can enable streaming of lambda execution logs by setting the --logs flag.

Usage:
  mantil invoke <function>[/method] [flags]

Flags:
  -d, --data string    data for the method invoke request
  -i, --include        include response headers in the output
  -l, --logs           show lambda execution logs
  -s, --stage string   name of the stage to target

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil logs --help
```
Fetch logs for a specific function/api

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail flag is set the process will keep running and polling for new logs every second.

Usage:
  mantil logs [function] [flags]

Flags:
  -p, --filter-pattern string   filter pattern to use
  -s, --since duration          from what time to begin displaying logs, default is 3 hours ago (default 3h0m0s)
      --stage string            name of the stage to fetch logs for
  -t, --tail                    continuously poll for new logs

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil test --help
```
Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.

Usage:
  mantil test [flags]

Flags:
  -r, --run string     run only tests with this pattern in name
  -s, --stage string   stage name

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil watch --help
```
Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage flag.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test flags.

Usage:
  mantil watch [flags]

Flags:
  -d, --data string     data for the method invoke request
  -m, --method string   method to invoke after deploying changes
  -s, --stage string    name of the stage to deploy changes to
  -t, --test            run tests after deploying changes

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

## mantil generate --help
```
Automatically generate code in the project

Usage:
  mantil generate [command]

Available Commands:
  api         Generate Go code for a new API

Global Flags:
      --help       show command help
      --no-color   don't use colors in output

Use "mantil generate [command] --help" for more information about a command.
```

## mantil generate api --help
```
Generate Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods flag. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello

Usage:
  mantil generate api <function> [flags]

Flags:
  -m, --methods strings   additional function methods, if left empty only the Default method will be created

Global Flags:
      --help       show command help
      --no-color   don't use colors in output
```

