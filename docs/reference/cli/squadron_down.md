---
title: "squadron down"
---
# Squadron CLI Reference
## squadron down

uninstalls the squadron or given units

```
squadron down [SQUADRON] [UNIT...] [flags]
```

### Examples

```
  squadron down storefinder frontend backend --namespace demo
```

### Options

```
  -h, --help               help for down
  -n, --namespace string   set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}}) (default "default")
      --parallel int       run command in parallel (default 1)
      --tags strings       list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

