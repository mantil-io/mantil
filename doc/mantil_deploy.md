## mantil deploy

Deploys updates to stages

### Synopsis

Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage flag accepts any existing stage and defaults to the default stage if omitted.

```
mantil deploy [flags]
```

### Options

```
  -s, --stage string   the name of the stage to deploy to
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil](mantil.md)	 - Makes serverless development with Go and AWS Lambda joyful

