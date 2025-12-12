---
title: "squadron push"
---
# Squadron CLI Reference
## squadron push

pushes the squadron or given units

```
squadron push [SQUADRON] [UNIT...] [flags]
```

### Examples

```
  squadron push storefinder frontend backend --namespace demo --build
```

### Options

```
      --bake                     bakes or rebakes units
      --bake-args stringArray    additional docker buildx bake args
      --build                    builds or rebuilds units
      --build-args stringArray   additional docker buildx build args
  -h, --help                     help for push
  -n, --namespace string         set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}}) (default "default")
      --parallel int             run command in parallel (default 1)
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

