---
title: "squadron build"
---
# Squadron CLI Reference
## squadron build

build or rebuild squadron units

```
squadron build [SQUADRON.UNIT...] [flags]
```

### Examples

```
squadron build storefinder frontend backend
```

### Options

```
      --build-args stringArray   additional docker buildx build args
  -h, --help                     help for build
      --parallel int             run command in parallel (default 1)
      --push                     pushes built squadron units to the registry
      --push-args stringArray    additional docker push args
      --tags strings             list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

