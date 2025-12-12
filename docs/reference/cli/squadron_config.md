---
title: "squadron config"
---
# Squadron CLI Reference
## squadron config

generate and view the squadron config

```
squadron config [SQUADRON] [UNIT...] [flags]
```

### Examples

```
  squadron config storefinder frontend backend
```

### Options

```
  -h, --help            help for config
      --no-render       don't render the config template
      --output string   write the output to the given path
      --raw             print raw output without highlighting
      --tags strings    list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

