---
title: "squadron schema"
---
# Squadron CLI Reference
## squadron schema

generate squadron json schema

```
squadron schema [SQUADRON] [flags]
```

### Examples

```
  squadron schema
```

### Options

```
      --base-schema string   Base schema to use (default "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json")
  -h, --help                 help for schema
      --output string        write the output to the given path
      --raw                  print raw output without highlighting
      --tags strings         list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)
```

### Options inherited from parent commands

```
  -d, --debug          show all output
  -f, --file strings   specify alternative squadron files (default [squadron.yaml])
```

### SEE ALSO

* [squadron](/reference/cli/squadron.html)	 - Docker compose for kubernetes

