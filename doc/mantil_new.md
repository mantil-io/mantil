## mantil new

Initializes a new Mantil project

### Synopsis

Initializes a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from flag.
The source can either be an existing git repository or one of the predefined templates:
excuses - https://github.com/mantil-io/template-excuses
ping - https://github.com/mantil-io/go-mantil-template

If no source is provided it will default to the template "ping".

By default, the go module name of the initialized project will be the project name.
This can be changed by setting the --module-name flag.

```
mantil new <project> [flags]
```

### Options

```
      --from string          name of the template or URL of the repository that will be used as one
      --module-name string   replace module name and import paths
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil](mantil.md)	 - Makes serverless development with Go and AWS Lambda joyful

