# Mantil CLI

Mantil CLI is the tool used for setting up mantil on your AWS account and managing your project workflow.

## Configuring AWS profile

In order to bootstrap mantil on your AWS account you first need to configure your AWS profile. This can be done either by manually configuring necessary files on your local machine or by using AWS CLI. Detailed [instructions](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) are provided in AWS documentation.
AWS user the profile belongs to needs to have Administration permissions.

## Bootstraping

Once you've configured the AWS profile, the next step is to boostrap mantil backend on your AWS account. This will setup necessary infrastructure which will be used to manage your mantil projects.

Running `mantil bootstrap` will first ask you to choose the aws profile you want to boostrap mantil on and then create boostraping lambda on your account which will do the rest.

After bootstraping has finished successfully it will create `backend.json` in `~/.mantil` directory. This file will contain URL of the api gateway your CLI will use to communicate with mantil backend.

Command can also be run with flag `destroy` which will destroy all resources created by bootstrap process.

## Working on project

Now you can start creating mantil projects and working on them using various features of the CLI.

### init

Command `init` is used for initialisation of the project.

This command requires your local machine to have access to your Github account since it also initialises Github project for you with our project template. This can be done either by setting up `GITHUB_TOKEN` in your env variables or by having [github cli](https://github.com/cli/cli) installed on your local machine which will then be used to ask you for the credentials.

`init` command will first ask you for the name of the project and, optionally, github organisation you want the project to be initialised in. If github organisation is not provided project will be initialised in your personal account.

After project initialisation directory with the name of your project will be created in `~/.mantil` directory. This directory will contain config file with the token generated for your project. This token is used to authenticate you while making changes on your project.

Additionally, github project will contain github workflow which will automatically deploy any changes you push to your github repository. (**currently doesn't work**)

Template your project is initialised with shows project structure necessary to work with mantil. All lambda functions need to be placed in `functions` directory, in directory with their name which contains `main.go` file. This file is the entry point of your function. In most cases, it will only be used to initialise function's API and pass it to our lambda handler wrapper.

APIs of the functions are placed in the `api` directory. Each public method on the API struct will be open through API and can be invoked separately. If endpoint is not provided during request `Invoke` function will be invoked as default.

### deploy

Command `deploy` is used for deploying changes to your project.

This command will check if there are any changes in your project by comparing hashes of the functions' binaries and perform deploy if necessary. It will also create and destroy any functions you've created or deleted in your project.

### invoke

`deploy` will also output URL of the api gateway for your project. You can use this URL directly in your frontend applications or through terminal with `curl` command.

Alternatively, you can use `invoke` command in your project to make requests without having to deal with URL of the api gateway.

Running command such as `mantil invoke hello/world -d '{"name": "john"}'` will invoke endpoint `world` in function `hello` with provided data.

### generate

Command `generate` is provided to simplify generation of boilerplate code.

Subcommand `api` is used to create additional function in your project. It will create necessary files in `functions` and `api` directories with boilerplate code that is then easy to adapt to your needs.

Additionally, if you want to create more function methods in your function you can provide them through `methods` flag.

Command `mantil generate api hello -m foo,bar` will generate code for lambda hello which will contain default `Root` method, as well as `Foo` and `Bar` methods.

### watch

Command `watch` is provided to simplify process of deploy without having to manually run deploy command on each change.

Running `mantil watch` inside your project's directory will start a watcher which will run deploy on each change of the go files in your project.

Additionally, you can whitelist functions which will be deployed on changes through `functions` flag.

Command `mantil watch -f hello` will start a watcher which will deploy only `hello` function if something changes.

### destroy

Command `destroy` can be run inside your project's directory to destroy your project's infrastructre on the AWS and delete your local and Github project repository. It will ask you to input your project's name as a precaution.
