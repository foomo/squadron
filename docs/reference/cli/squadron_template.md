---
title: "squadron template"
---
# Squadron CLI Reference
## squadron template

render chart templates locally and display the output

```
squadron template [SQUADRON] [UNIT...] [flags]
```

### Examples

```
  squadron template storefinder frontend backend --namespace demo
```

### Options

```
  -h, --help               help for template
  -n, --namespace string   set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}}) (default "default")
      --output string      write the output to the given path
      --parallel int       run command in parallel (default 1)
      --raw                print raw output without highlighting
      --tags strings       list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

