### Unit tests

Your API's in Mantil are pure Go code. They don't have anything AWS or Lambda
specific. Mantil provides all infrastructure burden. Unit testing you API'a are
like unit testing any other Go struct.  
Our example project ping provides also example of [trivial API
test](https://github.com/mantil-io/template-ping/blob/master/api/ping/ping_test.go).
It is there to show idea of where and how to unit test API's.

### Integration tests

Integration tests are category of tests which are depending on some other
outside resources. In the other example project,
[excuses](https://github.com/mantil-io/template-excuses), we have example of
both
[unit](https://github.com/mantil-io/template-excuses/blob/0a8c06a6d0d40fd4659c1538c772b7eaa8c7d5f5/api/excuses/excuses_test.go#L15)
and
[integration](https://github.com/mantil-io/template-excuses/blob/0a8c06a6d0d40fd4659c1538c772b7eaa8c7d5f5/api/excuses/excuses_test.go#L28)
test. In unit we are mocking outside service with in process HTTP server. And in
integration we are using real URL from the internet. Holding your integration
tests side by side with unit or moving them to some other place are both valid
options. It really depends on project.


### End to end tests

Mantil project holds end to end tests in `/test` folder (from the project root).
[Here](https://github.com/mantil-io/template-ping/blob/master/test/ping_test.go)
is example of an end to end test for our ping project. You can run it with
`mantil test`. It uses current project stage to run HTTP request against
deployed API's. 

[httpexpect](https://github.com/gavv/httpexpect) is great library for quickly
building requests and examining HTTP responses.
