[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/squadron?style=flat-square)](https://goreportcard.com/report/github.com/foomo/squadron)
[![GoDoc](https://img.shields.io/badge/GoDoc-✓-informational.svg?style=flat-square&logo=go)](https://godoc.org/github.com/foomo/squadron)
[![Coverage](https://img.shields.io/codecov/c/github/foomo/squadron?style=flat-square&logo=github)](https://app.codecov.io/gh/foomo/squadron)
[![GitHub Downloads](https://img.shields.io/github/downloads/foomo/squadron/total.svg?style=flat-square&logo=github)](https://github.com/foomo/squadron/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/foomo/squadron.svg?style=flat-square&logo=docker)](https://hub.docker.com/r/foomo/squadron)
[![GitHub Stars](https://img.shields.io/github/stars/foomo/squadron.svg?style=flat-square&logo=github)](https://github.com/foomo/squadron)

<p align="center">
  <img alt="squadron" src="docs/public/logo.png" width="400" height="400"/>
</p>

# Squadron

**Docker Compose for Kubernetes.**

Squadron is a CLI that orchestrates multiple Helm charts and Docker image builds
as one cohesive deployment, driven by a single declarative `squadron.yaml`. It
brings the familiar `docker-compose` workflow — define your services once, bring
the whole stack up with one command — to Kubernetes, where each service is a real
Helm release. In production, a squadron is just another set of Helm charts.

📖 **[Read the documentation →](https://foomo.github.io/squadron)**

## Install

```shell
go install github.com/foomo/squadron/cmd/squadron@latest
```

See the [installation guide](https://foomo.github.io/squadron/guide/installation)
for release binaries and the Docker image.

## Example

```yaml
# squadron.yaml
version: '2.3'
squadron:
  storefinder:
    backend:
      chart: <% env "PROJECT_ROOT" %>/charts/backend
      builds:
        default:
          image: docker.mycompany.com/storefinder/backend
          tag: latest
          context: ./app
      values:
        image:
          repository: <% .Squadron.storefinder.backend.builds.default.image %>
          tag: <% .Squadron.storefinder.backend.builds.default.tag | quote %>
```

```shell
squadron build   # build the images
squadron up      # install / upgrade the releases
squadron status  # check release status
squadron down    # tear it down
```

See the [Quick Start](https://foomo.github.io/squadron/guide/quickstart) and the
[configuration reference](https://foomo.github.io/squadron/guide/configuration)
for the full picture.

## How to Contribute

Contributions are welcome! Please read the [contributing guide](docs/CONTRIBUTING.md).

![Contributors](https://contributors-table.vercel.app/image?repo=foomo/squadron&width=50&columns=15)

## License

Distributed under MIT License, please see the [license](LICENSE) file for more details.

_Made with ♥ [foomo](https://www.foomo.org) by [bestbytes](https://www.bestbytes.com)_
