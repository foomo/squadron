---
title: Introduction
---

# Introduction

**Squadron** is _Docker Compose for Kubernetes_: a single CLI that orchestrates
multiple Helm charts and Docker image builds as one cohesive deployment, driven
by one declarative file — `squadron.yaml`.

## Why Squadron?

`docker-compose` made multi-service local development easy: one file lists every
service, and one command brings the whole stack up. Kubernetes and Helm are the
production standard, but the developer experience is different — every service
becomes its own Helm chart and release, each with its own values, lifecycle, and
`helm` invocation. Coordinating a dozen of those by hand is tedious and
error-prone.

Squadron closes that gap. You describe your services once, grouped into
**squadrons** of **units**, and Squadron handles the rest:

- **One file, many charts** — group related releases and manage them together.
- **Build and deploy in one workflow** — define Docker builds (or `buildx bake`
  targets) alongside the chart that consumes them.
- **Templated configuration** — share and compute values across units with Go
  templates and helpers like `env`, `file`, and `git`.
- **The full Helm lifecycle** — `up`, `down`, `diff`, `status`, `rollback`, and
  `template`, scoped to the whole squadron, individual units, or by tag.

In production a squadron is just another set of Helm releases, so you keep all
the tooling and guarantees you already rely on.

## Coming from docker-compose

| docker-compose            | Squadron                                              |
| ------------------------- | ----------------------------------------------------- |
| `docker-compose.yml`      | `squadron.yaml`                                        |
| a `service`               | a **unit** (a Helm release + optional image build)    |
| the whole compose project | a **squadron** (a named group of units)               |
| `build:`                  | `builds:` / `bakes:` (Docker build / `buildx bake`)   |
| `docker compose up`       | `squadron up`                                          |
| `docker compose down`     | `squadron down`                                        |
| environment interpolation | Go templates with `env`, `file`, `git`, and more      |

The big difference: each unit is backed by a **Helm chart**, so you get real
Kubernetes deployments instead of local containers — without writing a `helm`
command per service.

## Production-ready

Squadron is battle-tested and used in production at BestBytes.

## Next steps

- [Installation](/guide/installation) — install the CLI.
- [Quick Start](/guide/quickstart) — deploy the `helloworld` example.
- [Core Concepts](/guide/concepts) — squadrons, units, builds, bakes, templating.
- [Configuration](/guide/configuration) — the `squadron.yaml` reference.
