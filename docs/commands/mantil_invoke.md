
# mantil invoke

Invokes API method on the project stage

Makes HTTP request to the gateway endpoint of the project stage. That invokes
Lambda function of that project api. If API method is not specified default
(named Default in Go code) is assumed.

Mantil project is determined by the current shell folder.
You can be anywhere in the project tree.
If not specified (--stage option) default project stage is used.

During lambda function execution their logs are shown in terminal. Each lambda
function log line is preffixed with λ symbol. You can hide that logs with the
--no-log option.

This is a convenience method and provides similar output to calling:
$ curl -X POST https://&lt;stage_endpoint_url&gt;/&lt;api&gt;[/method] [-d '&lt;data&gt;'] [-i]

### USAGE
<pre>
  mantil invoke &lt;api&gt;[/method] [options]
</pre>
### ARGUMENTS
<pre>
  &lt;api&gt;      Name of the API. Your APIs are in /api folder.
  [/method]  Method name in Go source code.
            Default method will called if not spedified.
</pre>
### OPTIONS
<pre>
  -d, --data string    Data for the method invoke request
  -i, --include        Include response headers in the output
  -n, --no-logs        Hide lambda execution logs
  -s, --stage string   Project stage to target instead of default
</pre>
### EXAMPLES
<pre>
  ==&gt; invoke Default method in Ping api
  $ mantil invoke ping
  200 OK
  pong

  ==&gt; invoke Hello method in Ping api with 'Mantil' data
  $ mantil invoke ping/hello -d 'Mantil'
  200 OK
  Hello, Mantil

  ==&gt; invoke ReqRsp method in Ping api with json data payload
  $ mantil invoke ping/reqrsp -d '{"name":"Mantil"}'
  200 OK
  {
     "Response": "Hello, Mantil"
  }
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
