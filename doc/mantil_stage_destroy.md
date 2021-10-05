## mantil stage destroy

Destroy a stage

### Synopsis

Destroy a stage

This command will destroy all resources belonging to a stage.
Optionally, you can set the --all flag to destroy all stages.

By default you will be asked to confirm the destruction by typing in the project name.
This behavior can be disabled using the --force flag.

```
mantil stage destroy <name> [flags]
```

### Options

```
      --all     destroy all stages
      --force   don't ask for confirmation
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil stage](mantil_stage.md)	 - Manage project stages

