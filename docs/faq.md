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

* The one about AWS Console - use it for exploring, use other repeatable tool for modifiying

* In what AWS Regions is Mantil supported?
-->

