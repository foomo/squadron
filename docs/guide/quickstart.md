---
title: Quick Start
---

# Quick Start

This walkthrough uses the [`helloworld`
example](https://github.com/foomo/squadron/tree/main/_examples/helloworld) from
the repository. It defines a single squadron (`storefinder`) with one unit
(`backend`) that builds a small Go HTTP server and deploys it with a Helm chart.

## The configuration

`_examples/helloworld/squadron.yaml`:

```yaml
version: '2.3'

squadron:
  storefinder:
    backend:
      chart: <% env "PROJECT_ROOT" %>/../common/charts/backend
      builds:
        default:
          tag: latest
          context: <% env "PROJECT_ROOT" %>/app
          dockerfile: <% env "PROJECT_ROOT" %>/../common/docker/backend.Dockerfile
          image: helloworld/app
      values:
        image:
          tag: <% .Squadron.storefinder.backend.builds.default.tag | quote %>
          repository: <% .Squadron.storefinder.backend.builds.default.image %>
        service:
          ports:
            - 80
```

A few things to notice:

- `<% ... %>` are Go template expressions — Squadron uses `<% %>` delimiters, not
  the `{{ }}` Helm uses.
- `env "PROJECT_ROOT"` reads an environment variable, so paths resolve relative
  to the example directory.
- The `values` reference the build's own fields
  (`.Squadron.storefinder.backend.builds.default...`), so the image tag and
  repository stay in sync with the build definition automatically.

## Run it

From `_examples/helloworld` (which exports `PROJECT_ROOT`):

```shell
# 1. Inspect the fully merged & rendered config
squadron config

# 2. List the squadrons and units
squadron list

# 3. Build the Docker image for the unit
squadron build

# 4. Render the Helm templates locally (no cluster needed)
squadron template

# 5. Install / upgrade the releases on your current kube-context
squadron up

# 6. Check release status
squadron status

# 7. Tear it down
squadron down
```

::: tip
`squadron template` and `squadron config` don't touch your cluster, so they're a
safe way to explore what Squadron will do before running `up`.
:::

## Scope your commands

Every lifecycle command accepts squadron and unit selectors, so you can act on
part of the fleet:

```shell
squadron up storefinder           # everything in the storefinder squadron
squadron up storefinder backend   # just the backend unit
squadron up --tags web            # only units tagged "web"
squadron up --tags web,-legacy    # include "web", exclude "legacy"
```

## Next steps

- [Core Concepts](/guide/concepts) — how squadrons, units, builds, and bakes fit together.
- [Configuration](/guide/configuration) — the full `squadron.yaml` reference.
- [CLI Reference](/reference/cli/squadron) — every command and flag.
