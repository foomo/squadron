---
title: "squadron bake"
---
# Squadron CLI Reference
## squadron bake

bake or rebake squadron units

```
squadron bake [SQUADRON.UNIT...] [flags]
```

### Examples

```
squadron bake storefinder frontend backend
```

### Options

```
      --bake-args stringArray   additional docker bake args
  -h, --help                    help for bake
      --output string           write the output to the given path
      --parallel int            run command in parallel (default 1)
      --push                    pushes built squadron units to the registry
      --push-args stringArray   additional docker push args
      --tags strings            list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

