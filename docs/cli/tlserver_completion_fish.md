## tlserver completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	tlserver completion fish | source

To load completions for every new session, execute once:

	tlserver completion fish > ~/.config/fish/completions/tlserver.fish

You will need to start a new shell for this setup to take effect.


```
tlserver completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [tlserver completion](tlserver_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 17-Aug-2024