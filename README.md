# Squadron

[![Build Status](https://github.com/foomo/squadron/actions/workflows/test.yml/badge.svg?branch=main&event=push)](https://github.com/foomo/squadron/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/squadron)](https://goreportcard.com/report/github.com/foomo/squadron)
[![Coverage Status](https://coveralls.io/repos/github/foomo/squadron/badge.svg?branch=main&)](https://coveralls.io/github/foomo/squadron?branch=main)
[![GoDoc](https://godoc.org/github.com/foomo/squadron?status.svg)](https://godoc.org/github.com/foomo/squadron)

Application for managing kubernetes microservice environments.

Use it, if a helm chart is not enough in order to organize multiple services into an effective squadron.

Another way to think of it would be `helm-compose`, because it makes k8s and helm way more approachable, not matter if it is development or production (where it just becomes another helm chart)

## Quickstart

Configure your squadron

```yaml
# squadron.yaml
version: '2.0'

squadron:
  site:
    frontend:
      chart:
        name: mychart
        version: 0.1.0
        repository: http://helm.mycompany.com/repository
      builds:
        service:
          tag: latest
          dockerfile: Dockerfile
          image: docker.mycompany.com/mycomapny/frontend
          args:
            - "foo=foo"
            - "bar=bar"
      values:
        image: docker.mycompany.com/mycomapny/frontend:latest
    backend:
      chart: <% env "PROJECT_ROOT" %>/path/to/chart
      kustomize: <% env "PROJECT_ROOT" %>/path/to/kustomize
      builds:
        service:
          tag: latest
          dockerfile: Dockerfile
          image: docker.mycompany.com/mycomapny/backend
          args:
            - "foo=foo"
            - "bar=bar"
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
