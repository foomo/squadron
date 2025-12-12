---
title: "squadron completion"
---
# Squadron CLI Reference
## squadron completion

Generate completion script

### Synopsis

To load completions:

Bash:

  $ source <(squadron completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ squadron completion bash > /etc/bash_completion.d/squadron
  # macOS:
  $ squadron completion bash > /usr/local/etc/bash_completion.d/squadron

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ squadron completion zsh > "${fpath[1]}/_squadron"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ squadron completion fish | source

  # To load completions for each session, execute once:
  $ squadron completion fish > ~/.config/fish/completions/squadron.fish

PowerShell:

  PS> squadron completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> squadron completion powershell > squadron.ps1
  # and source this file from your PowerShell profile.


```
squadron completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

