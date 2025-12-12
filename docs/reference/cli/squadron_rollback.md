---
title: "squadron rollback"
---
# Squadron CLI Reference
## squadron rollback

rolls back the squadron or given units

```
squadron rollback [SQUADRON] [UNIT...] [flags]
```

### Examples

```
  squadron rollback storefinder frontend backend --namespace demo
```

### Options

```
  -h, --help               help for rollback
  -n, --namespace string   set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}}) (default "default")
      --parallel int       run command in parallel (default 1)
  -r, --revision string    specifies the revision to roll back to
      --tags strings       list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

