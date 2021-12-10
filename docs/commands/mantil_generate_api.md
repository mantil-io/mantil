
# mantil generate api

Generates Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods option. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello

### USAGE
<pre>
  mantil generate api &lt;name&gt; [options]
</pre>
### ARGUMENTS
<pre>
  &lt;name&gt;      Name of the API to generate.
</pre>
### OPTIONS
<pre>
  -m, --methods strings   Additional function methods, if left empty only the Default method will be created
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
