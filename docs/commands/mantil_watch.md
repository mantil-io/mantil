
# mantil watch

Watches for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any *.go files
in the project directory and automatically deploys changes to the stage.

You can set a method to invoke after every deploy using the --method and --data options.
Or run tests after every deploy with --test options.

### USAGE
<pre>
  mantil watch [options]
</pre>
### OPTIONS
<pre>
  -d, --data string     Data for the method invoke request
  -m, --method string   Method to invoke after deploying changes
  -s, --stage string    Project stage to target instead of default
  -t, --test            Run tests after deploying changes
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
