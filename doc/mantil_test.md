## mantil test

Run project integration tests

### Synopsis

Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.


```
mantil test [flags]
```

### Options

```
  -r, --run string     run only tests with this pattern in name
  -s, --stage string   stage name
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil](mantil.md)	 - Makes serverless development with Go and AWS Lambda joyful

