# FAQ

## How Mantil is different than (Cloudformation, SAM, CDK, Terraform, Serverless Framework)?

Most of the other frameworks are focused on building infrastructure. As that they are focused to the ops side. Mantil is dev tool.

Mantil is focused on the developer who uses Go to build backend service in the cloud. It aims to remove any ops work from the development workflow.

We are using those cloud frameworks internally. Currently Cloudformation, Terraform and AWS SDK. Our goal is not to fight with these great tools feature by feature but to build on top of them. In Mantil developers are writing pure Go code without any notion of AWS or Lambda. They don't need upfront understanding of complex AWS ecosystem and tooling.  

What makes Mantil different?
 * tailored for the Go developers
 * promotes cloud first build/deploy/test cycle
 * supports getting logs during the execution of lambda function, not after the function is completed
 * enables cloud first end to end testing
 * enables easy firing multiple project stages (deployments)

## What I need to start using Mantil?
Go, [Mantil cli](https://github.com/mantil-io/mantil#installation) and an AWS account credentials.  
Mantil is tool for Go developers so you need Go to build you APIs code into Lambda functions. You also need access to an AWS account. 

## Do I need my own AWS account?
Yes.  
In current version you are bringing you own AWS account. If you don't have one you should [create it](https://portal.aws.amazon.com/billing/signup#/start).

## How much does it cost to try Mantil?
Nothing.  
For trying Mantil you can for sure stay into [free tier](https://aws.amazon.com/free/?all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc&awsf.Free%20Tier%20Types=*all&awsf.Free%20Tier%20Categories=*all) of all the AWS services. When you create new AWS account you have pretty generous monthly limits for many services. Two most important you will for sure use with Mantil are Lambda functions and API Gateway. Here are their free tier monthly limits:

> The Amazon API Gateway free tier includes one million API calls received for REST APIs, one million API calls received for HTTP APIs, and one million messages and 750,000 connection minutes for WebSocket APIs per month for up to 12 months.

> The AWS Lambda free tier includes one million free requests per month and 400,000 GB-seconds of compute time per month.

Until you don't have some significant user base or you are not mining bitcoins in you Lambda function you will for sure stay into limits of free tier. So trying Mantil will cost you nothing. 

## What AWS account rights do I need to use Mantil? 
You need AWS account Admin rights for installing node into your AWS account. Mantil node install/uninstall phases are only time you need to provide AWS account credentials. 

Mantil is not storing your credentials, it is only used to setup a node into AWS account. After the node is installed all other communication is between Mantil command line and the node. Node functions have only necessary IAM permissions. All the resources created for Mantil node (API Gateway, Lambda function, IAM roles) have 'mantil-' prefix. You can list node resources by `mantil aws resources` command.

Mantil uninstall command will again need your credentials to remove all node resources. After the uninstall you account is in the original state. Mantil will remove anything it created. 

## Is there local development environment for Mantil?
No.  
In Mantil we chose to both develop and run production using cloud services. There is no copy of the cloud services for the local machine. Instead of trying to make copy of the cloud services for the local development we are making effort to get the feeling of the local development while using real services in the cloud. 

Under my experience having local development while using cloud services for production leads to supporting two environments. It is easy for trivial cases. But while the number of services or complexity of their interactions raise, as they always do in the real world, supporting two different environments becomes more and more painful. 

Developers like to have their own sandbox. With Mantil they have that private sandbox but instead of local it is using cloud resources. In this serverless world that means that development and production are essentially the same environment. It is not that one is using less capable servers, simplified network or something like that. All environments are the same!  
In Mantil we have concept of stage which is deployment of a project into cloud environment. By supporting infinite number of stages for each project development can be organized, besides private environment for each developer, into as many as needed integration, staging or show case stages. We make creating and deploying to the new stage a trivial step.

With Mantil you get all the benefits of the local development: isolated environment, instant feedback.   
While at the same time got other benefits:
 * no need to maintain two different environments
 * dev production and all the stages in between use exactly the same resources
 * Mantil handles everything no need to setup anything locally
 
But SAM, Serverless... have local development?  
Yes, but Mantil choose way to use cloud resources for both development and production. With the little change in mindset I believe that is long term right choice.

Questions I ask myself about team development environments:
 * how much time is needed to bootstrap new developer
   That developer can be a part time. To solve specific problem. Product manger who will edit texts. Designer who jump into project to make it usable. 
 * what is maintenance cost
   How much of his time developers spend building environment instead of business features.
 * how complex it is
   How many developers from the team are actually capable of extending development environment.
   What happens when initial developer leaves.
   

## How do I debug, can I set breakpoints in my function code?
No.  
When I started programming, many years ago, firing debugger and setting breakpoints was a way of life. That is especially convenient into Visual Studio or some other specialized IDE. The developers who are used to that kind of environments will feel unpleasant in any tool which doesn't support debug/breakpoint. I had the same feeling early in my career. After .Net I was developing in Ruby, Erlang, Go. In each of that environments I tried to setup some kind of breakpoint development style. That was short episodes and no one really useful. But I was not missing breakpoints. Breakpoint development is essentially a way to understand what is happening in the code. Once you have that mental model and can read code without need to fire debugger and go step by step you don't need it any more. Most of the experienced developers I know are asking for the breakpoint development style from the habit and from feeling insecure without it.  
Usually answer is that you don't need breakpoints just put some log lines to understand what is happening. My recommendation is to first build mental model about the code, then build test and to require that behavior from the code. Tests are repeatable and long lasting contract. Breakpoints are one time single mind explanation. 

Here is an ode to the debugging-less programming by the two legends in our filed. Quote is from "[The Best Programming Advice I Ever Got](http://www.informit.com/articles/article.aspx?p=1941206)" with Rob Pike:

> Ken taught me that thinking before debugging is extremely important. If you dive into the bug, you tend to fix the local issue in the code, but if you think about the bug first, how the bug came to be, you often find and correct a higher-level problem in the code that will improve the design and prevent further bugs.  
> I recognize this is largely a matter of style. Some people insist on line-by-line tool-driven debugging for everything. But I now believe that thinking—without looking at the code—is the best debugging tool of all, because it leads to better software.



<!--
+* usporedba s drugim alatima

+* Postoji li lokalna razvojna okolina - ne
+* Moram li imati svoj AWS account - da
+* Moram li imati prava na AWS-u - da, ali samo za install fazu, nakon toga vise ne treba, u buducim verzijama nece morati imati nikakva AWS prava napomenuti to
+* Koliko ce me kostatiti to na AWS-u - ma nista,
+* Sto moram imati na svom racunalu - mantil cli i Go, sve ostalo je u cloudu

* Sto ce Mantil kreirati na mom AWS accountu - popis za node, za project, objasniti naming, tagging
* Kako da znam koji su resursi kreirani od strane Mantila - objasniti naming, tagging
* Kako da znam sto se dogadja u mojoj lambda funkciji - invoke pokazuje logove
* Mogu li imati vise deploymenta jednog projekta

* Postoji li Visual Studio Code Mantil plugin
* Podrzava li Mantil Step Functions?
-->

