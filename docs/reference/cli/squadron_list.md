---
title: "squadron list"
---
# Squadron CLI Reference
## squadron list

list squadron units

```
squadron list [SQUADRON] [flags]
```

### Examples

```
  squadron list storefinder
```

### Options

```
  -h, --help            help for list
      --output string   write the output to the given path
      --tags strings    list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
      --with-bakes      include bakes
      --with-builds     include builds
      --with-charts     include charts
      --with-priority   include priority
      --with-tags       include tags
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

