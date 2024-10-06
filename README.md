# Squadron

[![Build Status](https://github.com/foomo/squadron/actions/workflows/pr.yml/badge.svg?branch=main&event=push)](https://github.com/foomo/squadron/actions/workflows/pr.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/squadron)](https://goreportcard.com/report/github.com/foomo/squadron)
[![Coverage Status](https://coveralls.io/repos/github/foomo/squadron/badge.svg?branch=main&)](https://coveralls.io/github/foomo/squadron?branch=main)
[![GoDoc](https://godoc.org/github.com/foomo/squadron?status.svg)](https://godoc.org/github.com/foomo/squadron)

Application for managing kubernetes microservice environments.

Use it, if a helm chart is not enough in order to organize multiple services into an effective squadron.

Another way to think of it would be `helm-compose`, because it makes k8s and helm way more approachable, not matter if it is development or production (where it just becomes another helm chart)

## Quickstart

Configure your squadron

```yaml
# https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json
version: '2.1'

# squadron template vars
vars: {}

# helm global vars
global: {}

# squadron definitions
squadron:
  # squadron units
  site:
    # squadron unit
    frontend:
      # helm chart definition
      chart:
        name: mychart
        version: 0.1.0
        repository: http://helm.mycompany.com/repository
      # container builds
      builds:
        service:
          tag: latest
          file: Dockerfile
          image: docker.mycompany.com/mycomapny/frontend
          build_arg:
            - "foo=foo"
            - "bar=bar"
      # helm chart values
      values:
        image: docker.mycompany.com/mycomapny/frontend:latest
    # squadron unit
    backend:
      # helm chart definition
      chart: <% env "PROJECT_ROOT" %>/path/to/chart
      # kustomize path
      kustomize: <% env "PROJECT_ROOT" %>/path/to/kustomize
      # container builds
      builds:
        service:
          tag: latest
          file: Dockerfile
          image: docker.mycompany.com/mycomapny/backend
          build_arg:
            - "foo=foo"
            - "bar=bar"
      # helm chart values
      values:
        image: docker.mycompany.com/mycomapny/backend:latest
```

Install the squadron squadron and namespace:

```text
$ squadron up --build --push --namespace default
```

Uninstall the squadron again:

```text
$ squadron down
```

## Commands

```text
# See:
$ squadron help
```

## See also

Sometimes as a sailor or a pirate you might need to get a grapple : go get [github.com/foomo/gograpple/...](https//:github.com/foomo/gograpple)

## How to Contribute

Make a pull request...

## License

Distributed under MIT License, please see license file within the code for more details.
