# Testing in Mantil Project

## Unit tests

Your API's in Mantil are pure Go code. They don't have anything AWS or Lambda
specific. Mantil provides all infrastructure burden. Unit testing you API'a are
like unit testing any other Go struct.  
Our example project ping also provides example of a [trivial API
test](https://github.com/mantil-io/template-ping/blob/master/api/ping/ping_test.go).
It is there to show the idea of where and how to unit test API's.

## Integration tests

Integration tests are the category of tests that depend on some other
outside resources. In the other example project,
[excuses](https://github.com/mantil-io/template-excuses), we have examples of
both
[unit](https://github.com/mantil-io/template-excuses/blob/0a8c06a6d0d40fd4659c1538c772b7eaa8c7d5f5/api/excuses/excuses_test.go#L15)
and
[integration](https://github.com/mantil-io/template-excuses/blob/0a8c06a6d0d40fd4659c1538c772b7eaa8c7d5f5/api/excuses/excuses_test.go#L28)
tests. In unit we are mocking outside service with in process HTTP server. And in
integration we are using real URL from the internet. So holding your integration tests side-by-side with the unit or moving them to another place are both valid options. It really depends on project.


## End to end tests

Mantil project holds end to end tests in `/test` folder (from the project root).
[Here](https://github.com/mantil-io/template-ping/blob/master/test/ping_test.go)
is an example of an end to end test for our ping project. You can run it with the `mantil test`. It uses the current project stage to run HTTP request against
deployed API's. 

[httpexpect](https://github.com/gavv/httpexpect) is a great library for quickly
building requests and examining HTTP responses.



<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#documentation">↵ Back to Documentation Home!</a></p>
