## mantil stage new

Create a new stage

### Synopsis

Create a new stage

This command will create a new stage with the given name. If the name is left empty it will default to "dev".

If only one account is set up in the workspace, the stage will be deployed to that account by default.
Otherwise, you will be asked to pick an account. The account can also be specified via the --account flag.

```
mantil stage new <name> [flags]
```

### Options

```
  -a, --account string   account in which the stage will be created
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil stage](mantil_stage.md)	 - Manage project stages

