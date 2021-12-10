
# mantil env

Exports project environment variables

Then you can use that variables in other shell comands.

Mantil project is determined by the current shell folder.
You can be anywhere in the project tree.

If not specified (--stage option) default project stage is used.

### USAGE
<pre>
  mantil env [options]
</pre>
### OPTIONS
<pre>
  -s, --stage string   Project stage to target instead of default
  -u, --url            Show only project API url
</pre>
### EXAMPLES
<pre>
  ==&gt; Set environment variables in terminal
  $ eval $(mantil env)

  ==&gt; Use current stage api url in other shell commands
  $ curl -X POST $(mantil env -url)/ping
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
