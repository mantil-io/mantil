
# Anonymous Analytics

Every execution of the Mantil CLI sends an [CliCommand](https://github.com/mantil-io/mantil/blob/4ef981e9c89025f3ebcd3937b4872071caafb80e/domain/event.go#L22) piece of data. It contains metrics about the current Mantil workspace, stage, project and series of [Events](https://github.com/mantil-io/mantil/blob/4ef981e9c89025f3ebcd3937b4872071caafb80e/domain/event.go#L35) which are command-specific. Look into the definition of the [Event](https://github.com/mantil-io/mantil/blob/4ef981e9c89025f3ebcd3937b4872071caafb80e/domain/event.go#L109) to get the feeling about what data we collect.

For example, for [Deploy](https://github.com/mantil-io/mantil/blob/4ef981e9c89025f3ebcd3937b4872071caafb80e/domain/event.go#L128) command, we collect metrics about how many Lambda functions are added, updated and removed during the deploy command execution. Further, we collect durations of the build, upload and update phases. How many bytes were transferred to the S3 bucket and whether it was just function updates or we made some infrastructure changes (new function, API gateway).

Events are collected anonymously with the primary purpose of helping us understand how people use Mantil and allowing us to prioritize fixes and features as well as to catch errors as soon as those happen. We are a small team with the product in early beta and would really appreciate keeping the events collection on. However, if you are working on a super-secret government project, there is an option to disable them by setting [MANTIL_NO_EVENTS](https://github.com/mantil-io/mantil/blob/5d0ee4a609a63821eb319776c9981af6e0df4049/domain/workspace.go#L33) environment variable. Internally we use it just when running integration tests.

We also pay special attention to the kind of data we are collecting, so,  e.g., if you put your AWS credentials into the command line, we recognize that and [remove](https://github.com/mantil-io/mantil/blob/4ef981e9c89025f3ebcd3937b4872071caafb80e/domain/event.go#L213) them.


<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>
