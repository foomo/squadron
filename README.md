# Squadron

[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/squadron)](https://goreportcard.com/report/github.com/foomo/squadron)
[![godoc](https://godoc.org/github.com/foomo/squadron?status.svg)](https://godoc.org/github.com/foomo/squadron)
[![goreleaser](https://github.com/foomo/squadron/workflows/goreleaser/badge.svg)](https://github.com/foomo/squadron/actions)

Application for managing kubernetes microservice environments.

Use it, if a helm chart is not enough in order to organize multiple services into an effective squadron.

Another way to think of it would be `helm-compose`, because it makes k8s and helm way more approachable, not matter if it is development or production (where it just becomes another helm chart)

## Quickstart

Configure your squadron

```yaml
# squadron.yaml
version: "1.0"

squadron:
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
