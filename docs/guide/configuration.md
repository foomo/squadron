---
title: Configuration
---

# Configuration

Squadron is configured with one or more YAML files (default: `squadron.yaml`), schema version `2.3`. Values support Go
templating with `<% %>` delimiters — see
[Core Concepts](/guide/concepts#templating).

You can pass several files with repeated `-f` flags; they are merged in order, with later files overriding earlier ones.

## Top-level structure

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json
version: '2.3'        # required, schema version

vars: { }             # template variables, available as <% .Vars.* %>
global: { }           # Helm global values shared across units

builds: { }           # reusable top-level builds, referenced by units
bake: ''             # path/override for the generated buildx bake file

squadron: # the squadrons → units tree
  <squadron>:
    <unit>:
      ...
```

| Field      | Type   | Description                                        |
|------------|--------|----------------------------------------------------|
| `version`  | string | Schema version. Required. Currently `2.3`.         |
| `vars`     | map    | Template variables, referenced as `<% .Vars.x %>`. |
| `global`   | map    | Helm global values merged into every unit.         |
| `builds`   | map    | Shared build definitions units can reference.      |
| `bake`     | string | Override for the generated `buildx bake` file.     |
| `squadron` | map    | The squadrons, each containing units.              |

## Unit

```yaml
squadron:
  storefinder:
    backend:
      name: my-backend          # optional release name (default: unit name)
      namespace: my-namespace   # optional target namespace
      priority: 0               # higher is installed first
      tags: [ web, api ]          # filter labels for --tags (matched with OR)
      extends: ./defaults.yaml  # merge values from an external file
      kustomize: ./kustomize    # path to Kustomize resources
      helmArgs: [ --skip-crds ]   # extra helm flags for this unit
      chart: ...                # Helm chart (see below)
      builds: ...               # docker build targets (see below)
      bakes: ...                # docker buildx bake targets (see below)
      values: { }                # Helm values for the chart
```

| Field       | Type       | Description                                             |
|-------------|------------|---------------------------------------------------------|
| `chart`     | string/map | Chart to deploy. Inline path string, or a chart map.    |
| `values`    | map        | Helm values passed to the chart.                        |
| `builds`    | map        | Named `docker build` targets.                           |
| `bakes`     | map        | Named `docker buildx bake` targets.                     |
| `tags`      | list       | Labels used by `--tags` filtering (ORed; `-` excludes). |
| `priority`  | int        | Install ordering; higher comes first.                   |
| `name`      | string     | Override the Helm release name.                         |
| `namespace` | string     | Override the target namespace.                          |
| `extends`   | string     | File whose values are merged into this unit.            |
| `kustomize` | string     | Path to Kustomize resources.                            |
| `helmArgs`  | list       | Extra helm flags for this unit (see below).             |

### helmArgs

`helmArgs` is a raw pass-through of additional flags to the underlying helm invocation, scoped to a single unit. Use it
when a flag should apply to one unit rather than the whole squadron — which is what the command line `--` pass-through
(`squadron up -- --skip-crds`) would do.

```yaml
squadron:
  storefinder:
    backend:
      helmArgs:
        - --skip-crds
        - --no-hooks
        - --timeout=<% env "HELM_TIMEOUT" | default "5m" %>
```

Notes:

- Applies to `squadron up`, `squadron diff` and `squadron template`. It is **not**
  passed to `down`, `rollback` or `status`, which reject most of these flags.
- Flags are not validated — passing one the target helm subcommand does not accept makes helm fail.
- Args from the command line `--` pass-through are appended *after* these, so they win: a unit setting `--skip-crds` can
  be overridden with
  `squadron up -- --skip-crds=false`.
- Note that `helm upgrade` never installs CRDs, so `--skip-crds` only has an effect on the initial install and on
  `squadron template`.
- Values are rendered as Go templates like the rest of the config.

## Chart

`chart` can be an inline path string:

```yaml
chart: <% env "PROJECT_ROOT" %>/path/to/chart
```

…or a structured reference to a packaged chart:

```yaml
chart:
  name: mychart
  version: 0.1.0
  repository: https://helm.mycompany.com/repository
  alias: my-alias       # optional
  schema: ./values.schema.json   # optional values schema
```

## Builds

Each entry under `builds` is a `docker build` target. Common fields:

```yaml
builds:
  default:
    image: docker.mycompany.com/mycompany/backend
    tag: latest
    context: <% env "PROJECT_ROOT" %>/app
    dockerfile: <% env "PROJECT_ROOT" %>/Dockerfile
    buildArg:
      - "FOO=foo"
    platform:
      - linux/amd64
    target: production
```

Builds expose the full `docker build` / Buildx option surface — including
`platform`, `target`, `secret`, `ssh`, `cacheFrom`/`cacheTo`, `provenance`,
`sbom`, `output`, and `push`. See [`squadron build`](/reference/cli/squadron_build).

## Bakes

Each entry under `bakes` is a `docker buildx bake` target, mapping to the HCL bake schema:

```yaml
bakes:
  service:
    tags:
      - docker.mycompany.com/mycompany/frontend:latest
    dockerfile: Dockerfile
    context: .
    args:
      FOO: foo
    platforms:
      - linux/amd64
      - linux/arm64
    inherits:
      - base
```

Bake targets support `inherits`, `contexts`, `cacheFrom`/`cacheTo`, `secret`,
`ssh`, `outputs`, and the other Buildx bake attributes. See
[`squadron bake`](/reference/cli/squadron_bake).

## JSON schema

The full machine-readable schema lives at
[`squadron.schema.json`](https://github.com/foomo/squadron/blob/main/squadron.schema.json)
and can be regenerated with [`squadron schema`](/reference/cli/squadron_schema). Add the language-server hint shown
above to get autocompletion and validation in your editor.
