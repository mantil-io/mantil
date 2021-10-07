## mantil logs

Fetch logs for a specific function/api

### Synopsis

Fetch logs for a specific function/api

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail flag is set the process will keep running and polling for new logs every second.

```
mantil logs [function] [flags]
```

### Options

```
  -p, --filter-pattern string   filter pattern to use
  -s, --since duration          from what time to begin displaying logs, default is 3 hours ago (default 3h0m0s)
      --stage string            name of the stage to fetch logs for
  -t, --tail                    continuously poll for new logs
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil](mantil.md)	 - Makes serverless development with Go and AWS Lambda joyful

