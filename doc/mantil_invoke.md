## mantil invoke

Invoke function methods through the project's API Gateway

### Synopsis

Invoke function methods through the project's API Gateway

This is a convenience method and provides similar output to calling:
curl -X POST https://<stage_api_url>/<function>[/method] [-d '<data>'] [-I]

Additionally, you can enable streaming of lambda execution logs by setting the --logs flag.

```
mantil invoke <function>[/method] [flags]
```

### Options

```
  -d, --data string    data for the method invoke request
  -i, --include        include response headers in the output
  -l, --logs           show lambda execution logs
  -s, --stage string   name of the stage to target
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil](mantil.md)	 - Makes serverless development with Go and AWS Lambda joyful

