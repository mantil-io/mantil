package texts

import (
	"fmt"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
)

type Command struct {
	Short     string
	Long      string
	Arguments string
	NextSteps string
	Examples  string
}

func setupExamples(commandName string) string {
	return strings.ReplaceAll(`  You must provide credentials for Mantil to access your AWS account.
  There are three ways to provide credentials.

  ==> specifiy access keys as arguments:
  $ mantil aws {.CommandName} --aws-access-key-id=AKIAIOSFODNN7EXAMPLE \
                       --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
                       --aws-region=us-east-1

  ==> read access keys from environment variables:
  $ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
  $ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  $ export AWS_DEFAULT_REGION=us-east-1
  $ mantil aws {.CommandName} --aws-env

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

  ==> use your named AWS profile form ~/.aws/config
  $ mantil aws {.CommandName} --aws-profile=my-named-profile

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html`, "{.CommandName}", commandName)
}

var AwsInstall = Command{
	Short: "Install Mantil node into AWS account",
	Long: `Install Mantil node into AWS account

Command will install node into AWS account. Node consists of few Lambda function, API Gateway and S3 bucket.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run option which will show you what credentials will be used
and what account will be managed by command.`,
	Arguments: `
  [node-name]  Mantil node name.
               If not provided default name dev will be used for the first node.`,
	NextSteps: `
* Run 'mantil new' to start a project from scratch or choose from an existing template.
Check documentation at https://github.com/mantil-io/docs for additional inspiration.
`,
	Examples: setupExamples("install"),
}

var AwsUninstall = Command{
	Short: "Uninstall Mantil node from AWS account",
	Long: `Uninstall Mantil node from AWS account

Command will remove node from AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run option which will show you what credentials will be used
and what account will be managed by command.

By default you will be asked to confirm the destruction.
This behaviour can be disabled using the --force option.`,
	Arguments: `
  [node-name]  Mantil node name.
               If not provided default name dev will be used for the first node.`,
	NextSteps: `
* We are sorry to see you go. Help us make Mantil better by letting us know what you didn’t like at support@mantil.com.
`,
	Examples: setupExamples("uninstall"),
}

var Env = Command{
	Short: "Export project environment variables",
	Long: `Export project environment variables
for use in other shell commands.

Mantil project is determined by the current shell folder. It can be anywhere in
the project tree.
If not specified (--stage option) default project stage is used.`,
	Examples: `  ==> Set environment variables in terminal
  $ eval $(mantil env)

  ==> Use current stage api url in other shell commands
  $ curl -X POST $(mantil env -url)/ping`,
}

var Invoke = Command{
	Short: "Invoke API method on the project stage",
	Long: `Invoke API method on the project stage

Makes HTTP request to the gateway endpoint of the project stage. That invokes
lambda function of that project api. If api method is not specified default
(named Default in Go code) is assumed.

Mantil project is determined by the current shell folder. It can be anywhere in
the project tree.
If not specified (--stage option) default project stage is used.

During lambda function execution their logs are shown in terminal. Each lambda
function log line is preffixed with λ symbol. You can hide that logs with the
--no-log option.

This is a convenience method and provides similar output to calling:
$ curl -X POST https://<stage_endpoint_url>/<api>[/method] [-d '<data>'] [-i]`,
	Examples: `==> invoke Default method in Ping api
$ mantil invoke ping
200 OK
pong

==> invoke Hello method in Ping api with 'Mantil' data
$ mantil invoke ping/hello -d 'Mantil'
200 OK
Hello, Mantil

==> invoke ReqRsp method in Ping api with json data payload
$ mantil invoke ping/reqrsp -d '{"name":"Mantil"}'
200 OK
{
   "Response": "Hello, Mantil"
}`,
}

var Logs = Command{
	Short: "Fetch logs for a specific API",
	Long: `Fetch logs for a specific API

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail option is set the process will keep running and polling for new logs every second.`,
}

var New = Command{
	Short: "Create a new Mantil project",
	Long: `Create a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from option.
The source can either be an existing git repository or one of the predefined templates:
ping    - https://github.com/mantil-io/template-ping
excuses - https://github.com/mantil-io/template-excuses
chat    - https://github.com/mantil-io/template-chat

If no source is provided it will default to the template "ping".

By default, the go module name of the initialized project will be the project name.
This can be changed by setting the --module-name option.`,
	NextSteps: `
* It's time to start developing in the cloud. Run 'mantil stage new' to
create your first development environment or check the documentation at
https://docs.mantil.io for more details.
`,
}

var Test = Command{
	Short: "Run project tests",
	Long: `Run project tests

Project end to end tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.
`,
}

var Watch = Command{
	Short: "Watch for file changes and automatically deploy them",
	Long: `Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage option.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test options.`,
}

var Stage = Command{
	Short: "Manage project stages",
	Long: `Manage project stages

A stage represents a named deployment of the project. Each stage creates a set of resources
which can be managed and configured separately.

Stages can be deployed to any node in the workspace.`,
}

var StageNew = Command{
	Short: "Create a new stage",
	Long: `Create a new stage

This command will create a new stage with the given name. If the name is left empty it will default to "dev".

If only one node is set up in the workspace, the stage will be deployed to that node by default.
Otherwise, you will be asked to pick a node. The node can also be specified via the --node option.`,
	NextSteps: `
* Try 'mantil invoke' to see your fully functional Mantil serverless application in action.
`,
}

var StageDestroy = Command{
	Short: "Destroy a stage",
	Long: `Destroy a stage

This command will destroy all resources belonging to a stage.
Optionally, you can set the --all option to destroy all stages.

By default you will be asked to confirm the destruction.
This behavior can be disabled using the --force option.`,
}

var StageList = Command{
	Short: "List stages in project",
}

var StageUse = Command{
	Short: "Set default project stage",
}

var Generate = Command{
	Short: "Automatically generate code in the project",
}

var GenerateApi = Command{
	Short: "Generate Go code for a new API",
	Long: `Generate Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods option. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello`,
}

var Deploy = Command{
	Short: "Deploy project updates to a stage",
	Long: `Deploy project updates to a stage

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage option accepts any existing stage and defaults to the default stage if omitted.`,
	NextSteps: `
* Use 'mantil logs' to see those directly in terminal in an instant.
`,
}

var User = Command{
	Short: "Manages Mantil registration",
	Long: `Manages Mantil registration

Mantil is in early beta and access is granted only to registered users. To
register you will be asked to provide your email address where the activation
token will be sent.`,
}

var UserRegister = Command{
	Short: "Initiate Mantil registration",
	Long: `Mantil is in early beta and access is granted only to registered users. This
command initiates the signup process for Mantil application.`,
}

var UserActivate = Command{
	Short: "Finalize Mantil registration",
	Long: `Mantil is in early beta and access is granted only to registered users. With the
activation token received in your email this command finalizes Mantil
registration.`,
}

func logsDir() string {
	logsDir, _ := log.LogsDir()
	return logsDir
}

var Report = Command{
	Short: "Make a bug report",
	Long: fmt.Sprintf(`Make a bug report

Mantil logs are located at %s.

This command sends us those log files so we can analyze them and help you with
the issue you're having.

By default last 3 days of logs are included, you can change that with --days option.`, logsDir()),
}
