---
title: Installation
---

# Installation

Squadron is a single Go binary. Pick whichever method fits your workflow.

## Prerequisites

Squadron shells out to the standard toolchain, so make sure these are available
on your `PATH`:

- [`helm`](https://helm.sh) — chart install / upgrade / diff / rollback
- [`docker`](https://docs.docker.com/get-docker/) (with `buildx`) — image builds and bakes
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/) — access to your cluster

## Install with `go install`

```shell
go install github.com/foomo/squadron/cmd/squadron@latest
```

This installs the `squadron` binary into `$GOPATH/bin`.

## Download a release binary

Pre-built binaries for Linux, macOS, and Windows are attached to every
[GitHub release](https://github.com/foomo/squadron/releases). Download the
archive for your platform, extract it, and move `squadron` onto your `PATH`.

## Docker image

```shell
docker run --rm foomo/squadron version
```

Images are published to [Docker Hub](https://hub.docker.com/r/foomo/squadron).

## From source

```shell
git clone https://github.com/foomo/squadron.git
cd squadron
make install   # builds and installs to $GOPATH/bin
```

This repository uses [mise](https://mise.jdx.dev) to pin tool versions (Go,
golangci-lint, lefthook, bun, biome); run `mise install` once to match them.

## Verify

```shell
squadron version
```

## Shell completion

Squadron can generate completion scripts for bash, zsh, fish, and PowerShell:

```shell
squadron completion zsh > "${fpath[1]}/_squadron"
```

See [`squadron completion`](/reference/cli/squadron_completion) for the other shells.
