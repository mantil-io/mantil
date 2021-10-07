## mantil generate api

Generate Go code for a new API

### Synopsis

Generate Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods flag. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello

```
mantil generate api <function> [flags]
```

### Options

```
  -m, --methods strings   additional function methods, if left empty only the Default method will be created
```

### Options inherited from parent commands

```
      --help       show command help
      --no-color   don't use colors in output
```

### SEE ALSO

* [mantil generate](mantil_generate.md)	 - Automatically generate code in the project

