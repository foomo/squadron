---
title: Core Concepts
---

# Core Concepts

## Squadrons and units

A `squadron.yaml` describes one or more **squadrons**, and each squadron
contains one or more **units**:

```yaml
squadron:
  <squadron>:      # a named group, e.g. "storefinder"
    <unit>:        # a deployable service, e.g. "backend"
      chart: ...
      builds: ...
      values: ...
```

A **unit** is the core building block. It maps to a single Helm release and,
optionally, the Docker image(s) that release runs. A **squadron** is just a
named group of units you want to manage together.

## What a unit contains

| Field       | Purpose                                                                 |
| ----------- | ----------------------------------------------------------------------- |
| `chart`     | The Helm chart to deploy (inline path, or `name`/`version`/`repository`). |
| `values`    | Helm values passed to the chart.                                        |
| `builds`    | Docker images to build with `docker build`.                             |
| `bakes`     | Docker images to build with `docker buildx bake`.                       |
| `tags`      | Labels used to filter which units a command acts on.                    |
| `priority`  | Install ordering — higher priority is applied first.                    |
| `name`      | Override the Helm release name (defaults to the unit name).             |
| `namespace` | Override the target namespace.                                          |
| `extends`   | Merge values from an external file.                                     |
| `kustomize` | Path to Kustomize resources to include.                                 |

See [Configuration](/guide/configuration) for the full field reference.

## Builds vs. bakes

Both produce the container images your charts run; they differ in the engine:

- **`builds`** invoke `docker build` directly — one image per build entry, with
  the familiar build options (`context`, `dockerfile`, `buildArg`, `platform`,
  `target`, `secret`, `cacheFrom`/`cacheTo`, …).
- **`bakes`** generate a `docker buildx bake` HCL file and build through Buildx,
  which is better for building many targets together, advanced caching, and
  multi-platform output. Bake targets can `inherit` from one another.

Use `builds` for straightforward single-image units; reach for `bakes` when you
want Buildx orchestration across multiple targets.

## The lifecycle

When you run a command, Squadron processes your configuration in stages:

1. **Merge** — multiple `-f` config files are conflated into one (later files
   override earlier ones).
2. **Filter** — narrow to the requested squadron, units, or `--tags`.
3. **Render** — execute Go templates in the configuration values.
4. **Build / Bake** — build and (optionally) push images.
5. **Deploy** — run the Helm operation (`up`, `diff`, `down`, `rollback`, …).

The build and deploy stages run concurrently across units where possible, and
`priority` controls install ordering.

## Templating

Configuration values are rendered as Go templates **before** they reach Helm.
Squadron uses `<% %>` delimiters (so they don't clash with Helm's `{{ }}`) and
includes the full [Sprig](https://masterminds.github.io/sprig/) function set
plus Squadron-specific helpers:

| Helper            | Description                                              |
| ----------------- | -------------------------------------------------------- |
| `env "NAME"`      | Read an environment variable (errors if missing).        |
| `envDefault`      | Read an environment variable with a fallback.            |
| `file`            | Read and render a file from disk.                        |
| `git`             | Read git metadata (e.g. commit, branch).                 |
| `op` / `opDoc`    | Fetch secrets / documents from 1Password.                |
| `kubeseal`        | Encrypt a value with Sealed Secrets.                     |
| `quote`/`quoteAll`| Quote a value / all values in a list.                    |
| `toYaml`/`fromYaml`, `toJson`/`fromJson`, `toToml`/`fromToml` | Convert between formats. |

Within templates you can also reference the rendered configuration itself —
for example `.Squadron.<squadron>.<unit>.builds.<name>.tag` — to keep values in
sync with builds, as the [Quick Start](/guide/quickstart) shows.

::: warning Deprecated helpers
`indent`, `base64`, and `defaultIndex` still work but are deprecated; prefer the
Sprig equivalents.
:::
