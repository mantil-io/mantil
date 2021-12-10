
# mantil test

Runs project tests

Project end to end tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.

### USAGE
<pre>
  mantil test [options]
</pre>
### OPTIONS
<pre>
  -r, --run string     Run only tests with this pattern in name
  -s, --stage string   Project stage to target instead of default
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
